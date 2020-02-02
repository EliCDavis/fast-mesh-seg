package main

// DeleteNodeDiff represents a node to be deleted.
type DeleteNodeDiff struct {
	nodeID uint64
}

// NewDeleteNodeDiff creates a new delete node diff
func NewDeleteNodeDiff(id uint64) *DeleteNodeDiff {
	return &DeleteNodeDiff{
		nodeID: id,
	}
}

// Apply deletes the node if it's id matches
func (d DeleteNodeDiff) Apply(n *Node) (*Node, bool) {
	if n.id != d.nodeID {
		return n, false
	}
	return nil, true
}

// NodeID is the id of the node we want to apply the dif too
func (d DeleteNodeDiff) NodeID() uint64 {
	return d.nodeID
}
