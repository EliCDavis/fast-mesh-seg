package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"sort"

	"github.com/EliCDavis/mesh"

	"github.com/EliCDavis/vector"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

var timer Timer

func loadModel(modelName string, jobs chan<- []*Node, fbx chan<- *FBX) {
	f, err := os.Open(modelName)
	check(err)
	defer f.Close()

	reader := NewReaderWithFilters(
		MatchStackAndSubNodes("Objects/Geometry", "Vertices", "PolygonVertexIndex"),
		jobs,
		// EITHER(
		// 	FilterName("Objects/Geometry/Vertices"),
		// 	FilterName("Objects/Geometry/PolygonVertexIndex"),
		// ),
	)
	reader.ReadFrom(f)
	check(reader.Error)
	fbx <- reader.FBX
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

func markPoly(current byte, new byte) byte {
	if current == 0 {
		return new
	}
	if current != new {
		return 3
	}
	return new
}

func WrapToIndex(i int32) int32 {
	return i ^ -1 // i*-1 - 1
}

// SplitByPlane accumulates all geometry nodes and splits them by some plane
func SplitByPlane(geomNode *Node, clippingPlane Plane) ([]Diff, []Diff) {

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

	// marked 1 if it's retained, 2 if it's clipped, 3 if it's in both.
	// eventually those marked 3 will disapear as I have to create new polys
	// for a proper split by plane.
	vertMarks := make([]byte, len(vertice))
	vertexPolyIndexMarks := make([]byte, len(verticeIndexes))

	numFaces := len(verticeIndexes) / 3

	// Mark which tris belong in retained or clipped
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

		if pos == 3 {
			vertexPolyIndexMarks[faceIndex] = markPoly(vertexPolyIndexMarks[faceIndex], 1)
			vertexPolyIndexMarks[faceIndex+1] = markPoly(vertexPolyIndexMarks[faceIndex+1], 1)
			vertexPolyIndexMarks[faceIndex+2] = markPoly(vertexPolyIndexMarks[faceIndex+2], 1)

			vertMarks[firstInd] = markPoly(vertMarks[firstInd], 1)
			vertMarks[firstInd+1] = markPoly(vertMarks[firstInd+1], 1)
			vertMarks[firstInd+2] = markPoly(vertMarks[firstInd+2], 1)
			vertMarks[secondInd] = markPoly(vertMarks[secondInd], 1)
			vertMarks[secondInd+1] = markPoly(vertMarks[secondInd+1], 1)
			vertMarks[secondInd+2] = markPoly(vertMarks[secondInd+2], 1)
			vertMarks[wrapInd] = markPoly(vertMarks[wrapInd], 1)
			vertMarks[wrapInd+1] = markPoly(vertMarks[wrapInd+1], 1)
			vertMarks[wrapInd+2] = markPoly(vertMarks[wrapInd+2], 1)
		}
		if neg == 3 {
			vertexPolyIndexMarks[faceIndex] = markPoly(vertexPolyIndexMarks[faceIndex], 2)
			vertexPolyIndexMarks[faceIndex+1] = markPoly(vertexPolyIndexMarks[faceIndex+1], 2)
			vertexPolyIndexMarks[faceIndex+2] = markPoly(vertexPolyIndexMarks[faceIndex+2], 2)

			vertMarks[firstInd] = markPoly(vertMarks[firstInd], 2)
			vertMarks[firstInd+1] = markPoly(vertMarks[firstInd+1], 2)
			vertMarks[firstInd+2] = markPoly(vertMarks[firstInd+2], 2)
			vertMarks[secondInd] = markPoly(vertMarks[secondInd], 2)
			vertMarks[secondInd+1] = markPoly(vertMarks[secondInd+1], 2)
			vertMarks[secondInd+2] = markPoly(vertMarks[secondInd+2], 2)
			vertMarks[wrapInd] = markPoly(vertMarks[wrapInd], 2)
			vertMarks[wrapInd+1] = markPoly(vertMarks[wrapInd+1], 2)
			vertMarks[wrapInd+2] = markPoly(vertMarks[wrapInd+2], 2)
		}
	}

	clippedVertexes := make([]float64, 0)
	clippedVertexOffsets := make([]int32, len(vertice)/3)
	curClippedOffset := 0

	retainedVertexes := make([]float64, 0)
	retainedVertexOffsets := make([]int32, len(vertice)/3)
	curRetainedOffset := 0

	numPoints := len(vertMarks) / 3

	for p := 0; p < numPoints; p++ {
		clippedVertexOffsets[p] = int32(curClippedOffset)
		retainedVertexOffsets[p] = int32(curRetainedOffset)

		startingVertIndex := p * 3
		mark := vertMarks[startingVertIndex]
		if mark == 0 {
			curClippedOffset++
			curRetainedOffset++
		}

		if mark == 1 {
			curClippedOffset++
			retainedVertexes = append(retainedVertexes, vertice[startingVertIndex], vertice[startingVertIndex+1], vertice[startingVertIndex+2])
		}

		if mark == 2 {
			curRetainedOffset++
			clippedVertexes = append(clippedVertexes, vertice[startingVertIndex], vertice[startingVertIndex+1], vertice[startingVertIndex+2])
		}

		if mark == 3 {
			retainedVertexes = append(retainedVertexes, vertice[startingVertIndex], vertice[startingVertIndex+1], vertice[startingVertIndex+2])
			clippedVertexes = append(clippedVertexes, vertice[startingVertIndex], vertice[startingVertIndex+1], vertice[startingVertIndex+2])
		}

	}

	clippedPolyVertexIndices := make([]int32, 0)
	retainedPolyVertexIndices := make([]int32, 0)

	for f := 0; f < numFaces; f++ {
		faceIndex := f * 3
		mark := vertexPolyIndexMarks[faceIndex]

		if mark == 1 || mark == 3 {
			offsetOne := retainedVertexOffsets[verticeIndexes[faceIndex]]
			offsetTwo := retainedVertexOffsets[verticeIndexes[faceIndex+1]]
			offsetThree := retainedVertexOffsets[WrapToIndex(verticeIndexes[faceIndex+2])]
			retainedPolyVertexIndices = append(retainedPolyVertexIndices, verticeIndexes[faceIndex]-offsetOne, verticeIndexes[faceIndex+1]-offsetTwo, verticeIndexes[faceIndex+2]+offsetThree)
			continue
		}

		if mark == 2 || mark == 3 {
			offsetOne := clippedVertexOffsets[verticeIndexes[faceIndex]]
			offsetTwo := clippedVertexOffsets[verticeIndexes[faceIndex+1]]
			offsetThree := clippedVertexOffsets[WrapToIndex(verticeIndexes[faceIndex+2])]
			clippedPolyVertexIndices = append(clippedPolyVertexIndices, verticeIndexes[faceIndex]-offsetOne, verticeIndexes[faceIndex+1]-offsetTwo, verticeIndexes[faceIndex+2]+offsetThree)
		}
	}

	// log.Printf("Retained: %d", len(retainedPolyVertexIndices)/3)
	// log.Printf("clipped: %d", len(clippedPolyVertexIndices)/3)

	return []Diff{
		// NewArrayPropertyDiff(vertexNodes[0].id, NewArrayPropertyFloat64CompressedSlice(retainedVertexes)),
		// NewArrayPropertyDiff(polyVertexNodes[0].id, NewArrayPropertyInt32CompressedSlice(retainedPolyVertexIndices)),
		},
		[]Diff{
			NewArrayPropertyDiff(vertexNodes[0].id, NewArrayPropertyFloat64CompressedSlice(clippedVertexes)),
			NewArrayPropertyDiff(polyVertexNodes[0].id, NewArrayPropertyInt32CompressedSlice(clippedPolyVertexIndices)),
		}
}

func insertNewDiff(existingDiffs []Diff, newDiff Diff) []Diff {
	result := append(existingDiffs, newDiff)

	for i := len(result) - 1; i > 0; i-- {
		next := i - 1
		if result[i].NodeID() < result[next].NodeID() {
			result[next], result[i] = result[i], result[next]
		} else {
			break
		}
	}

	return result
}

func combineSorted(sortedArrays ...[]Diff) []Diff {
	resultLen := 0
	for _, a := range sortedArrays {
		resultLen += len(a)
	}

	result := make([]Diff, resultLen)
	sortedArrayIndexes := make([]int, len(sortedArrays))

	for resultIndex := 0; resultIndex < resultLen; resultIndex++ {

		currentLowestArray := -1
		lowestValue := uint64(math.MaxInt64)

		for sortedArrayIndex, sortedArray := range sortedArrays {
			if sortedArrayIndexes[sortedArrayIndex] < len(sortedArray) {
				if currentLowestArray == -1 || sortedArray[sortedArrayIndexes[sortedArrayIndex]].NodeID() < uint64(lowestValue) {
					currentLowestArray = sortedArrayIndex
					lowestValue = sortedArray[sortedArrayIndexes[sortedArrayIndex]].NodeID()
				}
			}
		}

		result[resultIndex] = sortedArrays[currentLowestArray][sortedArrayIndexes[currentLowestArray]]
		sortedArrayIndexes[currentLowestArray]++
	}

	return result
}

func worker(id int, plane Plane, jobs <-chan []*Node, results chan<- WorkerResult) {
	allRetainedPolygons := make([]Diff, 0)
	allClippedPolygons := make([]Diff, 0)

	for j := range jobs {
		for _, n := range j {
			retained, clipped := SplitByPlane(n, plane)
			allRetainedPolygons = append(allRetainedPolygons, retained...)
			// for _, d := range retained {
			// 	allRetainedPolygons = append(allRetainedPolygons, d)
			// }
			allClippedPolygons = append(allClippedPolygons, clipped...)
			// for _, d := range clipped {
			// 	allClippedPolygons = append(allClippedPolygons, d)
			// }
		}
	}

	sort.Sort(SortDiff(allClippedPolygons))
	sort.Sort(SortDiff(allRetainedPolygons))
	results <- WorkerResult{clipped: allClippedPolygons, retained: allRetainedPolygons}
}

// SplitByPlaneProgram loads in a FBX model and splits it
func SplitByPlaneProgram(
	modelName string,
	plane Plane,
	workers int,
	retained io.Writer,
	clipped io.Writer,
) *FBX {
	timer.begin(fmt.Sprintf("Loading and splitting %s by plane with %d workers", modelName, workers))

	jobs := make(chan []*Node, 10000)
	workerOutput := make(chan WorkerResult, 10000)
	finalFBX := make(chan *FBX)

	// start workers before attempting to load model
	for w := 0; w < workers; w++ {
		go worker(w, plane, jobs, workerOutput)
	}

	go loadModel(modelName, jobs, finalFBX)

	workerResultsRetained := make([][]Diff, workers)
	workerResultsClipped := make([][]Diff, workers)

	for i := 0; i < workers; i++ {
		r := <-workerOutput
		workerResultsRetained[i] = r.retained
		workerResultsClipped[i] = r.clipped
	}
	allRetainedPolygons := combineSorted(workerResultsRetained...)
	allClippedPolygons := combineSorted(workerResultsClipped...)

	fbx := <-finalFBX
	timer.end()

	timer.begin(fmt.Sprintf("Writing results"))
	defer timer.end()

	retainedWriterError := make(chan error)
	clippedWriterError := make(chan error)
	retainedWriter := NewPatchWriter(fbx, allRetainedPolygons, func(n int, e error) {
		retainedWriterError <- e
		close(retainedWriterError)
	})
	clippedWriter := NewPatchWriter(fbx, allClippedPolygons, func(n int, e error) {
		clippedWriterError <- e
		close(clippedWriterError)
	})

	go retainedWriter.Write(retained)
	go clippedWriter.Write(clipped)

	retErr := <-retainedWriterError
	if retErr != nil {
		log.Printf("Error writing to retained: %s", retErr.Error())
	}

	clipErr := <-clippedWriterError
	if clipErr != nil {
		log.Printf("Error writing to clipped: %s", clipErr.Error())
	}

	return fbx
}

func main() {

	retainedOut, err := os.Create("o-retained.fbx")
	check(err)
	defer retainedOut.Close()

	clippedOut, err := os.Create("o-clipped.fbx")
	check(err)
	defer clippedOut.Close()

	// out, err := os.Create("out.txt")
	// check(err)
	// defer out.Close()

	// log.Printf("Retained Model Polygon Count: %d", len(retained.GetFaces()))
	// log.Printf("Clipped Model Polygon Count: %d", len(clipped.GetFaces()))

	// SplitByPlaneProgram("dragon_vrip.fbx", NewPlane(vector.Vector3Zero(), vector.Vector3Forward()), 3, retainedOut, clippedOut)
	// SplitByPlaneProgram("spheres.fbx", NewPlane(vector.Vector3Zero(), vector.Vector3Forward()), 3, retainedOut, clippedOut)
	// expand(out, fbx.Top)
	// for _, c := range fbx.Nodes {
	// 	expand(out, c)
	// }

	// f, err := os.Open("retained.fbx")
	// check(err)
	// defer f.Close()

	// reader := NewReader()
	// reader.ReadFrom(f)
	// check(reader.Error)
	// expand(out, reader.FBX.Top)
	// for _, c := range reader.FBX.Nodes {
	// 	expand(out, c)
	// }

	// save(retained, "retained.obj")
	// save(clipped, "clipped.obj")

	SplitByPlaneProgram("HIB-model.fbx", NewPlane(vector.NewVector3(105.4350, 119.4877, 77.9060), vector.Vector3Up()), 3, retainedOut, clippedOut)
	// log.Printf("Retained Model Polygon Count: %d", len(retained.GetFaces()))
	// log.Printf("Clipped Model Polygon Count: %d", len(clipped.GetFaces()))
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
		return "S: " + s
	}
	if string(p.TypeCode) == "I" {
		return "I: " + fmt.Sprint(p.AsInt32())
	}

	if string(p.TypeCode) == "D" {
		return "D: " + fmt.Sprint(p.AsFloat64())
	}

	if string(p.TypeCode) == "L" {
		return "L: " + fmt.Sprint(p.AsInt64())
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
	fmt.Fprintf(out, "-> [%d] %s\n", node.id, node.Name)

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
