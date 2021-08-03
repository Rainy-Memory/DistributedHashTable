package main

import (
	"chord"
	"strconv"
)

func NewNode(port int) *chord.NodeWrapper {
	var n chord.NodeWrapper
	n.Initialize(GetLocalAddress() + ":" + strconv.Itoa(port))
	return &n
}
