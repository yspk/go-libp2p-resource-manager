package rcmgr

import (
	"testing"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-libp2p-core/test"
	"github.com/multiformats/go-multiaddr"
)

var dummyMA = multiaddr.StringCast("/ip4/1.2.3.4/tcp/1234")

func TestResourceManager(t *testing.T) {
	peerA := peer.ID("A")
	peerB := peer.ID("B")
	protoA := protocol.ID("/A")
	protoB := protocol.ID("/B")
	svcA := "A.svc"
	svcB := "B.svc"
	nmgr, err := NewResourceManager(
		NewFixedLimiter(LimitConfig{
			System: BaseLimit{
				Memory:          16384,
				StreamsInbound:  3,
				StreamsOutbound: 3,
				Streams:         6,
				ConnsInbound:    3,
				ConnsOutbound:   3,
				Conns:           6,
				FD:              2,
			},
			Transient: BaseLimit{
				Memory:          4096,
				StreamsInbound:  1,
				StreamsOutbound: 1,
				Streams:         2,
				ConnsInbound:    1,
				ConnsOutbound:   1,
				Conns:           2,
				FD:              1,
			},
			ServiceDefault: BaseLimit{
				Memory:          4096,
				StreamsInbound:  1,
				StreamsOutbound: 1,
				Streams:         2,
				ConnsInbound:    1,
				ConnsOutbound:   1,
				Conns:           2,
				FD:              1,
			},
			ServicePeerDefault: BaseLimit{
				Memory:          4096,
				StreamsInbound:  5,
				StreamsOutbound: 5,
				Streams:         10,
			},
			Service: map[string]BaseLimit{
				svcA: {
					Memory:          8192,
					StreamsInbound:  2,
					StreamsOutbound: 2,
					Streams:         4,
					ConnsInbound:    2,
					ConnsOutbound:   2,
					Conns:           4,
					FD:              1,
				},
				svcB: {
					Memory:          8192,
					StreamsInbound:  2,
					StreamsOutbound: 2,
					Streams:         4,
					ConnsInbound:    2,
					ConnsOutbound:   2,
					Conns:           4,
					FD:              1,
				},
			},
			ServicePeer: map[string]BaseLimit{
				svcB: {
					Memory:          8192,
					StreamsInbound:  1,
					StreamsOutbound: 1,
					Streams:         2,
				},
			},
			ProtocolDefault: BaseLimit{
				Memory:          4096,
				StreamsInbound:  1,
				StreamsOutbound: 1,
				Streams:         2,
			},
			Protocol: map[protocol.ID]BaseLimit{
				protoA: {
					Memory:          8192,
					StreamsInbound:  2,
					StreamsOutbound: 2,
					Streams:         2,
				},
			},
			ProtocolPeer: map[protocol.ID]BaseLimit{
				protoB: {
					Memory:          8192,
					StreamsInbound:  1,
					StreamsOutbound: 1,
					Streams:         2,
				},
			},
			PeerDefault: BaseLimit{
				Memory:          4096,
				StreamsInbound:  1,
				StreamsOutbound: 1,
				Streams:         2,
				ConnsInbound:    1,
				ConnsOutbound:   1,
				Conns:           2,
				FD:              1,
			},
			ProtocolPeerDefault: BaseLimit{
				Memory:          4096,
				StreamsInbound:  5,
				StreamsOutbound: 5,
				Streams:         10,
			},
			Peer: map[peer.ID]BaseLimit{
				peerA: {
					Memory:          8192,
					StreamsInbound:  2,
					StreamsOutbound: 2,
					Streams:         4,
					ConnsInbound:    2,
					ConnsOutbound:   2,
					Conns:           4,
					FD:              1,
				},
			},
			Conn: BaseLimit{
				Memory:        4096,
				ConnsInbound:  1,
				ConnsOutbound: 1,
				Conns:         1,
				FD:            1,
			},
			Stream: BaseLimit{
				Memory:          4096,
				StreamsInbound:  1,
				StreamsOutbound: 1,
				Streams:         1,
			},
		}),
	)

	if err != nil {
		t.Fatal(err)
	}

	mgr := nmgr.(*resourceManager)
	defer mgr.Close()

	checkRefCnt := func(s *resourceScope, count int) {
		t.Helper()
		if refCnt := s.refCnt; refCnt != count {
			t.Fatalf("expected refCnt of %d, got %d", count, refCnt)
		}
	}
	checkSystem := func(check func(s *resourceScope)) {
		if err := mgr.ViewSystem(func(s network.ResourceScope) error {
			check(s.(*systemScope).resourceScope)
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	checkTransient := func(check func(s *resourceScope)) {
		if err := mgr.ViewTransient(func(s network.ResourceScope) error {
			check(s.(*transientScope).resourceScope)
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	checkService := func(svc string, check func(s *resourceScope)) {
		if err := mgr.ViewService(svc, func(s network.ServiceScope) error {
			check(s.(*serviceScope).resourceScope)
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	checkProtocol := func(p protocol.ID, check func(s *resourceScope)) {
		if err := mgr.ViewProtocol(p, func(s network.ProtocolScope) error {
			check(s.(*protocolScope).resourceScope)
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	checkPeer := func(p peer.ID, check func(s *resourceScope)) {
		if err := mgr.ViewPeer(p, func(s network.PeerScope) error {
			check(s.(*peerScope).resourceScope)
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}

	checkSystem(func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{})
	})
	checkTransient(func(s *resourceScope) {
		checkRefCnt(s, 1)
		checkResources(t, &s.rc, network.ScopeStat{})
	})

	// open an inbound connection, using an fd
	conn, err := mgr.OpenConnection(network.DirInbound, true, dummyMA)
	if err != nil {
		t.Fatal(err)
	}

	checkSystem(func(s *resourceScope) {
		checkRefCnt(s, 3)
		checkResources(t, &s.rc, network.ScopeStat{NumConnsInbound: 1, NumFD: 1})
	})
	checkTransient(func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumConnsInbound: 1, NumFD: 1})
	})

	// the connection is transient, we shouldn't be able to open a second one
	if _, err := mgr.OpenConnection(network.DirInbound, true, dummyMA); err == nil {
		t.Fatal("expected OpenConnection to fail")
	}
	if _, err := mgr.OpenConnection(network.DirInbound, false, dummyMA); err == nil {
		t.Fatal("expected OpenConnection to fail")
	}

	// close it to check resources are reclaimed
	conn.Done()

	checkSystem(func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{})
	})
	checkTransient(func(s *resourceScope) {
		checkRefCnt(s, 1)
		checkResources(t, &s.rc, network.ScopeStat{})
	})

	// open another inbound connection, using an fd
	conn1, err := mgr.OpenConnection(network.DirInbound, true, dummyMA)
	if err != nil {
		t.Fatal(err)
	}

	checkSystem(func(s *resourceScope) {
		checkRefCnt(s, 3)
		checkResources(t, &s.rc, network.ScopeStat{NumConnsInbound: 1, NumFD: 1})
	})
	checkTransient(func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumConnsInbound: 1, NumFD: 1})
	})

	// check nility of current peer scope
	if conn1.PeerScope() != nil {
		t.Fatal("peer scope should be nil")
	}

	// attach to a peer
	if err := conn1.SetPeer(peerA); err != nil {
		t.Fatal(err)
	}

	checkPeer(peerA, func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumConnsInbound: 1, NumFD: 1})
	})
	checkSystem(func(s *resourceScope) {
		checkRefCnt(s, 4)
		checkResources(t, &s.rc, network.ScopeStat{NumConnsInbound: 1, NumFD: 1})
	})
	checkTransient(func(s *resourceScope) {
		checkRefCnt(s, 1)
		checkResources(t, &s.rc, network.ScopeStat{})
	})

	// we should be able to open a second transient connection now
	conn2, err := mgr.OpenConnection(network.DirInbound, true, dummyMA)
	if err != nil {
		t.Fatal(err)
	}

	checkSystem(func(s *resourceScope) {
		checkRefCnt(s, 5)
		checkResources(t, &s.rc, network.ScopeStat{NumConnsInbound: 2, NumFD: 2})
	})
	checkTransient(func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumConnsInbound: 1, NumFD: 1})
	})

	// but we shouldn't be able to attach it to the same peer due to the fd limit
	if err := conn2.SetPeer(peerA); err == nil {
		t.Fatal("expected SetPeer to fail")
	}

	checkPeer(peerA, func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumConnsInbound: 1, NumFD: 1})
	})
	checkSystem(func(s *resourceScope) {
		checkRefCnt(s, 5)
		checkResources(t, &s.rc, network.ScopeStat{NumConnsInbound: 2, NumFD: 2})
	})
	checkTransient(func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumConnsInbound: 1, NumFD: 1})
	})

	// close it and reopen without using an FD -- we should be able to attach now
	conn2.Done()

	conn2, err = mgr.OpenConnection(network.DirInbound, false, dummyMA)
	if err != nil {
		t.Fatal(err)
	}

	checkSystem(func(s *resourceScope) {
		checkRefCnt(s, 5)
		checkResources(t, &s.rc, network.ScopeStat{NumConnsInbound: 2, NumFD: 1})
	})
	checkTransient(func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumConnsInbound: 1, NumFD: 0})
	})

	if err := conn2.SetPeer(peerA); err != nil {
		t.Fatal(err)
	}

	checkPeer(peerA, func(s *resourceScope) {
		checkRefCnt(s, 3)
		checkResources(t, &s.rc, network.ScopeStat{NumConnsInbound: 2, NumFD: 1})
	})
	checkSystem(func(s *resourceScope) {
		checkRefCnt(s, 5)
		checkResources(t, &s.rc, network.ScopeStat{NumConnsInbound: 2, NumFD: 1})
	})
	checkTransient(func(s *resourceScope) {
		checkRefCnt(s, 1)
		checkResources(t, &s.rc, network.ScopeStat{})
	})

	// open a stream
	stream, err := mgr.OpenStream(peerA, network.DirInbound)
	if err != nil {
		t.Fatal(err)
	}

	checkPeer(peerA, func(s *resourceScope) {
		checkRefCnt(s, 4)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1, NumConnsInbound: 2, NumFD: 1})
	})
	checkSystem(func(s *resourceScope) {
		checkRefCnt(s, 6)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1, NumConnsInbound: 2, NumFD: 1})
	})
	checkTransient(func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1})
	})

	// the stream is transient we shouldn't be able to open a second one
	if _, err := mgr.OpenStream(peerA, network.DirInbound); err == nil {
		t.Fatal("expected OpenStream to fail")
	}

	// close the stream to check resource reclamation
	stream.Done()

	checkPeer(peerA, func(s *resourceScope) {
		checkRefCnt(s, 3)
		checkResources(t, &s.rc, network.ScopeStat{NumConnsInbound: 2, NumFD: 1})
	})
	checkSystem(func(s *resourceScope) {
		checkRefCnt(s, 5)
		checkResources(t, &s.rc, network.ScopeStat{NumConnsInbound: 2, NumFD: 1})
	})
	checkTransient(func(s *resourceScope) {
		checkRefCnt(s, 1)
		checkResources(t, &s.rc, network.ScopeStat{})
	})

	// open another stream, but this time attach it to a protocol
	stream1, err := mgr.OpenStream(peerA, network.DirInbound)
	if err != nil {
		t.Fatal(err)
	}

	checkPeer(peerA, func(s *resourceScope) {
		checkRefCnt(s, 4)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1, NumConnsInbound: 2, NumFD: 1})
	})
	checkSystem(func(s *resourceScope) {
		checkRefCnt(s, 6)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1, NumConnsInbound: 2, NumFD: 1})
	})
	checkTransient(func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1})
	})

	// check nility of protocol scope
	if stream1.ProtocolScope() != nil {
		t.Fatal("protocol scope should be nil")
	}

	if err := stream1.SetProtocol(protoA); err != nil {
		t.Fatal(err)
	}

	checkPeer(peerA, func(s *resourceScope) {
		checkRefCnt(s, 4)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1, NumConnsInbound: 2, NumFD: 1})
	})
	checkProtocol(protoA, func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1})
	})
	checkSystem(func(s *resourceScope) {
		checkRefCnt(s, 7)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1, NumConnsInbound: 2, NumFD: 1})
	})
	checkTransient(func(s *resourceScope) {
		checkRefCnt(s, 1)
		checkResources(t, &s.rc, network.ScopeStat{})
	})

	// and now we should be able to open another stream and attach it to the protocol
	stream2, err := mgr.OpenStream(peerA, network.DirInbound)
	if err != nil {
		t.Fatal(err)
	}

	checkPeer(peerA, func(s *resourceScope) {
		checkRefCnt(s, 5)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 2, NumConnsInbound: 2, NumFD: 1})
	})
	checkSystem(func(s *resourceScope) {
		checkRefCnt(s, 8)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 2, NumConnsInbound: 2, NumFD: 1})
	})
	checkTransient(func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1})
	})

	if err := stream2.SetProtocol(protoA); err != nil {
		t.Fatal(err)
	}

	checkPeer(peerA, func(s *resourceScope) {
		checkRefCnt(s, 5)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 2, NumConnsInbound: 2, NumFD: 1})
	})
	checkProtocol(protoA, func(s *resourceScope) {
		checkRefCnt(s, 3)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 2})
	})
	checkSystem(func(s *resourceScope) {
		checkRefCnt(s, 8)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 2, NumConnsInbound: 2, NumFD: 1})
	})
	checkTransient(func(s *resourceScope) {
		checkRefCnt(s, 1)
		checkResources(t, &s.rc, network.ScopeStat{})
	})

	// open a 3rd stream, and try to attach it to the same protocol
	stream3, err := mgr.OpenStream(peerB, network.DirInbound)
	if err != nil {
		t.Fatal(err)
	}

	checkPeer(peerA, func(s *resourceScope) {
		checkRefCnt(s, 5)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 2, NumConnsInbound: 2, NumFD: 1})
	})
	checkPeer(peerB, func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1})
	})
	checkSystem(func(s *resourceScope) {
		checkRefCnt(s, 10)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 3, NumConnsInbound: 2, NumFD: 1})
	})
	checkTransient(func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1})
	})

	if err := stream3.SetProtocol(protoA); err == nil {
		t.Fatal("expected SetProtocol to fail")
	}

	checkPeer(peerA, func(s *resourceScope) {
		checkRefCnt(s, 5)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 2, NumConnsInbound: 2, NumFD: 1})
	})
	checkPeer(peerB, func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1})
	})
	checkProtocol(protoA, func(s *resourceScope) {
		checkRefCnt(s, 3)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 2})
	})
	checkSystem(func(s *resourceScope) {
		checkRefCnt(s, 10)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 3, NumConnsInbound: 2, NumFD: 1})
	})
	checkTransient(func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1})
	})

	// but we should be able to set to another protocol
	if err := stream3.SetProtocol(protoB); err != nil {
		t.Fatal(err)
	}

	checkPeer(peerA, func(s *resourceScope) {
		checkRefCnt(s, 5)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 2, NumConnsInbound: 2, NumFD: 1})
	})
	checkPeer(peerB, func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1})
	})
	checkProtocol(protoA, func(s *resourceScope) {
		checkRefCnt(s, 3)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 2})
	})
	checkProtocol(protoB, func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1})
	})
	checkSystem(func(s *resourceScope) {
		checkRefCnt(s, 11)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 3, NumConnsInbound: 2, NumFD: 1})
	})
	checkTransient(func(s *resourceScope) {
		checkRefCnt(s, 1)
		checkResources(t, &s.rc, network.ScopeStat{})
	})

	// check nility of current service scope
	if stream1.ServiceScope() != nil {
		t.Fatal("service scope should be nil")
	}

	// we should be able to attach stream1 and stream2 to svcA, but stream3 should fail due to limit
	if err := stream1.SetService(svcA); err != nil {
		t.Fatal(err)
	}

	checkPeer(peerA, func(s *resourceScope) {
		checkRefCnt(s, 5)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 2, NumConnsInbound: 2, NumFD: 1})
	})
	checkPeer(peerB, func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1})
	})
	checkService(svcA, func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1})
	})
	checkProtocol(protoA, func(s *resourceScope) {
		checkRefCnt(s, 3)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 2})
	})
	checkProtocol(protoB, func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1})
	})
	checkSystem(func(s *resourceScope) {
		checkRefCnt(s, 12)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 3, NumConnsInbound: 2, NumFD: 1})
	})
	checkTransient(func(s *resourceScope) {
		checkRefCnt(s, 1)
		checkResources(t, &s.rc, network.ScopeStat{})
	})

	if err := stream2.SetService(svcA); err != nil {
		t.Fatal(err)
	}

	checkPeer(peerA, func(s *resourceScope) {
		checkRefCnt(s, 5)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 2, NumConnsInbound: 2, NumFD: 1})
	})
	checkPeer(peerB, func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1})
	})
	checkService(svcA, func(s *resourceScope) {
		checkRefCnt(s, 3)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 2})
	})
	checkProtocol(protoA, func(s *resourceScope) {
		checkRefCnt(s, 3)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 2})
	})
	checkProtocol(protoB, func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1})
	})
	checkSystem(func(s *resourceScope) {
		checkRefCnt(s, 12)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 3, NumConnsInbound: 2, NumFD: 1})
	})
	checkTransient(func(s *resourceScope) {
		checkRefCnt(s, 1)
		checkResources(t, &s.rc, network.ScopeStat{})
	})

	if err := stream3.SetService(svcA); err == nil {
		t.Fatal("expected SetService to fail")
	}

	checkPeer(peerA, func(s *resourceScope) {
		checkRefCnt(s, 5)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 2, NumConnsInbound: 2, NumFD: 1})
	})
	checkPeer(peerB, func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1})
	})
	checkService(svcA, func(s *resourceScope) {
		checkRefCnt(s, 3)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 2})
	})
	checkProtocol(protoA, func(s *resourceScope) {
		checkRefCnt(s, 3)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 2})
	})
	checkProtocol(protoB, func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1})
	})
	checkSystem(func(s *resourceScope) {
		checkRefCnt(s, 12)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 3, NumConnsInbound: 2, NumFD: 1})
	})
	checkTransient(func(s *resourceScope) {
		checkRefCnt(s, 1)
		checkResources(t, &s.rc, network.ScopeStat{})
	})

	// and now let's reclaim our resources to make sure we can gc unused peer and proto scopes
	// but first check internal refs
	mgr.mx.Lock()
	_, okProtoA := mgr.proto[protoA]
	_, okProtoB := mgr.proto[protoB]
	_, okPeerA := mgr.peer[peerA]
	_, okPeerB := mgr.peer[peerB]
	mgr.mx.Unlock()

	if !okProtoA {
		t.Fatal("protocol scope is not stored")
	}
	if !okProtoB {
		t.Fatal("protocol scope is not stored")
	}
	if !okPeerA {
		t.Fatal("peer scope is not stored")
	}
	if !okPeerB {
		t.Fatal("peer scope is not stored")
	}

	// ok, reclaim
	stream1.Done()
	stream2.Done()
	stream3.Done()
	conn1.Done()
	conn2.Done()

	// check everything released
	checkPeer(peerA, func(s *resourceScope) {
		checkRefCnt(s, 1)
		checkResources(t, &s.rc, network.ScopeStat{})
	})
	checkPeer(peerB, func(s *resourceScope) {
		checkRefCnt(s, 1)
		checkResources(t, &s.rc, network.ScopeStat{})
	})
	checkService(svcA, func(s *resourceScope) {
		checkRefCnt(s, 1)
		checkResources(t, &s.rc, network.ScopeStat{})
	})
	checkProtocol(protoA, func(s *resourceScope) {
		checkRefCnt(s, 1)
		checkResources(t, &s.rc, network.ScopeStat{})
	})
	checkProtocol(protoB, func(s *resourceScope) {
		checkRefCnt(s, 1)
		checkResources(t, &s.rc, network.ScopeStat{})
	})
	checkSystem(func(s *resourceScope) {
		checkRefCnt(s, 7)
		checkResources(t, &s.rc, network.ScopeStat{})
	})
	checkTransient(func(s *resourceScope) {
		checkRefCnt(s, 1)
		checkResources(t, &s.rc, network.ScopeStat{})
	})

	mgr.gc()

	checkSystem(func(s *resourceScope) {
		checkRefCnt(s, 3)
		checkResources(t, &s.rc, network.ScopeStat{})
	})
	checkTransient(func(s *resourceScope) {
		checkRefCnt(s, 1)
		checkResources(t, &s.rc, network.ScopeStat{})
	})

	mgr.mx.Lock()
	lenProto := len(mgr.proto)
	lenPeer := len(mgr.peer)
	mgr.mx.Unlock()

	if lenProto != 0 {
		t.Fatal("protocols were not gc'ed")
	}
	if lenPeer != 0 {
		t.Fatal("perrs were not gc'ed")
	}

	// check that per protocol peer scopes work as intended
	stream1, err = mgr.OpenStream(peerA, network.DirInbound)
	if err != nil {
		t.Fatal(err)
	}

	checkPeer(peerA, func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1})
	})
	checkSystem(func(s *resourceScope) {
		checkRefCnt(s, 5)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1})
	})
	checkTransient(func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1})
	})

	if err := stream1.SetProtocol(protoB); err != nil {
		t.Fatal(err)
	}

	checkPeer(peerA, func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1})
	})
	checkProtocol(protoB, func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1})
	})
	checkSystem(func(s *resourceScope) {
		checkRefCnt(s, 6)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1})
	})
	checkTransient(func(s *resourceScope) {
		checkRefCnt(s, 1)
		checkResources(t, &s.rc, network.ScopeStat{})
	})

	stream2, err = mgr.OpenStream(peerA, network.DirInbound)
	if err != nil {
		t.Fatal(err)
	}

	checkPeer(peerA, func(s *resourceScope) {
		checkRefCnt(s, 3)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 2})
	})
	checkSystem(func(s *resourceScope) {
		checkRefCnt(s, 7)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 2})
	})
	checkTransient(func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1})
	})

	if err := stream2.SetProtocol(protoB); err == nil {
		t.Fatal("expected SetProtocol to fail")
	}

	checkPeer(peerA, func(s *resourceScope) {
		checkRefCnt(s, 3)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 2})
	})
	checkSystem(func(s *resourceScope) {
		checkRefCnt(s, 7)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 2})
	})
	checkTransient(func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1})
	})
	checkProtocol(protoB, func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1})
	})

	stream1.Done()
	stream2.Done()

	// check that per service peer scopes work as intended
	stream1, err = mgr.OpenStream(peerA, network.DirInbound)
	if err != nil {
		t.Fatal(err)
	}

	checkPeer(peerA, func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1})
	})
	checkSystem(func(s *resourceScope) {
		checkRefCnt(s, 6)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1})
	})
	checkTransient(func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1})
	})

	if err := stream1.SetProtocol(protoA); err != nil {
		t.Fatal(err)
	}

	checkPeer(peerA, func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1})
	})
	checkProtocol(protoA, func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1})
	})
	checkSystem(func(s *resourceScope) {
		checkRefCnt(s, 7)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1})
	})
	checkTransient(func(s *resourceScope) {
		checkRefCnt(s, 1)
		checkResources(t, &s.rc, network.ScopeStat{})
	})

	stream2, err = mgr.OpenStream(peerA, network.DirInbound)
	if err != nil {
		t.Fatal(err)
	}

	checkPeer(peerA, func(s *resourceScope) {
		checkRefCnt(s, 3)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 2})
	})
	checkSystem(func(s *resourceScope) {
		checkRefCnt(s, 8)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 2})
	})
	checkTransient(func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1})
	})

	if err := stream2.SetProtocol(protoA); err != nil {
		t.Fatal(err)
	}

	checkPeer(peerA, func(s *resourceScope) {
		checkRefCnt(s, 3)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 2})
	})
	checkProtocol(protoA, func(s *resourceScope) {
		checkRefCnt(s, 3)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 2})
	})
	checkSystem(func(s *resourceScope) {
		checkRefCnt(s, 8)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 2})
	})
	checkTransient(func(s *resourceScope) {
		checkRefCnt(s, 1)
		checkResources(t, &s.rc, network.ScopeStat{})
	})

	if err := stream1.SetService(svcB); err != nil {
		t.Fatal(err)
	}

	checkPeer(peerA, func(s *resourceScope) {
		checkRefCnt(s, 3)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 2})
	})
	checkService(svcB, func(s *resourceScope) {
		checkRefCnt(s, 2)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 1})
	})
	checkProtocol(protoA, func(s *resourceScope) {
		checkRefCnt(s, 3)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 2})
	})
	checkSystem(func(s *resourceScope) {
		checkRefCnt(s, 9)
		checkResources(t, &s.rc, network.ScopeStat{NumStreamsInbound: 2})
	})
	checkTransient(func(s *resourceScope) {
		checkRefCnt(s, 1)
		checkResources(t, &s.rc, network.ScopeStat{})
	})

	// now we should fail to set the service for stream2 to svcB because of the service peer limit
	if err := stream2.SetService(svcB); err == nil {
		t.Fatal("expected SetService to fail")
	}

	// now release resources and check interior gc of per service peer scopes
	stream1.Done()
	stream2.Done()

	mgr.gc()

	checkSystem(func(s *resourceScope) {
		checkRefCnt(s, 4)
		checkResources(t, &s.rc, network.ScopeStat{})
	})
	checkTransient(func(s *resourceScope) {
		checkRefCnt(s, 1)
		checkResources(t, &s.rc, network.ScopeStat{})
	})

	mgr.mx.Lock()
	lenProto = len(mgr.proto)
	lenPeer = len(mgr.peer)
	mgr.mx.Unlock()

	svc := mgr.svc[svcB]
	svc.Lock()
	lenSvcPeer := len(svc.peers)
	svc.Unlock()

	if lenProto != 0 {
		t.Fatal("protocols were not gc'ed")
	}
	if lenPeer != 0 {
		t.Fatal("peers were not gc'ed")
	}
	if lenSvcPeer != 0 {
		t.Fatal("service peers were not gc'ed")
	}

}

func TestResourceManagerWithAllowlist(t *testing.T) {
	peerA := test.RandPeerIDFatal(t)

	limits := DefaultLimits.Scale(1<<30, 100)
	limits.System.Conns = 0
	limits.System.ConnsInbound = 0
	limits.System.ConnsOutbound = 0
	limits.Transient.Conns = 0
	limits.Transient.ConnsInbound = 0
	limits.Transient.ConnsOutbound = 0
	rcmgr, err := NewResourceManager(NewFixedLimiter(limits), WithAllowlistedMultiaddrs([]multiaddr.Multiaddr{
		multiaddr.StringCast("/ip4/1.2.3.4"),
		multiaddr.StringCast("/ip4/4.3.2.1/p2p/" + peerA.String()),
	}))
	if err != nil {
		t.Fatal(err)
	}

	{
		// Setup allowlist. TODO, replace this with a config once config changes are in
		r := rcmgr.(*resourceManager)

		sysLimit := limits.System
		sysLimit.Conns = 2
		sysLimit.ConnsInbound = 2
		sysLimit.ConnsOutbound = 1
		r.allowlistedSystem = newSystemScope(&sysLimit, r, "allowlistedSystem")
		r.allowlistedSystem.IncRef()

		transLimit := limits.Transient
		transLimit.Conns = 1
		transLimit.ConnsInbound = 1
		transLimit.ConnsOutbound = 1
		r.allowlistedTransient = newTransientScope(&transLimit, r, "allowlistedTransient", r.allowlistedSystem.resourceScope)
		r.allowlistedTransient.IncRef()
	}

	// A connection comes in from a non-allowlisted ip address
	_, err = rcmgr.OpenConnection(network.DirInbound, true, multiaddr.StringCast("/ip4/1.2.3.5"))
	if err == nil {
		t.Fatalf("Expected this to fail. err=%v", err)
	}

	// A connection comes in from an allowlisted ip address
	connScope, err := rcmgr.OpenConnection(network.DirInbound, true, multiaddr.StringCast("/ip4/1.2.3.4"))
	if err != nil {
		t.Fatal(err)
	}

	err = connScope.SetPeer(test.RandPeerIDFatal(t))
	if err != nil {
		t.Fatal(err)
	}

	// A connection comes in that looks like it should be allowlisted, but then has the wrong peer id.
	connScope, err = rcmgr.OpenConnection(network.DirInbound, true, multiaddr.StringCast("/ip4/4.3.2.1"))
	if err != nil {
		t.Fatal(err)
	}

	err = connScope.SetPeer(test.RandPeerIDFatal(t))
	if err == nil {
		t.Fatalf("Expected this to fail. err=%v", err)
	}

	// A connection comes in that looks like it should be allowlisted, and it has the allowlisted peer id
	connScope, err = rcmgr.OpenConnection(network.DirInbound, true, multiaddr.StringCast("/ip4/4.3.2.1"))
	if err != nil {
		t.Fatal(err)
	}

	err = connScope.SetPeer(peerA)
	if err != nil {
		t.Fatal(err)
	}
}
