package main

import "strings"

// NodeStack acts as a stack for nodes. Meant for minimal resizing and doesn't
// really need much of the functionality found in most stacks
type NodeStack struct {
	position int
	data     []*Node
}

func NewNodeStack() *NodeStack {
	return &NodeStack{
		position: -1,
		data:     make([]*Node, 0),
	}
}

func (s *NodeStack) push(n *Node) {
	if s.position == len(s.data)-1 {
		s.data = append(s.data, n)
	} else {
		s.data[s.position+1] = n
	}
	s.position++
}

func (s *NodeStack) pop() {
	s.position--
}

func (s NodeStack) String() string {
	sb := strings.Builder{}
	for i := 0; i <= s.position; i++ {
		sb.WriteString(s.data[i].Name)
		if i < s.position {
			sb.WriteString("/")
		}
	}
	return sb.String()
}
