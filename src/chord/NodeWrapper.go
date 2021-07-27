package chord

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
