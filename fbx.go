package main

type FBX struct {
	Header *Header
	Top    *Node
	Nodes  []*Node
}

// func (f *FBX) Filter(filter NodeFilter) (nodes []*Node) {
// 	for _, node := range f.Nodes {
// 		subNodes := node.Filter(filter)
// 		nodes = append(nodes, subNodes...)
// 	}
// 	return
// }

// GetNodes attempts to find a node from those contained in the fbx
func (f FBX) GetNodes(names ...string) []*Node {

	if len(names) == 0 {
		return nil
	}

	nodes := []*Node{}

	if f.Top.Name == names[0] {
		nodes = append(nodes, f.Top.GetNodes(names[1:]...)...)
	}

	for _, n := range f.Nodes {
		if n.Name == names[0] {
			nodes = append(nodes, n.GetNodes(names[1:]...)...)
		}
	}

	return nodes
}
