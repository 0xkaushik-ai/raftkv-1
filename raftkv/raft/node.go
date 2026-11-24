package raft

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"time"
)

// Node is one server in our cluster
type Node struct {
	id    int
	addr  string
	peers map[int]string

	listener net.Listener
	stopCh   chan struct{}
}

// NewNode creates a server but does not start it .

func NewNode(id int, addr string, peers map[int]string) *Node {
	return &Node{
		id:     id,
		addr:   addr,
		peers:  peers,
		stopCh: make(chan struct{}),
	}
}

// Start open the network port and start  the background loops
func (n *Node) Start() error {
	// Register our RPC methods under the name "Raft"
	server := rpc.NewServer()
	if err := server.RegisterName("Raft", n); err != nil {
		return err
	}
	ln, err := net.Listen("tcp", n.addr)
	if err != nil {
		return err
	}
	n.listener = ln
	n.logf("listening on %s", n.addr)
	go n.acceptLoop(server) // handle incoming RPCs
	go n.pingLoop()         // send pings to other
	return nil

}

// Stop shuts the server down

func (n *Node) Stop() {
	close(n.stopCh)
	if n.listener != nil {
		n.listener.Close()
	}
}

// acceptLoop waits for other  servers to connect and serves their RPCs .

func (n *Node) acceptLoop(server *rpc.Server) {
	for {
		conn, err := n.listener.Accept()
		if err != nil {
			return
		}
		go server.ServeConn(conn) // each connection it own goroutine
	}
}

func (n *Node) pingLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-n.stopCh:
			return
		case <-ticker.C:
			for peerID, peerAddr := range n.peers {

				go n.sendPing(peerID, peerAddr)
			}
		}
	}
}

func (n *Node) sendPing(peerID int, peerAddr string) {

	args := &PingArgs{From: n.id}
	reply := &PingReply{}

	// Dial the peer
	if err := n.call(peerAddr, "Raft.Ping", args, reply); err != nil {
		n.logf("failed to ping server %d at %s: %v", peerID, peerAddr, err)
		return
	}
	n.logf("received ping reply from server %d", reply.From)

}

func (n *Node) call(peerAddr, method string, args any, reply any) error {
	conn, err := net.DialTimeout("tcp", peerAddr, 5*time.Second)
	if err != nil {
		return err
	}
	client := rpc.NewClient(conn)
	defer conn.Close()

	done := client.Go(method, args, reply, nil)
	select {
	case <-done.Done:
		return done.Error
	case <-time.After(5 * time.Second):
		return fmt.Errorf("RPC call timed out")
	}
}

func (n *Node) logf(format string, args ...any) {
	log.Printf("[Server %d] %s", n.id, fmt.Sprintf(format, args...))
}
