package main

import (
	"testing"

	"github.com/EliCDavis/mesh"
	"github.com/EliCDavis/vector"
)

var resRetained, resClipped mesh.Model

func BenchmarkSplitByPlane(b *testing.B) {
	// fbx := loadModel("dragon_vrip.fbx")
	// plane := NewPlane(vector.Vector3Zero(), vector.Vector3Forward())

	fbx := loadModel("HIB-model.fbx")
	plane := NewPlane(vector.NewVector3(105.4350, 119.4877, 77.9060), vector.Vector3Up())

	geomNodes := fbx.GetNodes("Objects", "Geometry")
	var retained, clipped mesh.Model

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		// always record the result of func to prevent
		// the compiler eliminating the function call.
		retained, clipped = SplitByPlane(geomNodes, plane)
	}
	// always store the result to a package level variable
	// so the compiler cannot eliminate the Benchmark itself.
	resRetained = retained
	resClipped = clipped
}
