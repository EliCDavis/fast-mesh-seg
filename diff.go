package main

// Diff represents a thing to be changed in the original FBX format
type Diff interface {
	Apply(n *Node) (*Node, bool)
	NodeID() uint64
}

// SortDiff sorts diffs based on node ID
type SortDiff []Diff

func (a SortDiff) Len() int           { return len(a) }
func (a SortDiff) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SortDiff) Less(i, j int) bool { return a[i].NodeID() < a[j].NodeID() }
