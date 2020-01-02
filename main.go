package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/EliCDavis/mesh"

	"github.com/EliCDavis/vector"
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

func save(mesh mesh.Model, name string) error {
	defer timeTrack(time.Now(), "Saving Model")
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	err = mesh.Save(w)
	if err != nil {
		return err
	}
	return w.Flush()
}

func main() {

	out, err := os.Create("out.txt")
	check(err)

	fbx := loadModel()

	expand(out, fbx.Top)
	for _, c := range fbx.Nodes {
		expand(out, c)
	}

	vertice, _ := fbx.GetNode("Objects", "Geometry", "Vertices").Float64Slice()
	verticeIndexes, _ := fbx.GetNode("Objects", "Geometry", "PolygonVertexIndex").Int32Slice()

	numFaces := len(verticeIndexes) / 3
	polygons := make([]mesh.Polygon, 0)
	for f := 0; f < numFaces; f++ {
		faceIndex := f * 3
		firstInd := int(verticeIndexes[faceIndex]) * 3
		secondInd := int(verticeIndexes[faceIndex+1]) * 3
		wrapInd := (int(verticeIndexes[faceIndex+2])*-1 - 1) * 3
		points := []vector.Vector3{
			vector.NewVector3(
				vertice[firstInd],
				vertice[firstInd+1],
				vertice[firstInd+2],
			),
			vector.NewVector3(
				vertice[secondInd],
				vertice[secondInd+1],
				vertice[secondInd+2],
			),
			vector.NewVector3(
				vertice[wrapInd],
				vertice[wrapInd+1],
				vertice[wrapInd+2],
			),
		}

		p, _ := mesh.NewPolygon(
			points,
			points,
		)
		polygons = append(polygons, p)
	}

	m, err := mesh.NewModel(polygons)
	check(err)
	save(m, "out.obj")

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
