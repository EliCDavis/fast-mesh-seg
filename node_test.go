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
	reader.nodeHeaderSize = 25
	reader.nodeHeader = make([]byte, 25)
	reader.FBX.Header = &Header{
		data:    nil,
		version: 7500,
	}
	buffer := new(bytes.Buffer)
	node := NewNodeFloat64Slice("Float64 Test", data)

	// ******************************** ACT ***********************************
	_, writeErr := node.Write(buffer, 0)
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
