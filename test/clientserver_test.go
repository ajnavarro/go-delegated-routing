package test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-delegated-routing/client"
	proto "github.com/ipfs/go-delegated-routing/gen/proto"
	"github.com/ipfs/go-delegated-routing/server"
	"github.com/libp2p/go-libp2p-core/peer"
	multiaddr "github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multihash"
)

func TestClientServer(t *testing.T) {
	// start a server
	s := httptest.NewServer(server.FindProvidersAsyncHandler(testFindProvidersAsyncFunc))
	defer s.Close()
	// start a client
	q, err := proto.New_DelegatedRouting_Client(s.URL, proto.DelegatedRouting_Client_WithHTTPClient(s.Client()))
	if err != nil {
		t.Fatal(err)
	}
	c := client.NewClient(q)
	// verify result
	h, err := multihash.Sum([]byte("TEST"), multihash.SHA3, 4)
	if err != nil {
		t.Fatal(err)
	}
	infos, err := c.FindProviders(context.Background(), cid.NewCidV1(cid.Libp2pKey, h))
	if err != nil {
		t.Fatal(err)
	}
	if len(infos) != 1 {
		t.Fatalf("expecting 1 result, got %d", len(infos))
	}
	if infos[0].ID != testAddrInfo.ID {
		t.Errorf("expecting %#v, got %#v", testAddrInfo.ID, infos[0].ID)
	}
	if len(infos[0].Addrs) != 1 {
		t.Fatalf("expecting 1 address, got %d", len(infos[0].Addrs))
	}
	if !infos[0].Addrs[0].Equal(testAddrInfo.Addrs[0]) {
		t.Errorf("expecting %#v, got %#v", testAddrInfo.Addrs[0], infos[0].Addrs[0])
	}
	fmt.Println(infos)
}

const (
	testPeerID   = "QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N"
	testPeerAddr = "/ip4/7.7.7.7/tcp/4242"
)

var (
	testMultiaddr = multiaddr.StringCast(testPeerAddr + "/p2p/" + testPeerID)
	testAddrInfo  *peer.AddrInfo
)

func TestMain(m *testing.M) {
	var err error
	testAddrInfo, err = peer.AddrInfoFromP2pAddr(testMultiaddr)
	if err != nil {
		fmt.Printf("address info creation (%v)", err)
		os.Exit(-1)
	}
	code := m.Run()
	os.Exit(code)
}

func testFindProvidersAsyncFunc(key cid.Cid, ch chan<- client.FindProvidersAsyncResult) error {
	go func() {
		ch <- client.FindProvidersAsyncResult{AddrInfo: []peer.AddrInfo{*testAddrInfo}}
		close(ch)
	}()
	return nil
}
