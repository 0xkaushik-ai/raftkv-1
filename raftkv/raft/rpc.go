// This defines what a ping message looks like and what a server does when it receives one .

package raft

type PingArgs struct {
	From int
}

type PingReply struct {
	From int
}

func (n *Node) Ping(args *PingArgs, reply *PingReply) error {
	n.logf("go ping from server id %d", args.From)
	reply.From = n.id
	return nil
}
