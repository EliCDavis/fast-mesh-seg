package main

import (
	"fmt"
	"strings"
)

type Node struct {
	EndOffset       uint64
	NumProperties   uint64
	PropertyListLen uint64
	NameLen         uint8
	Name            string
	Properties      []*Property
	NestedNodes     []*Node
}

func (node *Node) IsEmpty() bool {
	if node.EndOffset == 0 &&
		node.NumProperties == 0 &&
		node.PropertyListLen == 0 &&
		node.NameLen == 0 {
		return true
	}
	return false
}

// Int32Slice treats as the node only has a single property and retrieves it as
// a Int32Slice
func (node *Node) Int32Slice() ([]int32, bool) {
	properties := node.Properties
	if len(properties) != 1 {
		return nil, false
	}
	return properties[0].AsInt32Slice()
}

// Float64Slice treats as the node only has a single property and retrieves it
// as a Float64Slice
func (node *Node) Float64Slice() ([]float64, bool) {
	properties := node.Properties
	if len(properties) != 1 {
		return nil, false
	}
	return properties[0].AsFloat64Slice()
}

func (n Node) GetNodes(names ...string) []*Node {

	if len(names) == 0 {
		return []*Node{&n}
	}

	nodes := []*Node{}

	for _, c := range n.NestedNodes {
		if c.Name == names[0] {
			nodes = append(nodes, c.GetNodes(names[1:]...)...)
		}
	}

	return nodes
}

func (n *Node) String() string {
	b := strings.Builder{}
	b.WriteString(n.Name)
	b.WriteString(":")
	if len(n.Properties) > 0 {
		b.WriteString(fmt.Sprint("", n.Properties, ""))
	}
	if len(n.NestedNodes) > 0 {
		b.WriteString(fmt.Sprint("{", n.NestedNodes, "}"))
	}
	b.WriteString("\n")
	return b.String()
}
