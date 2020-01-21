package main

// Diff represents a thing to be changed in the original FBX format
type Diff interface {
	Apply(n *Node) (*Node, bool)
}
