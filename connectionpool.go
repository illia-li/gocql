// Copyright (c) 2012 The gocql Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gocql

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"sync"
	"time"
)

// interface to implement to receive the host information
type SetHosts interface {
	SetHosts(hosts []*HostInfo)
}

// interface to implement to receive the partitioner value
type SetPartitioner interface {
	SetPartitioner(partitioner string)
}

func setupTLSConfig(sslOpts *SslOptions) (*tls.Config, error) {
	//  Config.InsecureSkipVerify | EnableHostVerification | Result
	//  Config is nil             | true                   | verify host
	//  Config is nil             | false                  | do not verify host
	//  false                     | false                  | verify host
	//  true                      | false                  | do not verify host
	//  false                     | true                   | verify host
	//  true                      | true                   | verify host
	var tlsConfig *tls.Config
	if sslOpts.Config == nil {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: !sslOpts.EnableHostVerification,
		}
	} else {
		// use clone to avoid race.
		tlsConfig = sslOpts.Config.Clone()
	}

	if tlsConfig.InsecureSkipVerify && sslOpts.EnableHostVerification {
		tlsConfig.InsecureSkipVerify = false
	}

	// ca cert is optional
	if sslOpts.CaPath != "" {
		if tlsConfig.RootCAs == nil {
			tlsConfig.RootCAs = x509.NewCertPool()
		}

		pem, err := ioutil.ReadFile(sslOpts.CaPath)
		if err != nil {
			return nil, fmt.Errorf("connectionpool: unable to open CA certs: %v", err)
		}

		if !tlsConfig.RootCAs.AppendCertsFromPEM(pem) {
			return nil, errors.New("connectionpool: failed parsing or CA certs")
		}
	}

	if sslOpts.CertPath != "" || sslOpts.KeyPath != "" {
		mycert, err := tls.LoadX509KeyPair(sslOpts.CertPath, sslOpts.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("connectionpool: unable to load X509 key pair: %v", err)
		}
		tlsConfig.Certificates = append(tlsConfig.Certificates, mycert)
	}

	return tlsConfig, nil
}

type policyConnPool struct {
	session *Session

	port     int
	numConns int
	keyspace string

	mu            sync.RWMutex
	hostConnPools map[string]*hostConnPool

	endpoints []string
}

func connConfig(cfg *ClusterConfig) (*ConnConfig, error) {
	var (
		err       error
		tlsConfig *tls.Config
	)

	// TODO(zariel): move tls config setup into session init.
	if cfg.SslOpts != nil {
		tlsConfig, err = setupTLSConfig(cfg.SslOpts)
		if err != nil {
			return nil, err
		}
	}

	return &ConnConfig{
		ProtoVersion:    cfg.ProtoVersion,
		CQLVersion:      cfg.CQLVersion,
		Timeout:         cfg.Timeout,
		ConnectTimeout:  cfg.ConnectTimeout,
		Dialer:          cfg.Dialer,
		Compressor:      cfg.Compressor,
		Authenticator:   cfg.Authenticator,
		AuthProvider:    cfg.AuthProvider,
		Keepalive:       cfg.SocketKeepalive,
		tlsConfig:       tlsConfig,
		disableCoalesce: tlsConfig != nil, // write coalescing doesn't work with framing on top of TCP like in TLS.
	}, nil
}

func newPolicyConnPool(session *Session) *policyConnPool {
	// create the pool
	pool := &policyConnPool{
		session:       session,
		port:          session.cfg.Port,
		numConns:      session.cfg.NumConns,
		keyspace:      session.cfg.Keyspace,
		hostConnPools: map[string]*hostConnPool{},
	}

	pool.endpoints = make([]string, len(session.cfg.Hosts))
	copy(pool.endpoints, session.cfg.Hosts)

	return pool
}

func (p *policyConnPool) SetHosts(hosts []*HostInfo) {
	p.mu.Lock()
	defer p.mu.Unlock()

	toRemove := make(map[string]struct{})
	for addr := range p.hostConnPools {
		toRemove[addr] = struct{}{}
	}

	pools := make(chan *hostConnPool)
	createCount := 0
	for _, host := range hosts {
		if !host.IsUp() {
			// don't create a connection pool for a down host
			continue
		}
		ip := host.ConnectAddress().String()
		if _, exists := p.hostConnPools[ip]; exists {
			// still have this host, so don't remove it
			delete(toRemove, ip)
			continue
		}

		createCount++
		go func(host *HostInfo) {
			// create a connection pool for the host
			pools <- newHostConnPool(
				p.session,
				host,
				p.port,
				p.numConns,
				p.keyspace,
			)
		}(host)
	}

	// add created pools
	for createCount > 0 {
		pool := <-pools
		createCount--
		if pool.Size() > 0 {
			// add pool only if there a connections available
			p.hostConnPools[string(pool.host.ConnectAddress())] = pool
		}
	}

	for addr := range toRemove {
		pool := p.hostConnPools[addr]
		delete(p.hostConnPools, addr)
		go pool.Close()
	}
}

func (p *policyConnPool) Size() int {
	p.mu.RLock()
	count := 0
	for _, pool := range p.hostConnPools {
		count += pool.Size()
	}
	p.mu.RUnlock()

	return count
}

func (p *policyConnPool) getPool(host *HostInfo) (pool *hostConnPool, ok bool) {
	ip := host.ConnectAddress().String()
	p.mu.RLock()
	pool, ok = p.hostConnPools[ip]
	p.mu.RUnlock()
	return
}

func (p *policyConnPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	// close the pools
	for addr, pool := range p.hostConnPools {
		delete(p.hostConnPools, addr)
		pool.Close()
	}
}

func (p *policyConnPool) addHost(host *HostInfo) {
	ip := host.ConnectAddress().String()
	p.mu.Lock()
	pool, ok := p.hostConnPools[ip]
	if !ok {
		pool = newHostConnPool(
			p.session,
			host,
			host.Port(), // TODO: if port == 0 use pool.port?
			p.numConns,
			p.keyspace,
		)

		p.hostConnPools[ip] = pool
	}
	p.mu.Unlock()

	pool.fill()
}

func (p *policyConnPool) removeHost(ip net.IP) {
	k := ip.String()
	p.mu.Lock()
	pool, ok := p.hostConnPools[k]
	if !ok {
		p.mu.Unlock()
		return
	}

	delete(p.hostConnPools, k)
	p.mu.Unlock()

	go pool.Close()
}

func (p *policyConnPool) hostDown(ip net.IP) {
	// TODO(zariel): mark host as down so we can try to connect to it later, for
	// now just treat it has removed.
	p.removeHost(ip)
}

// hostConnPool is a connection pool for a single host.
// Connection selection is based on a provided ConnSelectionPolicy
type hostConnPool struct {
	session  *Session
	host     *HostInfo
	port     int
	addr     string
	size     int
	keyspace string
	// protection for connPicker, closed, filling
	mu         sync.RWMutex
	connPicker ConnPicker
	closed     bool
	filling    bool
}

func (h *hostConnPool) String() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	size, _ := h.connPicker.Size()
	return fmt.Sprintf("[filling=%v closed=%v conns=%v size=%v host=%v]",
		h.filling, h.closed, size, h.size, h.host)
}

func newHostConnPool(session *Session, host *HostInfo, port, size int,
	keyspace string) *hostConnPool {

	pool := &hostConnPool{
		session:    session,
		host:       host,
		port:       port,
		addr:       (&net.TCPAddr{IP: host.ConnectAddress(), Port: host.Port()}).String(),
		size:       size,
		keyspace:   keyspace,
		connPicker: nopConnPicker{},
		filling:    false,
		closed:     false,
	}

	// the pool is not filled or connected
	return pool
}

// Pick a connection from this connection pool for the given query.
func (pool *hostConnPool) Pick(token token) *Conn {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	if pool.closed {
		return nil
	}

	size, missing := pool.connPicker.Size()
	if missing > 0 {
		// try to fill the pool
		go pool.fill()

		if size == 0 {
			return nil
		}
	}

	return pool.connPicker.Pick(token)
}

//Size returns the number of connections currently active in the pool
func (pool *hostConnPool) Size() int {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	size, _ := pool.connPicker.Size()
	return size
}

//Close the connection pool
func (pool *hostConnPool) Close() {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	if !pool.closed {
		pool.connPicker.Close()
	}
	pool.closed = true
}

// Fill the connection pool
func (pool *hostConnPool) fill() {
	pool.mu.RLock()
	// avoid filling a closed pool, or concurrent filling
	if pool.closed || pool.filling {
		pool.mu.RUnlock()
		return
	}

	// determine the filling work to be done
	startCount, fillCount := pool.connPicker.Size()

	// avoid filling a full (or overfull) pool
	if fillCount <= 0 {
		pool.mu.RUnlock()
		return
	}

	// switch from read to write lock
	pool.mu.RUnlock()
	pool.mu.Lock()

	startCount, fillCount = pool.connPicker.Size()
	if pool.closed || pool.filling || fillCount <= 0 {
		// looks like another goroutine already beat this
		// goroutine to the filling
		pool.mu.Unlock()
		return
	}

	// ok fill the pool
	pool.filling = true

	// allow others to access the pool while filling
	pool.mu.Unlock()
	// only this goroutine should make calls to fill/empty the pool at this
	// point until after this routine or its subordinates calls
	// fillingStopped

	// fill only the first connection synchronously
	if startCount == 0 {
		err := pool.connect()
		pool.logConnectErr(err)

		if err != nil {
			// probably unreachable host
			pool.fillingStopped(true)

			// this is call with the connection pool mutex held, this call will
			// then recursively try to lock it again. FIXME
			if pool.session.cfg.ConvictionPolicy.AddFailure(err, pool.host) {
				go pool.session.handleNodeDown(pool.host.ConnectAddress(), pool.port)
			}
			return
		}
		// notify the session that this node is connected
		go pool.session.handleNodeUp(pool.host.ConnectAddress(), pool.port)

		// filled one, let's reload it to see if it has changed
		pool.mu.RLock()
		_, fillCount = pool.connPicker.Size()
		pool.mu.RUnlock()
	}

	// fill the rest of the pool asynchronously
	go func() {
		err := pool.connectMany(fillCount)

		// mark the end of filling
		pool.fillingStopped(err != nil)

		if err == nil && startCount > 0 {
			// notify the session that this node is connected again
			go pool.session.handleNodeUp(pool.host.ConnectAddress(), pool.port)
		}
	}()
}

func (pool *hostConnPool) logConnectErr(err error) {
	if opErr, ok := err.(*net.OpError); ok && (opErr.Op == "dial" || opErr.Op == "read") {
		// connection refused
		// these are typical during a node outage so avoid log spam.
		if gocqlDebug {
			Logger.Printf("unable to dial %q: %v\n", pool.host.ConnectAddress(), err)
		}
	} else if err != nil {
		// unexpected error
		Logger.Printf("error: failed to connect to %s due to error: %v", pool.addr, err)
	}
}

// transition back to a not-filling state.
func (pool *hostConnPool) fillingStopped(hadError bool) {
	if hadError {
		// wait for some time to avoid back-to-back filling
		// this provides some time between failed attempts
		// to fill the pool for the host to recover
		time.Sleep(time.Duration(rand.Int31n(100)+31) * time.Millisecond)
	}

	pool.mu.Lock()
	pool.filling = false
	pool.mu.Unlock()
}

// connectMany creates new connections concurrent.
func (pool *hostConnPool) connectMany(count int) error {
	if count == 0 {
		return nil
	}
	var (
		wg         sync.WaitGroup
		mu         sync.Mutex
		connectErr error
	)
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			err := pool.connect()
			pool.logConnectErr(err)
			if err != nil {
				mu.Lock()
				connectErr = err
				mu.Unlock()
			}
		}()
	}
	// wait for all connections are done
	wg.Wait()

	return connectErr
}

// create a new connection to the host and add it to the pool
func (pool *hostConnPool) connect() (err error) {
	pool.mu.Lock()
	dialer := pool.connPicker.GetCustomDialer()
	pool.mu.Unlock()
	if dialer == nil {
		dialer = pool.session.cfg.Dialer
	}

	// TODO: provide a more robust connection retry mechanism, we should also
	// be able to detect hosts that come up by trying to connect to downed ones.
	// try to connect
	var conn *Conn
	reconnectionPolicy := pool.session.cfg.ReconnectionPolicy
	for i := 0; i < reconnectionPolicy.GetMaxRetries(); i++ {
		conn, err = pool.session.connectWithDialer(pool.session.ctx, pool.host, pool, dialer)
		if err == nil {
			break
		}
		if opErr, isOpErr := err.(*net.OpError); isOpErr {
			// if the error is not a temporary error (ex: network unreachable) don't
			//  retry
			if !opErr.Temporary() {
				break
			}
		}
		if gocqlDebug {
			Logger.Printf("connection failed %q: %v, reconnecting with %T\n",
				pool.host.ConnectAddress(), err, reconnectionPolicy)
		}
		time.Sleep(reconnectionPolicy.GetInterval(i))
	}

	if err != nil {
		return err
	}

	if pool.keyspace != "" {
		// set the keyspace
		if err = conn.UseKeyspace(pool.keyspace); err != nil {
			conn.Close()
			return err
		}
	}

	// add the Conn to the pool
	pool.mu.Lock()
	defer pool.mu.Unlock()

	if pool.closed {
		conn.Close()
		return nil
	}

	// lazily initialize the connPicker when we know the required type
	pool.initConnPicker(conn)
	pool.connPicker.Put(conn)

	return nil
}

func (pool *hostConnPool) initConnPicker(conn *Conn) {
	if _, ok := pool.connPicker.(nopConnPicker); !ok {
		return
	}

	if isScyllaConn(conn) {
		pool.connPicker = newScyllaConnPicker(conn)
		return
	}

	pool.connPicker = newDefaultConnPicker(pool.size)
}

// handle any error from a Conn
func (pool *hostConnPool) HandleError(conn *Conn, err error, closed bool) {
	if !closed {
		// still an open connection, so continue using it
		return
	}

	// TODO: track the number of errors per host and detect when a host is dead,
	// then also have something which can detect when a host comes back.
	pool.mu.Lock()
	defer pool.mu.Unlock()

	if pool.closed {
		// pool closed
		return
	}

	pool.connPicker.Remove(conn)
}
