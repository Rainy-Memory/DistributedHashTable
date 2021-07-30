package chord

import (
	"errors"
	"math/big"
	"net"
	"net/rpc"
	"sync"
	"time"
)

type ChordNode struct {
	addr          string
	predecessor   string
	preLock       sync.RWMutex
	successorList [SuccessorListLen]string
	sucLock       sync.RWMutex
	fingerTable   [M]string
	fingerLock    sync.RWMutex
	next          int

	store         map[string]string
	storeLock     sync.RWMutex
	preBackup     map[string]string
	preBackupLock sync.RWMutex

	online     bool
	onlineLock sync.RWMutex
	server     *rpc.Server
	listener   net.Listener
	quitSignal chan bool
}

func (n *ChordNode) initialize(addr string) {
	n.addr = addr
	n.store = make(map[string]string)
	n.preBackup = make(map[string]string)
	n.quitSignal = make(chan bool, 2)
}

func (n *ChordNode) FindSuccessor(kId *big.Int, ret *string) error {
	var suc string
	err := n.FirstAvailableSuccessor(NULL, &suc)
	if err != nil {
		return err
	}
	if within(kId, id(n.addr), id(suc), true) {
		*ret = suc
		return nil
	}
	var cpf string
	cpf, err = n.closestPrecedingFinger(kId)
	if err != nil {
		return err
	}
	return RPCCall(cpf, "ChordNode.FindSuccessor", kId, ret)
}

func (n *ChordNode) FirstAvailableSuccessor(_ string, ret *string) error {
	n.sucLock.RLock()
	suc0 := n.successorList[0]
	n.sucLock.RUnlock()
	if Ping(suc0) {
		*ret = suc0
		return nil
	}
	for i := 1; i < SuccessorListLen; i++ {
		n.sucLock.RLock()
		sucI := n.successorList[i]
		n.sucLock.RUnlock()
		if sucI != NULL && Ping(sucI) {
			*ret = sucI
			n.sucLock.Lock()
			for j := i; j < SuccessorListLen; j++ {
				n.successorList[j-i] = n.successorList[j]
			}
			n.sucLock.Unlock()
			time.Sleep(maintainPauseTime * 2)
			_ = RPCCall(sucI, "ChordNode.Notify", n.addr, nil)
			return nil
		}
	}
	*ret = NULL
	return errors.New("no available successor")
}

func (n *ChordNode) closestPrecedingFinger(kId *big.Int) (string, error) {
	nId := id(n.addr)
	n.fingerLock.RLock()
	defer n.fingerLock.RUnlock()
	for i := M - 1; i >= 0; i-- {
		finI := n.fingerTable[i]
		if finI != NULL && Ping(finI) && within(id(finI), nId, kId, false) {
			return finI, nil
		}
	}
	var suc string
	err := n.FirstAvailableSuccessor(NULL, &suc)
	if err != nil {
		return NULL, errors.New("not found")
	}
	return suc, nil
}

func (n *ChordNode) GetPredecessor(_ string, ret *string) error {
	n.preLock.RLock()
	*ret = n.predecessor
	n.preLock.RUnlock()
	return nil
}

func (n *ChordNode) SetPredecessor(pre string, _ *string) error {
	n.preLock.Lock()
	n.predecessor = pre
	n.preLock.Unlock()
	return nil
}

func (n *ChordNode) GetSuccessorList(_ string, ret *[SuccessorListLen]string) error {
	n.sucLock.RLock()
	*ret = n.successorList
	n.sucLock.RUnlock()
	return nil
}

func (n *ChordNode) initializeServer() {
	n.server = rpc.NewServer()
	err := n.server.Register(n)
	if err != nil {
		return
	}
	n.listener, err = net.Listen("tcp", n.addr)
	if err != nil {
		return
	}
	go Accept(n.server, n.listener, n)
}

func (n *ChordNode) run() {
	n.initializeServer()
	n.maintain()
}

func (n *ChordNode) Notify(nAlter string, _ *string) error {
	var pre string
	_ = n.GetPredecessor(NULL, &pre)
	if pre == NULL || pre != nAlter && within(id(nAlter), id(pre), id(n.addr), false) {
		_ = n.SetPredecessor(nAlter, nil)
		n.mergeBackup()
		n.updateSuccessorBackupAfterMerge()
		_ = RPCCall(nAlter, "ChordNode.GetStore", NULL, &n.preBackup)
	}
	return nil
}

func (n *ChordNode) stabilize() {
	var suc string
	err := n.FirstAvailableSuccessor(NULL, &suc)
	if err != nil {
		return
	}
	var x string
	_ = RPCCall(suc, "ChordNode.GetPredecessor", NULL, &x)

	if x != NULL && Ping(x) && within(id(x), id(n.addr), id(suc), false) {
		suc = x
	}
	var list [SuccessorListLen]string
	_ = RPCCall(suc, "ChordNode.GetSuccessorList", NULL, &list)
	n.sucLock.Lock()
	n.successorList[0] = suc
	cnt := 1
	for i := 1; i < SuccessorListLen; i++ {
		if Ping(list[i-1]) {
			n.successorList[cnt] = list[i-1]
			cnt++
		}
	}
	n.sucLock.Unlock()
	_ = RPCCall(suc, "ChordNode.Notify", n.addr, nil)
}

func (n *ChordNode) Stabilize(_ string, _ *string) error {
	n.stabilize()
	return nil
}

func (n *ChordNode) fixFinger() {
	var suc string
	tar := start(id(n.addr), n.next)
	err := n.FindSuccessor(tar, &suc)
	if err != nil {
		return
	}
	n.fingerLock.Lock()
	if n.fingerTable[n.next] != suc {
		n.fingerTable[n.next] = suc
	}
	n.fingerLock.Unlock()
	n.next = (n.next + 1) % M
}

func (n *ChordNode) checkPredecessor() {
	var pre string
	_ = n.GetPredecessor(NULL, &pre)
	if pre != NULL && !Ping(pre) {
		_ = n.SetPredecessor(NULL, nil)
		n.mergeBackup()
		n.updateSuccessorBackupAfterMerge()
	}
}

func (n *ChordNode) CheckPredecessor(_ string, _ *string) error {
	n.checkPredecessor()
	return nil
}

func (n *ChordNode) maintain() {
	go func() {
		for {
			if n.online {
				n.stabilize()
			}
			time.Sleep(maintainPauseTime)
		}
	}()
	go func() {
		for {
			if n.online {
				n.fixFinger()
			}
			time.Sleep(maintainPauseTime)
		}
	}()
	go func() {
		for {
			if n.online {
				n.checkPredecessor()
			}
			time.Sleep(maintainPauseTime)
		}
	}()
}

func (n *ChordNode) create() {
	n.onlineLock.Lock()
	n.online = true
	n.onlineLock.Unlock()
	n.sucLock.Lock()
	n.successorList[0] = n.addr
	n.sucLock.Unlock()
	_ = n.SetPredecessor(n.addr, nil)
	n.fingerLock.Lock()
	for i := 0; i < M; i++ {
		n.fingerTable[i] = n.addr
	}
	n.fingerLock.Unlock()
}

func (n *ChordNode) EraseRedundantPreBackup(redundant *map[string]string, _ *string) error {
	n.preBackupLock.Lock()
	for k := range *redundant {
		delete(n.preBackup, k)
	}
	n.preBackupLock.Unlock()
	return nil
}

func (n *ChordNode) TransferData(pre string, preStore *map[string]string) error {
	nId := id(pre)
	thisId := id(n.addr)
	n.storeLock.Lock()
	n.preBackupLock.Lock()
	n.preBackup = make(map[string]string)
	for k, v := range n.store {
		if !within(id(k), nId, thisId, true) {
			(*preStore)[k] = v
			n.preBackup[k] = v
			delete(n.store, k)
		}
	}
	n.storeLock.Unlock()
	n.preBackupLock.Unlock()
	var suc string
	err := n.FirstAvailableSuccessor(NULL, &suc)
	if err != nil {
		return err
	}
	if suc != pre {
		_ = RPCCall(suc, "ChordNode.EraseRedundantPreBackup", preStore, nil)
	}
	return nil
}

func (n *ChordNode) join(addr string) bool {
	if n.online {
		return false
	}
	_ = n.SetPredecessor(NULL, nil)
	var suc string
	err := RPCCall(addr, "ChordNode.FindSuccessor", id(n.addr), &suc)
	if err != nil {
		return false
	}
	var list [SuccessorListLen]string
	_ = RPCCall(suc, "ChordNode.GetSuccessorList", NULL, &list)
	n.sucLock.Lock()
	n.successorList[0] = suc
	cnt := 1
	for i := 1; i < SuccessorListLen; i++ {
		if Ping(list[i-1]) {
			n.successorList[cnt] = list[i-1]
			cnt++
		}
	}
	n.sucLock.Unlock()
	if suc != n.addr {
		n.storeLock.Lock()
		_ = RPCCall(suc, "ChordNode.TransferData", n.addr, &n.store)
		n.storeLock.Unlock()
	}
	n.fingerLock.Lock()
	n.fingerTable[0] = suc
	n.fingerLock.Unlock()
	nId := id(n.addr)
	for i := 1; i < M; i++ {
		var finI string
		err = RPCCall(suc, "ChordNode.FindSuccessor", start(nId, i), &finI)
		if err != nil {
			finI = NULL
		}
		n.fingerLock.Lock()
		n.fingerTable[i] = finI
		n.fingerLock.Unlock()
	}
	n.onlineLock.Lock()
	n.online = true
	n.onlineLock.Unlock()
	return true
}

func (n *ChordNode) AppendPreBackup(appendStore *map[string]string, _ *string) error {
	n.preBackupLock.Lock()
	for k, v := range *appendStore {
		n.preBackup[k] = v
	}
	n.preBackupLock.Unlock()
	return nil
}

func (n *ChordNode) mergeBackup() {
	n.storeLock.Lock()
	n.preBackupLock.RLock()
	for k, v := range n.preBackup {
		n.store[k] = v
	}
	n.storeLock.Unlock()
	n.preBackupLock.RUnlock()
}

func (n *ChordNode) updateSuccessorBackupAfterMerge() {
	var suc string
	err := n.FirstAvailableSuccessor(NULL, &suc)
	if err != nil {
		return
	}
	if suc != n.addr {
		n.preBackupLock.Lock()
		_ = RPCCall(suc, "ChordNode.AppendPreBackup", &n.preBackup, nil)
		n.preBackup = make(map[string]string)
		n.preBackupLock.Unlock()
	}
}

func (n *ChordNode) GetStore(_ string, ret *map[string]string) error {
	n.storeLock.RLock()
	*ret = make(map[string]string)
	for k, v := range n.store {
		(*ret)[k] = v
	}
	n.storeLock.RUnlock()
	return nil
}

func (n *ChordNode) shutDownServer() {
	n.onlineLock.Lock()
	n.online = false
	n.onlineLock.Unlock()
	n.quitSignal <- true
	err := n.listener.Close()
	if err != nil {
		return
	}
}

func (n *ChordNode) clear() {
	n.storeLock.Lock()
	n.store = make(map[string]string)
	n.storeLock.Unlock()
	n.preBackupLock.Lock()
	n.preBackup = make(map[string]string)
	n.preBackupLock.Unlock()
	n.quitSignal = make(chan bool, 2)
}

func (n *ChordNode) quit() {
	if !n.online {
		return
	}
	n.shutDownServer()
	var suc, pre string
	_ = n.GetPredecessor(NULL, &pre)
	err := n.FirstAvailableSuccessor(NULL, &suc)
	if err != nil {
		return
	}
	err = RPCCall(suc, "ChordNode.CheckPredecessor", NULL, nil)
	if err != nil {
		return
	}
	err = RPCCall(pre, "ChordNode.Stabilize", NULL, nil)
	if err != nil {
		return
	}
	n.clear()
}

func (n *ChordNode) forceQuit() {
	if !n.online {
		return
	}
	n.shutDownServer()
	n.clear()
}

func (n *ChordNode) ping(addr string) bool {
	return Ping(addr)
}

func (n *ChordNode) put(key string, val string) bool {
	if !n.online {
		return false
	}
	var tar string
	err := n.FindSuccessor(id(key), &tar)
	if err != nil {
		return false
	}
	err = RPCCall(tar, "ChordNode.PutInStore", Pair{First: key, Second: val}, nil)
	if err != nil {
		return false
	}
	return true
}

func (n *ChordNode) PutInStore(kv Pair, _ *string) error {
	n.storeLock.Lock()
	n.store[kv.First] = kv.Second
	n.storeLock.Unlock()
	var suc string
	err := n.FirstAvailableSuccessor(NULL, &suc)
	if err != nil {
		return err
	}
	_ = RPCCall(suc, "ChordNode.PutInPreBackup", kv, nil)
	return nil
}

func (n *ChordNode) PutInPreBackup(kv Pair, _ *string) error {
	n.preBackupLock.Lock()
	n.preBackup[kv.First] = kv.Second
	n.preBackupLock.Unlock()
	return nil
}

func (n *ChordNode) get(key string) (ok bool, val string) {
	if !n.online {
		return false, NULL
	}
	var tar string
	err := n.FindSuccessor(id(key), &tar)
	if err != nil {
		return false, NULL
	}
	err = RPCCall(tar, "ChordNode.GetInStore", key, &val)
	if err != nil {
		return false, NULL
	}
	ok = true
	return
}

func (n *ChordNode) GetInStore(key string, val *string) error {
	var ok bool
	n.storeLock.RLock()
	*val, ok = n.store[key]
	n.storeLock.RUnlock()
	if !ok {
		*val = NULL
		return errors.New("not found")
	}
	return nil
}

func (n *ChordNode) delete(key string) bool {
	if !n.online {
		return false
	}
	var tar string
	err := n.FindSuccessor(id(key), &tar)
	if err != nil {
		return false
	}
	err = RPCCall(tar, "ChordNode.DeleteInStore", key, nil)
	if err != nil {
		return false
	}
	return true
}

func (n *ChordNode) DeleteInStore(key string, _ *string) error {
	n.storeLock.Lock()
	_, ok := n.store[key]
	delete(n.store, key)
	n.storeLock.Unlock()
	if !ok {
		return errors.New("trying to delete nonexistent key in store")
	}
	var suc string
	err := n.FirstAvailableSuccessor(NULL, &suc)
	if err != nil {
		return err
	}
	err = RPCCall(suc, "ChordNode.DeleteInPreBackup", key, nil)
	if err != nil {
		return err
	}
	return nil
}

func (n *ChordNode) DeleteInPreBackup(key string, _ *string) error {
	n.preBackupLock.Lock()
	_, ok := n.preBackup[key]
	delete(n.preBackup, key)
	n.preBackupLock.Unlock()
	if !ok {
		return errors.New("trying to delete nonexistent key in pre backup")
	}
	return nil
}
