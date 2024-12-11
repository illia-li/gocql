//go:build all || unit
// +build all unit

package gocql

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/gocql/gocql/internal/tests/mock"
)

func TestGetClusterPeerInfoZeroToken(t *testing.T) {
	host_id1, _ := ParseUUID("b2035fd9-e0ca-4857-8c45-e63c00fb7c43")
	host_id2, _ := ParseUUID("4b21ee4c-acea-4267-8e20-aaed5361a0dd")
	host_id3, _ := ParseUUID("dfef4a22-b8d8-47e9-aee5-8c19d4b7a9e3")

	schema_version1, _ := ParseUUID("af810386-a694-11ef-81fa-3aea73156247")

	peersRows := []map[string]interface{}{
		{
			"data_center":     "datacenter1",
			"host_id":         host_id1,
			"peer":            "127.0.0.3",
			"preferred_ip":    "127.0.0.3",
			"rack":            "rack1",
			"release_version": "3.0.8",
			"rpc_address":     "127.0.0.3",
			"schema_version":  schema_version1,
			"tokens":          []string{"-1296227678594315580994457470329811265"},
		},
		{
			"data_center":     "datacenter1",
			"host_id":         host_id2,
			"peer":            "127.0.0.2",
			"preferred_ip":    "127.0.0.2",
			"rack":            "rack1",
			"release_version": "3.0.8",
			"rpc_address":     "127.0.0.2",
			"schema_version":  schema_version1,
			"tokens":          []string{"-1129762924682054333"},
		},
		{
			"data_center":     "datacenter2",
			"host_id":         host_id3,
			"peer":            "127.0.0.5",
			"preferred_ip":    "127.0.0.5",
			"rack":            "rack1",
			"release_version": "3.0.8",
			"rpc_address":     "127.0.0.5",
			"schema_version":  schema_version1,
			"tokens":          []string{},
		},
	}

	translateAddressPort := func(addr net.IP, port int) (net.IP, int) {
		return addr, port
	}

	var logger StdLogger
	t.Run("OmitOneZeroTokenNode", func(t *testing.T) {
		peers, err := getPeersFromQuerySystemPeers(
			peersRows,
			9042,
			translateAddressPort,
			logger,
		)

		if err != nil {
			t.Fatalf("unable to get peers: %v", err)
		}
		assertEqual(t, "peers length", 2, len(peers))
	})

	t.Run("NoZeroTokenNodes", func(t *testing.T) {
		peersRows[2]["tokens"] = []string{"-1129762924682054333"}
		peers, err := getPeersFromQuerySystemPeers(
			peersRows,
			9042,
			translateAddressPort,
			logger,
		)

		if err != nil {
			t.Fatalf("unable to get peers: %v", err)
		}
		assertEqual(t, "peers length", 3, len(peers))
	})
}

type mockConnection struct{}

func (*mockConnection) Close() {}
func (*mockConnection) exec(ctx context.Context, req frameBuilder, tracer Tracer) (*framer, error) {
	return nil, nil
}
func (*mockConnection) awaitSchemaAgreement(ctx context.Context) error     { return nil }
func (*mockConnection) executeQuery(ctx context.Context, qry *Query) *Iter { return nil }

var systemLocalResultMetadata = resultMetadata{
	flags:          0,
	pagingState:    []byte{},
	actualColCount: 18,
	columns: []ColumnInfo{{
		Keyspace: "system",
		Table:    "local",
		Name:     "key",
		TypeInfo: NativeType{proto: 4, typ: TypeVarchar},
	}, {
		Keyspace: "system",
		Table:    "local",
		Name:     "bootstrapped",
		TypeInfo: NativeType{proto: 4, typ: TypeVarchar},
	}, {
		Keyspace: "system",
		Table:    "local",
		Name:     "broadcast_address",
		TypeInfo: NativeType{proto: 4, typ: TypeInet},
	}, {
		Keyspace: "system",
		Table:    "local",
		Name:     "cluster_name",
		TypeInfo: NativeType{proto: 4, typ: TypeVarchar},
	}, {
		Keyspace: "system",
		Table:    "local",
		Name:     "cql_version",
		TypeInfo: NativeType{proto: 4, typ: TypeVarchar},
	}, {
		Keyspace: "system",
		Table:    "local",
		Name:     "data_center",
		TypeInfo: NativeType{proto: 4, typ: TypeVarchar},
	}, {
		Keyspace: "system",
		Table:    "local",
		Name:     "gossip_generation",
		TypeInfo: NativeType{proto: 4, typ: TypeInt},
	}, {
		Keyspace: "system",
		Table:    "local",
		Name:     "host_id",
		TypeInfo: NativeType{proto: 4, typ: TypeUUID},
	}, {
		Keyspace: "system",
		Table:    "local",
		Name:     "listen_address",
		TypeInfo: NativeType{proto: 4, typ: TypeInet},
	}, {
		Keyspace: "system",
		Table:    "local",
		Name:     "native_protocol_version",
		TypeInfo: NativeType{proto: 4, typ: TypeVarchar},
	}, {
		Keyspace: "system",
		Table:    "local",
		Name:     "partitioner",
		TypeInfo: NativeType{proto: 4, typ: TypeVarchar},
	}, {
		Keyspace: "system",
		Table:    "local",
		Name:     "rack",
		TypeInfo: NativeType{proto: 4, typ: TypeVarchar},
	}, {
		Keyspace: "system",
		Table:    "local",
		Name:     "release_version",
		TypeInfo: NativeType{proto: 4, typ: TypeVarchar},
	}, {
		Keyspace: "system",
		Table:    "local",
		Name:     "rpc_address",
		TypeInfo: NativeType{proto: 4, typ: TypeInet},
	}, {
		Keyspace: "system",
		Table:    "local",
		Name:     "schema_version",
		TypeInfo: NativeType{proto: 4, typ: TypeUUID},
	}, {
		Keyspace: "system",
		Table:    "local",
		Name:     "supported_features",
		TypeInfo: NativeType{proto: 4, typ: TypeVarchar},
	}, {
		Keyspace: "system",
		Table:    "local",
		Name:     "tokens",
		TypeInfo: CollectionType{
			NativeType: NativeType{proto: 4, typ: TypeSet},
			Elem:       NativeType{proto: 4, typ: TypeVarchar},
		},
	}, {
		Keyspace: "system",
		Table:    "local",
		Name:     "truncated_at",
		TypeInfo: CollectionType{
			NativeType: NativeType{proto: 4, typ: TypeMap},

			Key:  NativeType{proto: 4, typ: TypeUUID},
			Elem: NativeType{proto: 4, typ: TypeBlob},
		},
	}},
}

var systemPeersResultMetadata = resultMetadata{
	flags:          0,
	pagingState:    []byte{},
	actualColCount: 10,
	columns: []ColumnInfo{{
		Keyspace: "system",
		Table:    "local",
		Name:     "peer",
		TypeInfo: NativeType{proto: 4, typ: TypeInet},
	}, {
		Keyspace: "system",
		Table:    "local",
		Name:     "data_center",
		TypeInfo: NativeType{proto: 4, typ: TypeVarchar},
	}, {
		Keyspace: "system",
		Table:    "local",
		Name:     "host_id",
		TypeInfo: NativeType{proto: 4, typ: TypeUUID},
	}, {
		Keyspace: "system",
		Table:    "local",
		Name:     "preferred_ip",
		TypeInfo: NativeType{proto: 4, typ: TypeInet},
	}, {
		Keyspace: "system",
		Table:    "local",
		Name:     "rack",
		TypeInfo: NativeType{proto: 4, typ: TypeVarchar},
	}, {
		Keyspace: "system",
		Table:    "local",
		Name:     "release_version",
		TypeInfo: NativeType{proto: 4, typ: TypeVarchar},
	}, {
		Keyspace: "system",
		Table:    "local",
		Name:     "rpc_address",
		TypeInfo: NativeType{proto: 4, typ: TypeInet},
	}, {
		Keyspace: "system",
		Table:    "local",
		Name:     "schema_version",
		TypeInfo: NativeType{proto: 4, typ: TypeUUID},
	}, {
		Keyspace: "system",
		Table:    "local",
		Name:     "supported_features",
		TypeInfo: NativeType{proto: 4, typ: TypeVarchar},
	}, {
		Keyspace: "system",
		Table:    "local",
		Name:     "tokens",
		TypeInfo: CollectionType{
			NativeType: NativeType{proto: 4, typ: TypeSet},
			Elem:       NativeType{proto: 4, typ: TypeVarchar},
		},
	}},
}

func (*mockConnection) querySystem(ctx context.Context, query string) *Iter {
	host_id, _ := ParseUUID("045859a7-6b9f-4efd-a5e7-acd64a295e13")
	schema_version, _ := ParseUUID("daf4df2c-b708-11ef-5c25-3004361afd71")
	localData := []interface{}{"local", "COMPLETED", net.IPv4(192, 168, 100, 12), "", "3.3.1", "datacenter1", 1733834239, host_id, net.IPv4(192, 168, 100, 12), "4", "org.apache.cassandra.dht.Murmur3Partitioner", "rack1", "3.0.8", net.IPv4(192, 168, 100, 12), schema_version, "", []string{}, map[UUID]byte{}}

	host_id1, _ := ParseUUID("b953309f-6e68-41f2-baf5-0e60da317a9c")
	host_id2, _ := ParseUUID("8269e111-ea38-44bd-a73f-9d3d12cfaf78")
	schema_version, _ = ParseUUID("b6ed5bde-b318-11ef-8f58-aeba19e31273")
	peerData1 := []interface{}{net.IPv4(192, 168, 100, 13), "datacenter1", host_id1, net.IP{}, "rack1", "3.0.8", net.IPv4(192, 168, 100, 13), schema_version, "", []string{"-1032311531684407545", "-1112089412567859825"}}
	peerData2 := []interface{}{net.IPv4(192, 168, 100, 14), "datacenter1", host_id2, net.IP{}, "rack1", "3.0.8", net.IPv4(192, 168, 100, 14), schema_version, "", []string{}}

	if query == "SELECT * FROM system.local WHERE key='local'" {
		return &Iter{
			meta:    systemLocalResultMetadata,
			framer:  &mock.MockFramer{Data: marshalMetadataMust(systemLocalResultMetadata, localData)},
			numRows: 1,
			next:    nil,
		}
	} else if query == "SELECT * FROM system.peers" {
		return &Iter{
			meta:    systemPeersResultMetadata,
			framer:  &mock.MockFramer{Data: append(marshalMetadataMust(systemPeersResultMetadata, peerData1), marshalMetadataMust(systemPeersResultMetadata, peerData2)...)},
			numRows: 2,
			next:    nil,
		}
	}
	return nil
}

func (*mockConnection) getIsSchemaV2() bool { return false }
func (*mockConnection) setSchemaV2(s bool)  {}
func (*mockConnection) query(ctx context.Context, statement string, values ...interface{}) (iter *Iter) {
	return nil
}
func (*mockConnection) getScyllaSupported() scyllaSupported { return scyllaSupported{} }

type mockControlConn struct{}

func (m *mockControlConn) getConn() *connHost {
	return &connHost{
		conn: &mockConnection{},
		host: &HostInfo{},
	}
}

func (m *mockControlConn) awaitSchemaAgreement() error                                { return nil }
func (m *mockControlConn) query(statement string, values ...interface{}) (iter *Iter) { return nil }
func (m *mockControlConn) discoverProtocol(hosts []*HostInfo) (int, error)            { return 0, nil }
func (m *mockControlConn) connect(hosts []*HostInfo) error                            { return nil }
func (m *mockControlConn) close()                                                     {}
func (m *mockControlConn) getSession() *Session                                       { return nil }

func marshalMetadataMust(metadata resultMetadata, data []interface{}) [][]byte {
	if len(metadata.columns) != len(data) {
		panic("metadata length mismatch")
	}
	res := make([][]byte, len(metadata.columns))
	for id, col := range metadata.columns {
		var err error
		value := data[id]
		res[id], err = Marshal(col.TypeInfo, value)
		if err != nil {
			panic(fmt.Sprintf("unable to marshal column %d: %v", id, err))
		}
	}
	return res
}

func TestGetHosts(t *testing.T) {
	r := &ringDescriber{control: &mockControlConn{}, cfg: &ClusterConfig{}}

	hosts, _, err := r.GetHosts()
	if err != nil {
		t.Fatalf("unable to get hosts: %v", err)
	}

	// local host and one of the peers are zero token so only one peer should be returned with 2 tokens
	assertEqual(t, "hosts length", 1, len(hosts))
	assertEqual(t, "host token length", 2, len(hosts[0].tokens))
}