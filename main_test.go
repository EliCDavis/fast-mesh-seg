package main

// var resRetained, resClipped mesh.Model

// func BenchmarkSplitByPlane(b *testing.B) {
// 	// fbx := loadModel("dragon_vrip.fbx")
// 	// plane := NewPlane(vector.Vector3Zero(), vector.Vector3Forward())

// 	// fbx := loadModel("HIB-model.fbx")
// 	plane := NewPlane(vector.NewVector3(105.4350, 119.4877, 77.9060), vector.Vector3Up())

// 	var retained, clipped mesh.Model

// 	b.ResetTimer()
// 	for n := 0; n < b.N; n++ {
// 		// always record the result of func to prevent
// 		// the compiler eliminating the function call.
// 		_, retained, clipped = SplitByPlaneProgram("HIB-model.fbx", plane, 3)
// 	}
// 	// always store the result to a package level variable
// 	// so the compiler cannot eliminate the Benchmark itself.
// 	resRetained = retained
// 	resClipped = clipped
// }

// func TestSplitByPlane(t *testing.T) {

// 	plane := NewPlane(vector.NewVector3(105.4350, 119.4877, 77.9060), vector.Vector3Up())

// 	SplitByPlaneProgram("dragon_vrip.fbx", plane, 3)
// }
