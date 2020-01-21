package main

// ArrayPropertyDiff represents an array property that should change for a specific node.
type ArrayPropertyDiff struct {
	nodeID   uint64
	property *ArrayProperty
}

// NewArrayPropertyDiff creates a new array property diff
func NewArrayPropertyDiff(id uint64, prop *ArrayProperty) *ArrayPropertyDiff {
	return &ArrayPropertyDiff{
		nodeID:   id,
		property: prop,
	}
}

// Apply will check the node for matching specific criteria and if it passes an
// entirely new node will be created that contains the proper diff.
func (d ArrayPropertyDiff) Apply(n *Node) (*Node, bool) {
	if n.id != d.nodeID {
		return n, false
	}

	patchedNode := n.ShallowCopy()
	for i, p := range patchedNode.ArrayProperties {
		if p.TypeCode == d.property.TypeCode {
			patchedNode.ArrayProperties[i] = d.property
			return patchedNode, true
		}
	}

	return patchedNode, true
}
