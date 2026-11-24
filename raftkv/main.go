package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"raftkv/raft"
)

func main() {
	// Read command line flags
	// id , addr, and peers are the code that make the same programm bechanve like a different server each time . It reads  -id -addr and -peer from the command line
	id := flag.Int("id", 0, "unique ID of this server (1, 2, 3...)")
	addr := flag.String("addr", "", "address to listen on, e.g. localhost:5001")
	peerFlag := flag.String("peers", "", "other servers, e.g. 2=localhost:5002,3=localhost:5003")
	flag.Parse()

	if *id == 0 || *addr == "" || *peerFlag == "" {
		log.Fatal("usage: go run . -id 1 -addr localhost:5001 -peers 2=localhost:5002,3=localhost:5003")
	}

	// Turn "2=localhost:5002,3=localhost:5003" into a map[int]string
	peers := parsePeers(*peerFlag)

	// Create and start this server
	node := raft.NewNode(*id, *addr, peers)
	if err := node.Start(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}

	// Keep running until Ctrl+C
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	node.Stop()
	log.Println("server stopped")
}

// parsePeers converts "2=localhost:5002,3=localhost:5003"
// into map[int]string{2: "localhost:5002", 3: "localhost:5003"}
func parsePeers(s string) map[int]string {
	peers := make(map[int]string)
	for _, part := range strings.Split(s, ",") {
		idStr, addr, ok := strings.Cut(part, "=")
		if !ok {
			log.Fatalf("bad peer %q, expected id=address", part)
		}
		peerID, err := strconv.Atoi(idStr)
		if err != nil {
			log.Fatalf("bad peer id %q: %v", idStr, err)
		}
		peers[peerID] = addr
	}
	return peers
}
