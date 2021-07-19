package chord

import (
	"crypto/sha1"
	log "github.com/sirupsen/logrus"
	"math/big"
	"net/rpc"
	"time"
)

const (
	attempt       = 3
	dialPauseTime = time.Second / 2
	pingPauseTime = time.Second / 2
)

func id(x string) (ret *big.Int) {
	h := sha1.New()
	h.Write([]byte(x))
	ret = new(big.Int)
	ret.SetBytes(h.Sum(nil))
	return
}

func within(tar, start, end *big.Int, endClosed bool) bool {
	if endClosed {
		return start.Cmp(tar) < 0 && tar.Cmp(end) <= 0
	} else {
		return start.Cmp(tar) < 0 && tar.Cmp(end) < 0
	}
}

func Dial(addr string) (*rpc.Client, bool) {
	for i := 0; i < attempt; i++ {
		client, err := rpc.Dial("tcp", addr)
		if err == nil {
			return client, true
		} else {
			time.Sleep(dialPauseTime)
		}
	}
	return nil, false
}

func Ping(addr string) bool {
	var ordinal = [3]string{"first", "second", "third"}
	if addr == "" {
		return false
	}
	for i := 0; i < attempt; i++ {
		errorChannel := make(chan error)
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
				log.Infof("[Info] Ping address %v success.", addr)
				return true
			} else {
				log.Error(err)
			}
		case <-time.After(pingPauseTime):
			log.Errorf("[Error] Ping address %v the %v time encountered a time out error.", addr, ordinal[i])
		}
		close(errorChannel)
	}
	log.Errorf("[Error] Ping address %v failed.", addr)
	return false
}
