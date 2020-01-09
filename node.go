package main

import (
	"fmt"
	"strings"
)

type Node struct {
	NumProperties   uint64
	PropertyListLen uint64
	NameLen         uint8
	Name            string
	Properties      []*Property
	ArrayProperties []*ArrayProperty
	NestedNodes     []*Node
	Length          uint64
}

// NewNode creates a new node and calculates some properties required to qrite to file
func NewNode(name string, properties []*Property, arrayProperties []*ArrayProperty, nestedNodes []*Node) *Node {
	var propertyLength uint64

	for _, p := range properties {
		propertyLength += p.Size()
	}

	for _, p := range arrayProperties {
		propertyLength += p.Size()
	}

	var nestedLength uint64
	for _, n := range nestedNodes {
		if n == nil {
			continue
		}
		nestedLength += n.Length
	}

	return &Node{
		Name:            name,
		NameLen:         uint8(len(name)),
		NumProperties:   uint64(len(properties) + len(arrayProperties)),
		PropertyListLen: propertyLength,
		Length:          nestedLength + propertyLength + uint64(len(name)) + 25, // 8 + 8 + 8 + 1
	}
}

// PropertyInfo looks at all properties contained within the node and computes
// how much space it takes up
// func (node Node) PropertyInfo() (int64, int64, []byte) {

// }

// Int32Slice treats as the node only has a single property and retrieves it as
// a Int32Slice
func (node *Node) Int32Slice() ([]int32, bool) {
	if len(node.ArrayProperties) != 1 {
		return nil, false
	}
	return node.ArrayProperties[0].AsInt32Slice(), true
}

// Float64Slice treats as the node only has a single property and retrieves it
// as a Float64Slice
func (node *Node) Float64Slice() ([]float64, bool) {
	if len(node.ArrayProperties) != 1 {
		return nil, false
	}
	return node.ArrayProperties[0].AsFloat64Slice(), true
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
