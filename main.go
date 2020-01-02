package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

func loadModel() *FBX {
	defer timeTrack(time.Now(), "Load Model")
	f, err := os.Open("./dragon_vrip.fbx")
	check(err)
	defer f.Close()

	reader := NewReaderWithFilters(FilterName("Objects/Geometry")) // FilterName("Objects/Geometry")
	reader.ReadFrom(f)
	check(reader.Error)
	return reader.FBX
}

func main() {

	out, err := os.Create("out.txt")
	check(err)

	fbx := loadModel()

	expand(out, fbx.Top)
	for _, c := range fbx.Nodes {
		expand(out, c)
	}

	expand(os.Stdout, fbx.GetNode("Objects", "Geometry"))
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

	if string(p.TypeCode) == "L" {
		return fmt.Sprint(p.AsInt64())
	}

	if string(p.TypeCode) == "d" {
		s, _ := p.AsFloat64Slice()
		return fmt.Sprintf("[float64 array len: %d]", len(s))
	}

	if string(p.TypeCode) == "i" {
		s, _ := p.AsInt32Slice()
		return fmt.Sprintf("[int32 array len: %d]", len(s))
	}

	return "typecode: " + string(p.TypeCode)
}

func expand(out *os.File, node *Node) {
	for i := 0; i < depth; i++ {
		out.WriteString("--")
	}
	out.WriteString("-> ")
	out.WriteString(node.Name + "\n")

	for _, p := range node.Properties {
		for i := 0; i < depth; i++ {
			out.WriteString("--")
		}
		out.WriteString("---- ")
		out.WriteString(propertyToString(p) + "\n")
	}

	depth++
	for _, child := range node.NestedNodes {
		expand(out, child)
	}
	depth--
}
