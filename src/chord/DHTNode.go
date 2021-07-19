package chord

import (
	"fmt"
	"net/rpc"
)

type Node struct {
	node   *ChordNode
	server *rpc.Server
}

func (n *Node) Initialize(localAddress string, port int) {
	n.node = &ChordNode{
		addr:  localAddress + ":" + fmt.Sprintf("%v", port),
		store: make(map[string]string),
	}
}

func (n *Node) Run() {
	n.server = rpc.NewServer()

}

func (n *Node) Create() {

}

func (n *Node) Join(addr string) bool {

}

func (n *Node) Quit() {

}

func (n *Node) ForceQuit() {

}

func (n *Node) Ping(addr string) bool {

}

func (n *Node) Put(key string, value string) bool {

}

func (n *Node) Get(key string) (bool, string) {
	return n.node.Get(key)
}

func (n *Node) Delete(key string) bool {

}
