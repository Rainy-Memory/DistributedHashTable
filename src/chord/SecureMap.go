package chord

import "sync"

type SecureMap struct{
	store map[string]string
	lock sync.Mutex
}

