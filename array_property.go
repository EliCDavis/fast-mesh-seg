package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
)

// ArrayProperty contains the byte data and type of property to a specific
// node. Data thought to represent an array
type ArrayProperty struct {
	TypeCode         byte
	Data             []byte
	ArrayLength      uint32
	Encoding         uint32
	CompressedLength uint32
}

// AsFloat32Slice attempts to parse the buffer as an array of 32bit floats
func (p ArrayProperty) AsFloat32Slice() []float32 {
	data := make([]float32, p.ArrayLength)
	if p.Encoding == 0 {
		buf := bytes.NewReader(p.Data)
		/*err :=*/ binary.Read(buf, binary.LittleEndian, &data)
	} else {
		/*err :=*/ p.uncompress(data)
	}
	return data
}

// AsFloat64Slice attempts to parse the buffer as an array of 64bit floats
func (p ArrayProperty) AsFloat64Slice() []float64 {
	data := make([]float64, p.ArrayLength)
	if p.Encoding == 0 {
		buf := bytes.NewReader(p.Data)
		/*err :=*/ binary.Read(buf, binary.LittleEndian, &data)
	} else {
		/*err :=*/ p.uncompress(data)
	}
	return data
}

// AsInt32Slice attempts to parse the buffer as an array of 32bit ints
func (p ArrayProperty) AsInt32Slice() []int32 {
	data := make([]int32, p.ArrayLength)
	if p.Encoding == 0 {
		buf := bytes.NewReader(p.Data)
		/*err :=*/ binary.Read(buf, binary.LittleEndian, &data)
	} else {
		/*err :=*/ p.uncompress(data)
	}
	return data
}

// AsInt64Slice attempts to parse the buffer as an array of 64bit ints
func (p ArrayProperty) AsInt64Slice() []int64 {
	data := make([]int64, p.ArrayLength)
	if p.Encoding == 0 {
		buf := bytes.NewReader(p.Data)
		/*err :=*/ binary.Read(buf, binary.LittleEndian, &data)
	} else {
		/*err :=*/ p.uncompress(data)
	}
	return data
}

func (p ArrayProperty) uncompress(data interface{}) error {
	buf := bytes.NewBuffer(p.Data)
	r, err := zlib.NewReader(buf)
	if err != nil {
		return err
	}
	defer r.Close()
	err = binary.Read(r, binary.LittleEndian, data)
	return err
}
