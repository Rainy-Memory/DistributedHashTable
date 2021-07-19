package chord

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"math/big"
	"net/rpc"
	"sync"
)

const (
	M                int = 160
	SuccessorListLen int = 5
)

func GetClient(addr string) (*rpc.Client, error) {
	client, err := rpc.Dial("tcp", addr)
	if err != nil {
		return nil, err
	} else {
		return client, nil
	}
}

type ChordNode struct {
	addr        string
	predecessor string
	fingerTable [M]string
	successor   [SuccessorListLen]string

	store     SecureMap
	storeLock sync.RWMutex
}

func InitializeFingerTable(addr string) {
	for i := 0; i < M; i++ {

	}
}

func (n *ChordNode) Get(key string) (ok bool, val string) {
	var tarAddr string
	_ = n.FindSuccessor(id(key), &tarAddr)
	client, err := GetClient(tarAddr)
	if err != nil {
		log.Errorln("GetClient failed in ChordNode.Get")
		return false, ""
	}
	err = client.Call("ChordNode.Query", key, &val)
	if err != nil {
		log.Errorln("Not Found in ChordNode.Get")
		return false, ""
	}
	ok = true
	return
}

func (n *ChordNode) Query(key string, val *string) error {
	n.storeLock.RLock()
	defer n.storeLock.RUnlock()
	var ok bool
	*val, ok = n.store[key]
	if !ok {
		*val = ""
		return errors.New("not found")
	}
	return nil
}

func (n *ChordNode) FindSuccessor(kId *big.Int, ret *string) error {
	client := n.FindPredecessor(kId)
	_ = client.Call("ChordNode.GetSuccessor", "", ret)
	return nil
}

func (n *ChordNode) GetSuccessor(_ string, ret *string) error {
	*ret = n.fingerTable[0].node
	return nil
}

func (n *ChordNode) FindPredecessor(kId *big.Int) *rpc.Client {
	nId := id(n.addr)
	sucStr := n.fingerTable[0].node
	sucId := id(sucStr)
	client, _ := GetClient(n.addr)
	for !within(kId, nId, sucId, true) {
		var nextAddr string
		_ = client.Call("ChordNode.ClosetPrecedingFinger", kId, &nextAddr)
		client, _ = GetClient(nextAddr)
		nId = id(nextAddr)
		_ = client.Call("ChordNode.GetSuccessor", "", &sucStr)
		sucId = id(sucStr)
	}
	return client
}

func (n *ChordNode) ClosetPrecedingFinger(kId *big.Int, ret *string) error {
	nId := id(n.addr)
	for i := M - 1; i >= 0; i-- {
		if within(id(n.fingerTable[i].node), nId, kId, false) {
			*ret = n.fingerTable[i].node
			return nil
		}
	}
	*ret = n.addr
	return nil
}
