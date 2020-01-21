package main

// PropertyDiff represents a property that should change for a specific node.
type PropertyDiff struct {
	nodeID   uint64
	property *Property
}

// Apply will check the node for matching specific criteria and if it passes an
// entirely new node will be created that contains the proper diff.
func (d PropertyDiff) Apply(n *Node) (*Node, bool) {
	if n.id != d.nodeID {
		return n, false
	}

	patchedNode := n.ShallowCopy()
	for i, p := range patchedNode.Properties {
		if p.TypeCode == d.property.TypeCode {
			patchedNode.Properties[i] = d.property
			return patchedNode, true
		}
	}

	return patchedNode, true
}
