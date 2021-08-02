package chord

import (
	"fmt"
	"strconv"
)

type NodeWrapper struct {
	node *ChordNode
}

func (w *NodeWrapper) Initialize(addr string) {
	w.node = new(ChordNode)
	w.node.initialize(addr)
}

func (w *NodeWrapper) Run() {
	w.node.run()
}

func (w *NodeWrapper) Create() {
	w.node.create()
}

func (w *NodeWrapper) Join(addr string) bool {
	return w.node.join(addr)
}

func (w *NodeWrapper) Quit() {
	w.node.quit()
}

func (w *NodeWrapper) ForceQuit() {
	w.node.forceQuit()
}

func (w *NodeWrapper) Ping(addr string) bool {
	return w.node.ping(addr)
}

func (w *NodeWrapper) Put(key string, value string) bool {
	return w.node.put(key, value)
}

func (w *NodeWrapper) Get(key string) (bool, string) {
	return w.node.get(key)
}

func (w *NodeWrapper) Delete(key string) bool {
	return w.node.delete(key)
}

//   \/ --- chatroom functions --- \/

func (w *NodeWrapper) GetAddress() string {
	return w.node.addr
}

/*
storage format:
{[name] message number}           -> chatroom [name]'s message number.
{[name] [number]}                 -> chatroom [name] No.[number] message.
{[name] public key}               -> chatroom [name]'s public key.
{[name] user number}              -> chatroom [name]'s user number.
{[name] key [number]}             -> chatroom [name] No.[number] user's public key. will be deleted after successfully transfer room private key.
{[name] encrypted [number] [num]} -> temporary save chatroom [name] No.[number] user's transfer data's No.[num] part. will be deleted also.
*/

func (w *NodeWrapper) GetMessageNumber(roomName string) (bool, int) {
	ok, numStr := w.Get(appHash(roomName + " message number"))
	if !ok {
		return false, -1
	}
	ret, _ := strconv.Atoi(numStr)
	return true, ret
}

func (w *NodeWrapper) PutMessageNumber(roomName string, messageNumber int) bool {
	numStr := strconv.Itoa(messageNumber)
	return w.Put(appHash(roomName+" message number"), numStr)
}

func (w *NodeWrapper) NewRoom(roomName string) bool {
	ok := w.PutMessageNumber(roomName, 0)
	if ok {
		w.PutUserNumber(roomName, 0)
	}
	return ok
}

func (w *NodeWrapper) GetMessage(roomName string, messageNumber int) (bool, string) {
	return w.Get(appHash(roomName + " message " + strconv.Itoa(messageNumber)))
}

func (w *NodeWrapper) PutMessage(roomName string, messageNumber int, message string) {
	w.Put(appHash(roomName+" message "+strconv.Itoa(messageNumber)), message)
}

func (w *NodeWrapper) GetRoomPublicKey(roomName string) []byte {
	_, pubKeyStr := w.Get(appHash(roomName + " public key"))
	return str2bytes(pubKeyStr)
}

func (w *NodeWrapper) PutRoomPublicKey(roomName string, roomPublicKeyBytes []byte) {
	w.Put(appHash(roomName+" public key"), bytes2str(roomPublicKeyBytes))
}

func (w *NodeWrapper) GetUserNumber(roomName string) int {
	_, numStr := w.Get(appHash(roomName + " user number"))
	ret, _ := strconv.Atoi(numStr)
	return ret
}

func (w *NodeWrapper) PutUserNumber(roomName string, userNumber int) {
	numStr := strconv.Itoa(userNumber)
	w.Put(appHash(roomName+" user number"), numStr)
}

func (w *NodeWrapper) GetUserPublicKey(roomName string, userNumber int) []byte {
	_, pubKeyStr := w.Get(appHash(roomName + " key " + strconv.Itoa(userNumber)))
	return str2bytes(pubKeyStr)
}

func (w *NodeWrapper) PutUserPublicKey(roomName string, userNumber int, userPublicKeyBytes []byte) {
	w.Put(appHash(roomName+" key "+strconv.Itoa(userNumber)), bytes2str(userPublicKeyBytes))
}

func (w *NodeWrapper) DeleteUserPublicKey(roomName string, userNumber int) {
	w.Delete(appHash(roomName + " key " + strconv.Itoa(userNumber)))
}

func (w *NodeWrapper) GetUserEncryptedPart(roomName string, userNumber, partNumber int) []byte {
	_, pubKeyStr := w.Get(appHash(fmt.Sprintf("%v encrypted %v %v", roomName, userNumber, partNumber)))
	return str2bytes(pubKeyStr)
}

func (w *NodeWrapper) PutUserEncryptedPart(roomName string, userNumber, partNumber int, userPublicKeyBytesPart []byte) {
	w.Put(appHash(fmt.Sprintf("%v encrypted %v %v", roomName, userNumber, partNumber)), bytes2str(userPublicKeyBytesPart))
}

func (w *NodeWrapper) DeleteUserEncrypted(roomName string, userNumber, partNumber int) {
	for j := 0; j <= partNumber; j++ {
		w.Delete(appHash(fmt.Sprintf("%v encrypted %v %v", roomName, userNumber, partNumber)))
	}
}
