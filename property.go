package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Property contains the byte data and type of property to a specific node
type Property struct {
	TypeCode byte
	Data     []byte
}

// AsString interprets the byte data as a string
func (p *Property) AsString() string {
	return string(p.Data)
}

// AsBytes just returns the raw bytes found in the property
func (p *Property) AsBytes() []byte {
	return p.Data
}

// AsInt8 interprets the first byte of our data as an 8bit integer
func (p *Property) AsInt8() int8 {
	var data int8
	buf := bytes.NewReader(p.Data)
	/*err :=*/ binary.Read(buf, binary.LittleEndian, &data)
	return data
}

// AsInt16 interprets the data buffer as a 16bit integer
func (p *Property) AsInt16() int16 {
	var data int16
	buf := bytes.NewReader(p.Data)
	/*err :=*/ binary.Read(buf, binary.LittleEndian, &data)
	return data
}

// AsInt32 interprets the data buffer as a 32bit integer
func (p *Property) AsInt32() int32 {
	var data int32
	buf := bytes.NewReader(p.Data)
	/*err :=*/ binary.Read(buf, binary.LittleEndian, &data)
	return data
}

// AsInt64 interprets the data buffer as a 64bit integer
func (p *Property) AsInt64() int64 {
	var data int64
	buf := bytes.NewReader(p.Data)
	/*err :=*/ binary.Read(buf, binary.LittleEndian, &data)
	return data
}

// AsFloat32 interprets the data buffer as a 32bit float
func (p *Property) AsFloat32() float32 {
	var data float32
	buf := bytes.NewReader(p.Data)
	/*err :=*/ binary.Read(buf, binary.LittleEndian, &data)
	return data
}

// AsFloat64 interprets the data buffer as a 64bit float
func (p *Property) AsFloat64() float64 {
	var data float64
	buf := bytes.NewReader(p.Data)
	/*err :=*/ binary.Read(buf, binary.LittleEndian, &data)
	return data
}

// AsBool interprets the first byte of data in the buffer as a boolean
func (p *Property) AsBool() bool {
	return p.Data[0] != 0
}

func (p *Property) String() string {
	return fmt.Sprintf("%v", p.Data)
}
