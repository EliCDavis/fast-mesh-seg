package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// Property contains the byte data and type of property to a specific node
type Property struct {
	TypeCode byte
	Data     []byte
}

// NewPropertyInt32 creates a property that holds data for an Int32
func NewPropertyInt32(p int32) *Property {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, p)
	return &Property{
		TypeCode: 'I',
		Data:     buf.Bytes(),
	}
}

// NewPropertyString creates a property that holds data for a string
func NewPropertyString(s string) *Property {
	return &Property{
		TypeCode: 'S',
		Data:     []byte(s),
	}
}

// Size returns how much space the property would take up in an FBX
func (p *Property) Size() uint64 {
	size := uint64(len(p.Data)) + 1 // +1 comes from the typecode byte

	// Strings and byte arrays have 4 bytes that represent the size of the
	// string/[]byte
	if p.TypeCode == 'S' || p.TypeCode == 'R' {
		size += 4
	}
	return size
}

func (p Property) Write(w io.Writer) error {
	_, err := w.Write([]byte{p.TypeCode})
	if err != nil {
		return err
	}

	if p.TypeCode == 'S' || p.TypeCode == 'R' {
		err = binary.Write(w, binary.LittleEndian, uint32(len(p.Data)))
		if err != nil {
			return err
		}
	}

	_, err = w.Write(p.Data)
	if err != nil {
		return err
	}
	return nil
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
