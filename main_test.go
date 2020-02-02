package main

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/EliCDavis/vector"
)

// var resRetained, resClipped mesh.Model

func BenchmarkSplitByPlane(b *testing.B) {
	// fbx := loadModel("dragon_vrip.fbx")
	// plane := NewPlane(vector.Vector3Zero(), vector.Vector3Forward())

	// fbx := loadModel("HIB-model.fbx")
	// plane := NewPlane(vector.NewVector3(105.4350, 119.4877, 77.9060), vector.Vector3Up())

	// b.ResetTimer()
	for n := 0; n < b.N; n++ {
		// always record the result of func to prevent
		// the compiler eliminating the function call.
		SplitByPlaneProgram("dragon_vrip.fbx", NewPlane(vector.Vector3Zero(), vector.Vector3Forward()), 3, ioutil.Discard, ioutil.Discard)
	}

}

func TestInsertionSort(t *testing.T) {
	// ****************************** ARRANGE *********************************
	diffs := []Diff{
		NewArrayPropertyDiff(1, nil),
		NewArrayPropertyDiff(2, nil),
		NewArrayPropertyDiff(4, nil),
		NewArrayPropertyDiff(5, nil),
	}

	newDif := NewArrayPropertyDiff(3, nil)

	// ******************************** ACT ***********************************
	results := insertNewDiff(diffs, newDif)

	// ******************************* ASSERT *********************************

	// Nothings changed
	assert.Len(t, diffs, 4)
	assert.Equal(t, uint64(1), diffs[0].NodeID())
	assert.Equal(t, uint64(2), diffs[1].NodeID())
	assert.Equal(t, uint64(4), diffs[2].NodeID())
	assert.Equal(t, uint64(5), diffs[3].NodeID())

	// Results
	assert.Len(t, results, 5)
	assert.Equal(t, uint64(1), results[0].NodeID())
	assert.Equal(t, uint64(2), results[1].NodeID())
	assert.Equal(t, uint64(3), results[2].NodeID())
	assert.Equal(t, uint64(4), results[3].NodeID())
	assert.Equal(t, uint64(5), results[4].NodeID())
}

func TestInsertionSortOnEmptyArray(t *testing.T) {
	// ****************************** ARRANGE *********************************
	var diffs []Diff

	newDif := NewArrayPropertyDiff(3, nil)

	// ******************************** ACT ***********************************
	results := insertNewDiff(diffs, newDif)

	// ******************************* ASSERT *********************************
	assert.Len(t, results, 1)
	assert.Equal(t, uint64(3), results[0].NodeID())
}

func TestCombiningSortedArrays(t *testing.T) {
	// ****************************** ARRANGE *********************************
	diffs1 := []Diff{
		NewArrayPropertyDiff(1, nil),
		NewArrayPropertyDiff(2, nil),
		NewArrayPropertyDiff(4, nil),
		NewArrayPropertyDiff(5, nil),
	}

	diffs2 := []Diff{
		NewArrayPropertyDiff(0, nil),
		NewArrayPropertyDiff(2, nil),
		NewArrayPropertyDiff(6, nil),
		NewArrayPropertyDiff(7, nil),
	}

	var diffs3 []Diff

	diffs4 := []Diff{
		NewArrayPropertyDiff(1, nil),
		NewArrayPropertyDiff(7, nil),
		NewArrayPropertyDiff(8, nil),
		NewArrayPropertyDiff(9, nil),
	}

	// ******************************** ACT ***********************************
	results := combineSorted(diffs1, diffs2, diffs3, diffs4)

	// ******************************* ASSERT *********************************
	if assert.Len(t, results, 12) {
		assert.Equal(t, uint64(0), results[0].NodeID())
		assert.Equal(t, uint64(1), results[1].NodeID())
		assert.Equal(t, uint64(1), results[2].NodeID())
		assert.Equal(t, uint64(2), results[3].NodeID())
		assert.Equal(t, uint64(2), results[4].NodeID())
		assert.Equal(t, uint64(4), results[5].NodeID())
		assert.Equal(t, uint64(5), results[6].NodeID())
		assert.Equal(t, uint64(6), results[7].NodeID())
		assert.Equal(t, uint64(7), results[8].NodeID())
		assert.Equal(t, uint64(7), results[9].NodeID())
		assert.Equal(t, uint64(8), results[10].NodeID())
		assert.Equal(t, uint64(9), results[11].NodeID())
	}
}
