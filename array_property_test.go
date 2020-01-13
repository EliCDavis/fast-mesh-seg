package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSaveAndLoadInt32SliceProperty(t *testing.T) {
	// ****************************** ARRANGE *********************************
	data := []int32{
		666,
		420,
		69,
		2020,
	}

	reader := FBXReader{}
	buffer := new(bytes.Buffer)
	prop := NewArrayPropertyInt32Slice(data)

	// ******************************** ACT ***********************************
	writeErr := prop.Write(buffer)
	propType, typeErr := buffer.ReadByte()
	propFromBuffer := reader.readArray(buffer, 4)

	// ******************************* ASSERT *********************************
	assert.NoError(t, writeErr)
	assert.NoError(t, typeErr)
	assert.NoError(t, reader.Error)
	assert.Equal(t, byte('i'), propType)
	if assert.NotNil(t, propFromBuffer) {
		dataBack := propFromBuffer.AsInt32Slice()
		if assert.Len(t, dataBack, len(data)) {
			assert.Equal(t, data[0], dataBack[0])
			assert.Equal(t, data[1], dataBack[1])
			assert.Equal(t, data[2], dataBack[2])
			assert.Equal(t, data[3], dataBack[3])
		}
	}
}

func TestSaveAndLoadFloat64SliceProperty(t *testing.T) {
	// ****************************** ARRANGE *********************************
	data := []float64{
		666,
		420,
		69,
		2020,
	}

	reader := FBXReader{}
	buffer := new(bytes.Buffer)
	prop := NewArrayPropertyFloat64Slice(data)

	// ******************************** ACT ***********************************
	writeErr := prop.Write(buffer)
	propType, typeErr := buffer.ReadByte()
	propFromBuffer := reader.readArray(buffer, 8)

	// ******************************* ASSERT *********************************
	assert.NoError(t, writeErr)
	assert.NoError(t, typeErr)
	assert.NoError(t, reader.Error)
	assert.Equal(t, byte('d'), propType)
	if assert.NotNil(t, propFromBuffer) {
		dataBack := propFromBuffer.AsFloat64Slice()
		if assert.Len(t, dataBack, len(data)) {
			assert.Equal(t, data[0], dataBack[0])
			assert.Equal(t, data[1], dataBack[1])
			assert.Equal(t, data[2], dataBack[2])
			assert.Equal(t, data[3], dataBack[3])
		}
	}
}
