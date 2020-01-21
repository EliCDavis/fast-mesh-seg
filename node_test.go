package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSaveAndLoadNodeWithSliceProperty(t *testing.T) {
	// ****************************** ARRANGE *********************************
	data := []float64{
		666,
		420,
		69,
		2020,
	}

	reader := NewReader()
	reader.nodeHeader = make([]byte, 25)
	reader.FBX.Header = &Header{
		data:    nil,
		version: 7500,
	}
	buffer := new(bytes.Buffer)
	node := NewNodeFloat64Slice("Float64 Test", data)

	// ******************************** ACT ***********************************
	_, writeErr := node.Write(buffer, 0, false)
	nodeFromBuffer, _ := reader.ReadNodeFrom(bytes.NewReader(buffer.Bytes()))

	// ******************************* ASSERT *********************************
	assert.NoError(t, writeErr)
	assert.NoError(t, reader.Error)
	if assert.NotNil(t, nodeFromBuffer) {
		dataBack, _ := nodeFromBuffer.Float64Slice()
		if assert.Len(t, dataBack, len(data)) {
			assert.Equal(t, data[0], dataBack[0])
			assert.Equal(t, data[1], dataBack[1])
			assert.Equal(t, data[2], dataBack[2])
			assert.Equal(t, data[3], dataBack[3])
		}
	}
}

func TestSaveAndLoadNestedNodeStructuresWithSliceProperty(t *testing.T) {
	// ****************************** ARRANGE *********************************
	data := []float64{
		6.66,
		42.0,
		6.9,
		20.20,
	}
	data32 := []int32{
		666,
		420,
		69,
		2020,
	}

	reader := NewReader()
	buffer := new(bytes.Buffer)
	writer, err := NewWriter(buffer)
	if assert.NoError(t, err) == false {
		return
	}
	node := NewNodeParent(
		"Geometry",
		NewNodeParent(
			"The Parent Node",
			NewNodeString("cock and", "ball torture"),
			NewNodeFloat64Slice("Float64 Test", data),
			NewNodeInt32Slice("Int32 Test", data32),
		),
	)

	// ******************************** ACT ***********************************
	writeSuccesful := writer.WriteNode(node)
	completeSuccesful := writer.Complete()
	_, readErr := reader.ReadFrom(bytes.NewReader(buffer.Bytes()))
	nodeFromBuffer := reader.FBX.Nodes[2]

	// ******************************* ASSERT *********************************
	assert.True(t, writeSuccesful)
	assert.NoError(t, completeSuccesful)
	assert.NoError(t, readErr)
	assert.NoError(t, reader.Error)
	if assert.NotNil(t, nodeFromBuffer) == false {
		return
	}

	if assert.Len(t, nodeFromBuffer.NestedNodes, 1) == false {
		return
	}

	if assert.Len(t, nodeFromBuffer.NestedNodes[0].NestedNodes, 3) {
		assert.Equal(t, "cock and", nodeFromBuffer.NestedNodes[0].NestedNodes[0].Name)
		sData, _ := nodeFromBuffer.NestedNodes[0].NestedNodes[0].StringProperty()
		assert.Equal(t, "ball torture", sData)

		data64Back, _ := nodeFromBuffer.NestedNodes[0].NestedNodes[1].Float64Slice()
		assert.Equal(t, "Float64 Test", nodeFromBuffer.NestedNodes[0].NestedNodes[1].Name)
		if assert.Len(t, data64Back, len(data)) {
			assert.Equal(t, data[0], data64Back[0])
			assert.Equal(t, data[1], data64Back[1])
			assert.Equal(t, data[2], data64Back[2])
			assert.Equal(t, data[3], data64Back[3])
		}

		data32Back, _ := nodeFromBuffer.NestedNodes[0].NestedNodes[2].Int32Slice()
		assert.Equal(t, "Int32 Test", nodeFromBuffer.NestedNodes[0].NestedNodes[2].Name)
		if assert.Len(t, data32Back, len(data)) {
			assert.Equal(t, data32[0], data32Back[0])
			assert.Equal(t, data32[1], data32Back[1])
			assert.Equal(t, data32[2], data32Back[2])
			assert.Equal(t, data32[3], data32Back[3])
		}
	}
}
