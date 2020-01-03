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

func loadModel(modelName string) *FBX {
	defer timeTrack(time.Now(), "Loading Model: "+modelName)
	f, err := os.Open(modelName)
	check(err)
	defer f.Close()

	reader := NewReaderWithFilters(
		EITHER(
			FilterName("Objects/Geometry/Vertices"),
			FilterName("Objects/Geometry/PolygonVertexIndex"),
		),
	)
	reader.ReadFrom(f)
	check(reader.Error)
	return reader.FBX
}

func save(mesh mesh.Model, name string) error {
	defer timeTrack(time.Now(), fmt.Sprintf("Saving Model (%d tris) as '%s'", len(mesh.GetFaces()), name))
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

// ToModel accumulates all geometry nodes and combines them into a single mesh
func ToModel(geometryNodes []*Node) (mesh.Model, mesh.Model) {
	defer timeTrack(time.Now(), "Splitting model by plane")

	clippingPlane := NewPlane(vector.Vector3Zero(), vector.Vector3Forward())

	retainedPolygons := make([]mesh.Polygon, 0)
	clippedPolygons := make([]mesh.Polygon, 0)

	for _, geomNode := range geometryNodes {

		vertice, _ := geomNode.GetNodes("Vertices")[0].Float64Slice()
		verticeIndexes, _ := geomNode.GetNodes("PolygonVertexIndex")[0].Int32Slice()

		numFaces := len(verticeIndexes) / 3
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

			// d i s t a n c e = Dot ( c l i p p l a n e . N, V [ i ] . p o i n t ) − c l i p p l a n e . c ;
			aDist := clippingPlane.normal.Dot(points[0])
			bDist := clippingPlane.normal.Dot(points[1])
			cDist := clippingPlane.normal.Dot(points[2])
			pos := 0
			neg := 0

			if aDist > 0 {
				pos++
			} else {
				neg++
			}

			if bDist > 0 {
				pos++
			} else {
				neg++
			}

			if cDist > 0 {
				pos++
			} else {
				neg++
			}

			p, _ := mesh.NewPolygon(
				points,
				points,
			)

			if pos == 3 {
				retainedPolygons = append(retainedPolygons, p)
			}

			if neg == 3 {
				clippedPolygons = append(clippedPolygons, p)
			}
		}

	}

	retained, _ := mesh.NewModel(retainedPolygons)
	clipped, _ := mesh.NewModel(clippedPolygons)

	return retained, clipped
}

func main() {
	out, err := os.Create("out.txt")
	check(err)

	fbx := loadModel("dragon_vrip.fbx")

	expand(out, fbx.Top)
	for _, c := range fbx.Nodes {
		expand(out, c)
	}

	geomNodes := fbx.GetNodes("Objects", "Geometry")

	retained, clipped := ToModel(geomNodes)
	save(retained, "retained.obj")
	save(clipped, "clipped.obj")
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
