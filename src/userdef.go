package main

import "main/chord"

/* In this file, you should implement function "NewNode" and
 * a struct which implements the interface "dhtNode".
 */

func NewNode(port int) dhtNode {
	// create a node and then return it.
	var n chord.Node
	n.Initialize(GetLocalAddress(), port)
	return &n
}
