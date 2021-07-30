package main

import (
	"chord"
	"strconv"
)

/* In this file, you should implement function "NewNode" and
 * a struct which implements the interface "dhtNode".
 */

func NewNode(port int) *chord.NodeWrapper {
	// create a node and then return it.
	var n chord.NodeWrapper
	n.Initialize(GetLocalAddress() + ":" + strconv.Itoa(port))
	return &n
}
