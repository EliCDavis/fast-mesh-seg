package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/EliCDavis/mesh"

	"github.com/EliCDavis/vector"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

var timer Timer

func loadModel(modelName string, jobs chan<- []*Node) *FBX {
	f, err := os.Open(modelName)
	check(err)
	defer f.Close()

	reader := NewReaderWithFilters(
		MatchStackAndSubNodes("Objects/Geometry", "Vertices", "PolygonVertexIndex"),
		jobs,
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
	// defer timeTrack(time.Now(), fmt.Sprintf("Saving Model (%d tris) as '%s'", len(mesh.GetFaces()), name))
	timer.begin(fmt.Sprintf("Saving Model (%d tris) as '%s'", len(mesh.GetFaces()), name))
	defer timer.end()

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

// SplitByPlane accumulates all geometry nodes and splits them by some plane
func SplitByPlane(geomNode *Node, clippingPlane Plane) ([]mesh.Polygon, []mesh.Polygon) {
	// timer.begin("Splitting model by plane")
	// defer timer.end()

	vertexNodes := geomNode.GetNodes("Vertices")
	if len(vertexNodes) == 0 {
		return nil, nil
	}

	polyVertexNodes := geomNode.GetNodes("PolygonVertexIndex")
	if len(polyVertexNodes) == 0 {
		return nil, nil
	}

	vertice, _ := vertexNodes[0].Float64Slice()
	verticeIndexes, _ := polyVertexNodes[0].Int32Slice()

	numFaces := len(verticeIndexes) / 3

	clippedPolygons := make([]mesh.Polygon, numFaces)
	clippedPolyIndex := 0
	retainedPolygons := make([]mesh.Polygon, numFaces)
	retainedPolyIndex := 0

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

		aDist := clippingPlane.normal.Dot(points[0].Sub(clippingPlane.origin))
		bDist := clippingPlane.normal.Dot(points[1].Sub(clippingPlane.origin))
		cDist := clippingPlane.normal.Dot(points[2].Sub(clippingPlane.origin))
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
			retainedPolygons[retainedPolyIndex] = p
			retainedPolyIndex++
		}
		if neg == 3 {
			clippedPolygons[clippedPolyIndex] = p
			clippedPolyIndex++
		}
	}

	return retainedPolygons[:retainedPolyIndex], clippedPolygons[:clippedPolyIndex]
}

func worker(id int, plane Plane, jobs <-chan []*Node, results chan<- Result) {
	allRetainedPolygons := make([]mesh.Polygon, 0)
	allClippedPolygons := make([]mesh.Polygon, 0)

	for j := range jobs {
		for _, n := range j {
			retained, clipped := SplitByPlane(n, plane)
			allRetainedPolygons = append(allRetainedPolygons, retained...)
			allClippedPolygons = append(allClippedPolygons, clipped...)
		}
	}
	results <- Result{clipped: allClippedPolygons, retained: allRetainedPolygons}
}

func SplitByPlaneProgram(modelName string, plane Plane, workers int) (*FBX, mesh.Model, mesh.Model) {
	timer.begin(fmt.Sprintf("Loading and splitting %s by plane with %d workers", modelName, workers))
	defer timer.end()

	jobs := make(chan []*Node, 10000)
	results := make(chan Result, 10000)

	// start workers before attempting to load model
	for w := 0; w < workers; w++ {
		go worker(w, plane, jobs, results)
	}

	go loadModel(modelName, jobs)

	allRetainedPolygons := make([]mesh.Polygon, 0)
	allClippedPolygons := make([]mesh.Polygon, 0)

	for i := 0; i < workers; i++ {
		r := <-results
		allRetainedPolygons = append(allRetainedPolygons, r.retained...)
		allClippedPolygons = append(allClippedPolygons, r.clipped...)
	}

	retained, _ := mesh.NewModel(allRetainedPolygons)
	clipped, _ := mesh.NewModel(allClippedPolygons)

	return nil, retained, clipped
}

func main() {

	// out, err := os.Create("out.txt")
	// check(err)

	// _, retained, clipped := SplitByPlaneProgram("dragon_vrip.fbx", NewPlane(vector.Vector3Zero(), vector.Vector3Forward()), 3)
	// log.Printf("Retained Model Polygon Count: %d", len(retained.GetFaces()))
	// log.Printf("Clipped Model Polygon Count: %d", len(clipped.GetFaces()))

	// expand(out, fbx.Top)
	// for _, c := range fbx.Nodes {
	// 	expand(out, c)
	// }

	// save(retained, "retained.obj")
	// save(clipped, "clipped.obj")

	_, retained, clipped := SplitByPlaneProgram("HIB-model.fbx", NewPlane(vector.NewVector3(105.4350, 119.4877, 77.9060), vector.Vector3Up()), 3)
	log.Printf("Retained Model Polygon Count: %d", len(retained.GetFaces()))
	log.Printf("Clipped Model Polygon Count: %d", len(clipped.GetFaces()))
	// log.Print(retained.GetCenterOfBoundingBox())

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

	return "typecode: " + string(p.TypeCode)
}

func arrayPropertyToString(p *ArrayProperty) string {
	if p == nil {
		return "nil property"
	}

	if string(p.TypeCode) == "d" {
		s := p.AsFloat64Slice()
		return fmt.Sprintf("[float64 array len: %d]", len(s))
	}

	if string(p.TypeCode) == "i" {
		s := p.AsInt32Slice()
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

	for _, p := range node.ArrayProperties {
		for i := 0; i < depth; i++ {
			out.WriteString("--")
		}
		out.WriteString("---- ")
		out.WriteString(arrayPropertyToString(p) + "\n")
	}

	depth++
	for _, child := range node.NestedNodes {
		expand(out, child)
	}
	depth--
}
