/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
/*
 * Content before git sha 34fdeebefcbf183ed7f916f931aa0586fdaa1b40
 * Copyright (c) 2016, The Gocql authors,
 * provided under the BSD-3-Clause License.
 * See the NOTICE file distributed with this work for additional information.
 */

package gocql

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type ExecutableQuery interface {
	borrowForExecution()    // Used to ensure that the query stays alive for lifetime of a particular execution goroutine.
	releaseAfterExecution() // Used when a goroutine finishes its execution attempts, either with ok result or an error.
	execute(ctx context.Context, conn *Conn) *Iter
	attempt(keyspace string, end, start time.Time, iter *Iter, host *HostInfo)
	retryPolicy() RetryPolicy
	speculativeExecutionPolicy() SpeculativeExecutionPolicy
	GetRoutingKey() ([]byte, error)
	Keyspace() string
	Table() string
	IsIdempotent() bool
	IsLWT() bool
	GetCustomPartitioner() Partitioner
	GetHostID() string

	withContext(context.Context) ExecutableQuery

	RetryableQuery

	GetSession() *Session
}

type queryExecutor struct {
	pool   *policyConnPool
	policy HostSelectionPolicy
}

func (q *queryExecutor) attemptQuery(ctx context.Context, qry ExecutableQuery, conn *Conn) *Iter {
	start := time.Now()
	iter := qry.execute(ctx, conn)
	end := time.Now()

	qry.attempt(q.pool.keyspace, end, start, iter, conn.host)

	return iter
}

func (q *queryExecutor) speculate(ctx context.Context, qry ExecutableQuery, sp SpeculativeExecutionPolicy,
	hostIter NextHost, results chan *Iter) *Iter {
	ticker := time.NewTicker(sp.Delay())
	defer ticker.Stop()

	for i := 0; i < sp.Attempts(); i++ {
		select {
		case <-ticker.C:
			qry.borrowForExecution() // ensure liveness in case of executing Query to prevent races with Query.Release().
			go q.run(ctx, qry, hostIter, results)
		case <-ctx.Done():
			return &Iter{err: ctx.Err()}
		case iter := <-results:
			return iter
		}
	}

	return nil
}

func (q *queryExecutor) executeQuery(qry ExecutableQuery) (*Iter, error) {
	var hostIter NextHost

	// check if the hostID is specified for the query,
	// if true  - the query execute at the specified host.
	// if false - the query execute at the host picked by HostSelectionPolicy
	if hostID := qry.GetHostID(); hostID != "" {
		pool, ok := q.pool.getPoolByHostID(hostID)
		if !ok {
			// if the specified host ID have no connection pool we return error
			return nil, fmt.Errorf("query is targeting unknown host id %s: %w", hostID, ErrNoPool)
		} else if pool.Size() == 0 {
			// if the pool have no connection we return error
			return nil, fmt.Errorf("query is targeting host id %s that driver is not connected to: %w", hostID, ErrNoConnectionsInPool)
		}
		hostIter = newSingleHost(pool.host, 5, 200*time.Millisecond).selectHost
	} else {
		hostIter = q.policy.Pick(qry)
	}

	// check if the query is not marked as idempotent, if
	// it is, we force the policy to NonSpeculative
	sp := qry.speculativeExecutionPolicy()
	if qry.GetHostID() != "" || !qry.IsIdempotent() || sp.Attempts() == 0 {
		return q.do(qry.Context(), qry, hostIter), nil
	}

	// When speculative execution is enabled, we could be accessing the host iterator from multiple goroutines below.
	// To ensure we don't call it concurrently, we wrap the returned NextHost function here to synchronize access to it.
	var mu sync.Mutex
	origHostIter := hostIter
	hostIter = func() SelectedHost {
		mu.Lock()
		defer mu.Unlock()
		return origHostIter()
	}

	ctx, cancel := context.WithCancel(qry.Context())
	defer cancel()

	results := make(chan *Iter, 1)

	// Launch the main execution
	qry.borrowForExecution() // ensure liveness in case of executing Query to prevent races with Query.Release().
	go q.run(ctx, qry, hostIter, results)

	// The speculative executions are launched _in addition_ to the main
	// execution, on a timer. So Speculation{2} would make 3 executions running
	// in total.
	if iter := q.speculate(ctx, qry, sp, hostIter, results); iter != nil {
		return iter, nil
	}

	select {
	case iter := <-results:
		return iter, nil
	case <-ctx.Done():
		return &Iter{err: ctx.Err()}, nil
	}
}

func (q *queryExecutor) do(ctx context.Context, qry ExecutableQuery, hostIter NextHost) *Iter {
	rt := qry.retryPolicy()
	if rt == nil {
		rt = &SimpleRetryPolicy{3}
	}

	lwtRT, isRTSupportsLWT := rt.(LWTRetryPolicy)

	var getShouldRetry func(qry RetryableQuery) bool
	var getRetryType func(error) RetryType

	if isRTSupportsLWT && qry.IsLWT() {
		getShouldRetry = lwtRT.AttemptLWT
		getRetryType = lwtRT.GetRetryTypeLWT
	} else {
		getShouldRetry = rt.Attempt
		getRetryType = rt.GetRetryType
	}

	var potentiallyExecuted bool

	execute := func(qry ExecutableQuery, selectedHost SelectedHost) (iter *Iter, retry RetryType) {
		host := selectedHost.Info()
		if host == nil || !host.IsUp() {
			return &Iter{
				err: &QueryError{
					err:                 ErrHostDown,
					potentiallyExecuted: potentiallyExecuted,
				},
			}, RetryNextHost
		}
		pool, ok := q.pool.getPool(host)
		if !ok {
			return &Iter{
				err: &QueryError{
					err:                 ErrNoPool,
					potentiallyExecuted: potentiallyExecuted,
				},
			}, RetryNextHost
		}
		conn := pool.Pick(selectedHost.Token(), qry)
		if conn == nil {
			return &Iter{
				err: &QueryError{
					err:                 ErrNoConnectionsInPool,
					potentiallyExecuted: potentiallyExecuted,
				},
			}, RetryNextHost
		}
		iter = q.attemptQuery(ctx, qry, conn)
		iter.host = selectedHost.Info()
		// Update host
		if iter.err == nil {
			return iter, RetryType(255)
		}

		switch {
		case errors.Is(iter.err, context.Canceled),
			errors.Is(iter.err, context.DeadlineExceeded):
			selectedHost.Mark(nil)
			potentiallyExecuted = true
			retry = Rethrow
		default:
			selectedHost.Mark(iter.err)
			retry = RetryType(255) // Don't enforce retry and get it from retry policy
		}

		var qErr *QueryError
		if errors.As(iter.err, &qErr) {
			potentiallyExecuted = potentiallyExecuted || qErr.PotentiallyExecuted()
			qErr.potentiallyExecuted = potentiallyExecuted
			qErr.isIdempotent = qry.IsIdempotent()
			iter.err = qErr
		} else {
			iter.err = &QueryError{
				err:                 iter.err,
				potentiallyExecuted: potentiallyExecuted,
				isIdempotent:        qry.IsIdempotent(),
			}
		}
		return iter, retry
	}

	var lastErr error
	selectedHost := hostIter()
	for selectedHost != nil {
		iter, retryType := execute(qry, selectedHost)
		if iter.err == nil {
			return iter
		}
		lastErr = iter.err

		// Exit if retry policy decides to not retry anymore
		if retryType == RetryType(255) {
			if !getShouldRetry(qry) {
				return iter
			}
			retryType = getRetryType(iter.err)
		}

		// If query is unsuccessful, check the error with RetryPolicy to retry
		switch retryType {
		case Retry:
			// retry on the same host
			continue
		case Rethrow, Ignore:
			return iter
		case RetryNextHost:
			// retry on the next host
			selectedHost = hostIter()
			continue
		default:
			// Undefined? Return nil and error, this will panic in the requester
			return &Iter{err: ErrUnknownRetryType}
		}
	}
	if lastErr != nil {
		return &Iter{err: lastErr}
	}
	return &Iter{err: ErrNoConnections}
}

func (q *queryExecutor) run(ctx context.Context, qry ExecutableQuery, hostIter NextHost, results chan<- *Iter) {
	select {
	case results <- q.do(ctx, qry, hostIter):
	case <-ctx.Done():
	}
	qry.releaseAfterExecution()
}
