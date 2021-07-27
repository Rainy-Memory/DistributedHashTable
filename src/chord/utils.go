package chord

import (
	"crypto/sha1"
	"errors"
	log "github.com/sirupsen/logrus"
	"math/big"
	"net"
	"net/rpc"
	"time"
)

const (
	M                 = 160
	NULL              = ""
	attempt           = 3
	SuccessorListLen  = 5
	dialPauseTime     = 500 * time.Millisecond
	pingPauseTime     = 500 * time.Millisecond
	maintainPauseTime = 100 * time.Millisecond
)

var (
	two     = big.NewInt(2)
	mod     = new(big.Int).Exp(two, big.NewInt(int64(M)), nil)
	ordinal = [attempt]string{"first", "second", "third"}
)

type Pair struct {
	First  string
	Second string
}

func id(x string) (ret *big.Int) {
	h := sha1.New()
	h.Write([]byte(x))
	ret = new(big.Int)
	ret.SetBytes(h.Sum(nil))
	return
}

func powOf2(power int) *big.Int {
	return new(big.Int).Exp(two, big.NewInt(int64(power)), nil)
}

func start(nId *big.Int, i int) *big.Int {
	return new(big.Int).Mod(new(big.Int).Add(nId, powOf2(i)), mod)
}

func within(tar, start, end *big.Int, endClosed bool) bool {
	if start.Cmp(end) < 0 {
		if endClosed {
			return start.Cmp(tar) < 0 && tar.Cmp(end) <= 0
		} else {
			return start.Cmp(tar) < 0 && tar.Cmp(end) < 0
		}
	} else {
		if endClosed {
			return start.Cmp(tar) < 0 || tar.Cmp(end) <= 0
		} else {
			return start.Cmp(tar) < 0 || tar.Cmp(end) < 0
		}
	}
}

func Dial(addr string) (*rpc.Client, error) {
	if addr == NULL {
		log.Errorf("Dial a null address.")
		return nil, errors.New("dial a null address")
	}
	var client *rpc.Client
	errorChannel := make(chan error)
	for i := 0; i < attempt; i++ {
		go func() {
			var err error
			client, err = rpc.Dial("tcp", addr)
			errorChannel <- err
		}()
		select {
		case err := <-errorChannel:
			if err == nil {
				log.Tracef("Dial address %v success.", addr)
				return client, nil
			} else {
				log.Tracef("Dial address [%v] failed, error message: [%v]", addr, err)
				return nil, err
			}
		case <-time.After(dialPauseTime):
			log.Tracef("Dial address %v the %v time encountered a time out error.", addr, ordinal[i])
		}
	}
	log.Errorf("Dial address %v time out.", addr)
	return nil, errors.New("dial time out")
}

func Ping(addr string) bool {
	if addr == NULL {
		log.Tracef("Ping a null address.")
		return false
	}
	errorChannel := make(chan error)
	for i := 0; i < attempt; i++ {
		go func() {
			client, err := rpc.Dial("tcp", addr)
			if err == nil {
				_ = client.Close()
			}
			errorChannel <- err
		}()
		select {
		case err := <-errorChannel:
			if err == nil {
				log.Tracef("Ping address %v success.", addr)
				return true
			} else {
				log.Tracef("Ping address [%v] failed, error message: [%v]", addr, err)
				return false
			}
		case <-time.After(pingPauseTime):
			log.Tracef("Ping address %v the %v time encountered a time out error.", addr, ordinal[i])
		}
	}
	log.Errorf("Ping address %v time out.", addr)
	return false
}

func logErrorFunctionCall(addr, fromFunc, toFunc string, err error) {
	log.Errorf("[Addr:%v] In call from [%v] to [%v] failed, error message: [%v].", addr, fromFunc, toFunc, err)
}

func CloseClient(client *rpc.Client) {
	err := client.Close()
	if err != nil {
		log.Errorf("[Error] Close client failed, error message: [%v]", err)
	}
}

func Accept(server *rpc.Server, listener net.Listener, n *ChordNode) {
	for {
		conn, err := listener.Accept()
		select {
		case <-n.quitSignal:
			return
		default:
			if err != nil {
				log.Print("rpc.Serve: accept:", err.Error())
				return
			}
			go server.ServeConn(conn)
		}
	}
}

func RPCCall(addr string, serviceMethod string, args interface{}, reply interface{}) error {
	client, err := Dial(addr)
	if err != nil {
		log.Errorf("Dial address [%v] failed in RPCCall, error message: [%v].", addr, err)
		return err
	}
	defer CloseClient(client)
	err = client.Call(serviceMethod, args, reply)
	if err != nil {
		log.Errorf("Calling function [%v] failed in RPCCall, error message: [%v].", serviceMethod, err)
		return err
	}
	return nil
}
