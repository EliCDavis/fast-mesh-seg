package main

import (
	"encoding/binary"
	"fmt"
	"io"
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
	id              uint64
}

// NewNode creates a new node and calculates some properties required to write to file
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
		NumProperties:   uint64(len(properties) + len(arrayProperties)),
		PropertyListLen: propertyLength,
		NameLen:         uint8(len(name)),
		Name:            name,
		Properties:      properties,
		ArrayProperties: arrayProperties,
		NestedNodes:     nestedNodes,
		Length:          nestedLength + propertyLength + uint64(len(name)) + 25, // 8 + 8 + 8 + 1
	}
}

// NewNodeSingleProperty creates a new node that only has one property
func NewNodeSingleProperty(name string, property *Property) *Node {
	return NewNode(name, []*Property{property}, nil, nil)
}

func NewNodeSingleArrayProperty(name string, property *ArrayProperty) *Node {
	return NewNode(name, nil, []*ArrayProperty{property}, nil)
}

func NewNodeInt32Slice(name string, i []int32) *Node {
	return NewNodeSingleArrayProperty(name, NewArrayPropertyInt32Slice(i))
}

func NewNodeFloat64Slice(name string, i []float64) *Node {
	return NewNodeSingleArrayProperty(name, NewArrayPropertyFloat64Slice(i))
}

// NewNodeInt32 creates a new node with a single int32 property
func NewNodeInt32(name string, i int32) *Node {
	return NewNode(name, []*Property{NewPropertyInt32(i)}, nil, nil)
}

// NewNodeInt64 creates a new node with a single int64 property
func NewNodeInt64(name string, i int64) *Node {
	return NewNode(name, []*Property{NewPropertyInt64(i)}, nil, nil)
}

// NewNodeString creates a new node with a single string property
func NewNodeString(name string, s string) *Node {
	return NewNode(name, []*Property{NewPropertyString(s)}, nil, nil)
}

// NewNodeParent creates a node who only has children node, no properties
func NewNodeParent(name string, children ...*Node) *Node {
	return NewNode(name, nil, nil, children)
}

// ShallowCopy returns a new node and shallow copies of any array type
// contained within the struct
func (n Node) ShallowCopy() *Node {
	props := make([]*Property, len(n.Properties))
	copy(props, n.Properties)

	arrayProps := make([]*ArrayProperty, len(n.ArrayProperties))
	copy(arrayProps, n.ArrayProperties)

	newNodes := make([]*Node, len(n.NestedNodes))
	copy(newNodes, n.NestedNodes)

	return &Node{
		NumProperties:   n.NumProperties,
		PropertyListLen: n.PropertyListLen,
		NameLen:         n.NameLen,
		Name:            n.Name,
		Properties:      props,
		ArrayProperties: arrayProps,
		NestedNodes:     newNodes,
		Length:          n.Length,
		id:              n.id,
	}
}

func (n *Node) ApplyDiffs(diffs []Diff) (*Node, []Diff) {

	if n.Length == 0 {
		return n, diffs
	}

	remainingDiffs := make([]Diff, 0)
	diffedNode := n
	for _, diff := range diffs {
		actuallyDiffed := false
		diffedNode, actuallyDiffed = diff.Apply(diffedNode)
		if actuallyDiffed == false {
			remainingDiffs = append(remainingDiffs, diff)
		}
	}

	for i, nested := range diffedNode.NestedNodes {
		diffedNode.NestedNodes[i], remainingDiffs = nested.ApplyDiffs(remainingDiffs)
	}

	var propertyLength uint64
	for _, p := range diffedNode.Properties {
		propertyLength += p.Size()
	}

	for _, p := range diffedNode.ArrayProperties {
		propertyLength += p.Size()
	}
	diffedNode.PropertyListLen = propertyLength

	var nestedLength uint64
	for _, n := range diffedNode.NestedNodes {
		if n == nil {
			continue
		}
		// Length of 0 denotes empty node, but it still takes up space when
		// we write it to disk, empty nodes take up 25 bytes
		if n.Length == 0 {
			nestedLength += 25
		} else {
			nestedLength += n.Length
		}
	}

	diffedNode.Length = nestedLength + propertyLength + uint64(len(diffedNode.Name)) + 25
	diffedNode.NameLen = uint8(len(diffedNode.Name))
	diffedNode.NumProperties = uint64(len(diffedNode.Properties) + len(diffedNode.ArrayProperties))

	return diffedNode, remainingDiffs
}

func (node Node) Write(writer io.Writer, currentOffset uint64, endOfList bool) (uint64, error) {
	offset := node.Length
	if offset != 0 {
		offset += currentOffset
	}
	err := binary.Write(writer, binary.LittleEndian, uint64(offset))
	if err != nil {
		return 0, err
	}

	err = binary.Write(writer, binary.LittleEndian, node.NumProperties)
	if err != nil {
		return 0, err
	}

	err = binary.Write(writer, binary.LittleEndian, node.PropertyListLen)
	if err != nil {
		return 0, err
	}

	err = binary.Write(writer, binary.LittleEndian, node.NameLen)
	if err != nil {
		return 0, err
	}

	_, err = writer.Write([]byte(node.Name))
	if err != nil {
		return 0, err
	}

	for _, p := range node.ArrayProperties {
		err := p.Write(writer)
		if err != nil {
			return 0, nil
		}
	}

	for _, p := range node.Properties {
		err := p.Write(writer)
		if err != nil {
			return 0, nil
		}
	}

	offsetSofar := currentOffset + 25 + uint64(node.NameLen) + node.PropertyListLen
	for i, p := range node.NestedNodes {
		offsetSofar, err = p.Write(writer, offsetSofar, len(node.NestedNodes)-1 == i)
		if err != nil {
			return 0, nil
		}
	}

	// bytesWritten := 0
	// if len(node.NestedNodes) > 0 {
	// 	_, err = writer.Write([]byte{
	// 		// bytesWritten, err = writer.Write([]byte{
	// 		0, 0, 0, 0, 0,
	// 		0, 0, 0, 0, 0,
	// 		0, 0, 0, 0, 0,
	// 		0, 0, 0, 0, 0,
	// 		0, 0, 0, 0, 0,
	// 	})
	// }

	return uint64(node.Length + currentOffset), err
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

func (node *Node) StringProperty() (string, bool) {
	if len(node.Properties) != 1 {
		return "", false
	}
	return node.Properties[0].AsString(), true
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
