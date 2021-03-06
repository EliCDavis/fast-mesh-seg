package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"io"
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

func NewArrayPropertyInt32Slice(p []int32) *ArrayProperty {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, p)
	return &ArrayProperty{
		TypeCode:         'i',
		Data:             buf.Bytes(),
		ArrayLength:      uint32(len(p)),
		Encoding:         0,
		CompressedLength: 0,
	}
}

func NewArrayPropertyInt32CompressedSlice(p []int32) *ArrayProperty {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, p)
	if err != nil {
		panic(err)
	}

	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(buf.Bytes())
	w.Close()
	compressedBytes := b.Bytes()

	return &ArrayProperty{
		TypeCode:         'i',
		Data:             compressedBytes,
		ArrayLength:      uint32(len(p)),
		Encoding:         1,
		CompressedLength: uint32(len(compressedBytes)),
	}
}

func NewArrayPropertyFloat64Slice(p []float64) *ArrayProperty {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, p)
	return &ArrayProperty{
		TypeCode:         'd',
		Data:             buf.Bytes(),
		ArrayLength:      uint32(len(p)),
		Encoding:         0,
		CompressedLength: 0,
	}
}

func NewArrayPropertyFloat64CompressedSlice(p []float64) *ArrayProperty {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, p)
	if err != nil {
		panic(err)
	}

	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(buf.Bytes())
	w.Close()
	compressedBytes := b.Bytes()

	return &ArrayProperty{
		TypeCode:         'd',
		Data:             compressedBytes,
		ArrayLength:      uint32(len(p)),
		Encoding:         1,
		CompressedLength: uint32(len(compressedBytes)),
	}
}

// Size returns how much space the property would take up in an FBX
func (p *ArrayProperty) Size() uint64 {
	// 13 =   4 (array length)
	//      + 4 (encoding)
	//      + 4 (compressed length)
	//      + 1  typecode byte
	return uint64(len(p.Data)) + 13
}

func (p ArrayProperty) Write(w io.Writer) error {
	_, err := w.Write([]byte{p.TypeCode})
	if err != nil {
		return err
	}

	err = binary.Write(w, binary.LittleEndian, p.ArrayLength)
	if err != nil {
		return err
	}

	err = binary.Write(w, binary.LittleEndian, p.Encoding)
	if err != nil {
		return err
	}

	err = binary.Write(w, binary.LittleEndian, p.CompressedLength)
	if err != nil {
		return err
	}

	_, err = w.Write(p.Data)

	return err
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

		// for i := range data {
		// 		data[i] = order.Uint32(bs[4*i:])
		// 	}

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
