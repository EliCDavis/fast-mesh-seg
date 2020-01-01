package main

import (
	"fmt"
	"os"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	f, err := os.Open("./dragon_vrip.fbx")
	check(err)

	fbx, err := ReadFrom(f)
	check(err)

	out := os.Stdout

	out, err = os.Create("out.txt")

	defer f.Close()

	expand(out, fbx.Top)
	for _, c := range fbx.Nodes {
		expand(out, c)
	}
}

var depth = 0

func propertyToString(p *Property) string {
	if p == nil {
		return "nil property"
	}
	if string(p.TypeCode) == "S" {
		s := p.AsString()
		if s == "" {
			return "[Empty String]"
		}
		return s
	}
	if string(p.TypeCode) == "I" {
		return fmt.Sprint(p.AsInt32())
	}
	if string(p.TypeCode) == "D" {
		return fmt.Sprint(p.AsFloat64())
	}

	return "typecode: " + string(p.TypeCode)
}

func expand(out *os.File, node *Node) {
	for i := 0; i < depth-1; i++ {
		out.WriteString("--")
	}
	out.WriteString("-> ")
	out.WriteString(node.Name + "\n")

	for _, p := range node.Properties {
		for i := 0; i < depth; i++ {
			out.WriteString("--")
		}
		out.WriteString("-- ")
		out.WriteString(propertyToString(p) + "\n")
	}

	depth++
	for _, child := range node.NestedNodes {
		expand(out, child)
	}
	depth--
}
