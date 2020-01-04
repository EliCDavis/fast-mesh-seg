package main

//https://github.com/o5h/fbx/blob/master/reader.go

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
)

// FBXReader builds an FBX file from a reader
type FBXReader struct {
	FBX      *FBX
	Position int64
	Error    error
	Filters  []NodeFilter
	stack    *NodeStack
}

// NewReader creates a new reader
func NewReader() *FBXReader {
	return &FBXReader{&FBX{}, 0, nil, nil, NewNodeStack()}
}

func NewReaderWithFilters(filters ...NodeFilter) *FBXReader {
	return &FBXReader{&FBX{}, 0, nil, filters, NewNodeStack()}
}

func (fr FBXReader) filter() bool {
	for _, filter := range fr.Filters {
		if filter(fr.stack) == false {
			return false
		}
	}
	return true
}

func (fr *FBXReader) ReadFrom(r io.Reader) (n int64, err error) {
	fr.FBX.Header = fr.ReadHeaderFrom(r)
	if err != nil {
		return
	}

	fr.FBX.Top = fr.ReadNodeFrom(r, true)
	if fr.Error != nil {
		return
	}

	for {
		node := fr.ReadNodeFrom(r, false)
		if fr.Error != nil {
			break
		}
		if node.IsEmpty() {
			break
		}
		fr.FBX.Nodes = append(fr.FBX.Nodes, node)
	}

	return
}

func (fr *FBXReader) ReadHeaderFrom(r io.Reader) (header *Header) {
	header = &Header{}
	var i int
	i, fr.Error = r.Read(header[:])
	fr.Position += int64(i)
	return
}

func (fr *FBXReader) ReadEndOffset(r io.Reader) uint64 {
	if fr.FBX.Header.Version() >= 7500 {
		return fr.readUint64(r)
	}
	return uint64(fr.readUint32(r))
}

func (fr *FBXReader) ReadNumProperties(r io.Reader) uint64 {
	if fr.FBX.Header.Version() >= 7500 {
		return fr.readUint64(r)
	}
	return uint64(fr.readUint32(r))
}

func (fr *FBXReader) ReadPropertyListLen(r io.Reader) uint64 {
	if fr.FBX.Header.Version() >= 7500 {
		return fr.readUint64(r)
	}
	return uint64(fr.readUint32(r))
}

func (fr *FBXReader) ReadNodeFrom(r io.Reader, top bool) (node *Node) {
	node = &Node{}
	fr.stack.push(node)
	defer fr.stack.pop()

	node.EndOffset = fr.ReadEndOffset(r)
	if fr.Error != nil {
		return
	}

	node.NumProperties = fr.ReadNumProperties(r)
	if fr.Error != nil {
		return
	}

	node.PropertyListLen = fr.ReadPropertyListLen(r)
	if fr.Error != nil {
		return
	}

	node.NameLen = fr.readUint8(r)
	if fr.Error != nil {
		return
	}

	if node.NameLen > 0 {
		bb := make([]byte, node.NameLen)
		var i int
		i, fr.Error = io.ReadFull(r, bb)
		if fr.Error != nil {
			return
		}
		node.Name = string(bb)
		fr.Position += int64(i)
	}

	if node.EndOffset == 0 {
		return
	}

	if fr.filter() == false {
		leftToRead := node.EndOffset - uint64(fr.Position)

		switch rCasted := r.(type) {
		case io.Seeker:
			rCasted.Seek(int64(leftToRead), io.SeekCurrent)
		default:
			io.CopyN(ioutil.Discard, r, int64(leftToRead))
		}

		fr.Position += int64(leftToRead)
		return
	}

	for np := uint64(0); np < node.NumProperties; np++ {
		fr.ReadPropertyFrom(r, node)
		if fr.Error != nil {
			return
		}
	}

	for {
		if fr.Position >= int64(node.EndOffset) {
			break
		}

		subNode := fr.ReadNodeFrom(r, false)
		if fr.Error != nil {
			break
		}

		if subNode.IsEmpty() {
			break
		}
		node.NestedNodes = append(node.NestedNodes, subNode)
	}

	return node
}

func (fr *FBXReader) ReadPropertyFrom(r io.Reader, node *Node) {
	var nn int64
	var prop *Property
	var arrayProp *ArrayProperty
	typeCode := fr.readUint8(r)

	switch typeCode {
	case 'S':
		prop = &Property{
			TypeCode: typeCode,
			Data:     fr.readString(r),
		}
	case 'R':
		prop = &Property{
			TypeCode: typeCode,
			Data:     fr.readBytes(r),
		}
	case 'Y':
		prop = &Property{
			TypeCode: typeCode,
			Data:     fr.readInt16(r),
		}
	case 'C':
		prop = &Property{
			TypeCode: typeCode,
			Data:     fr.readInt8(r),
		}
	case 'I':
		prop = &Property{
			TypeCode: typeCode,
			Data:     fr.readInt32(r),
		}
	case 'F':
		prop = &Property{
			TypeCode: typeCode,
			Data:     fr.readFloat32(r),
		}
	case 'D':
		prop = &Property{
			TypeCode: typeCode,
			Data:     fr.readFloat64(r),
		}
	case 'L':
		prop = &Property{
			TypeCode: typeCode,
			Data:     fr.readInt64(r),
		}
	case 'f':
		// p.Data = fr.readArray(r, 4,
		// 	func(len uint32) interface{} {
		// 		data := make([]float32, len)
		// 		return data
		// 	})
		arrayProp = fr.readArray(r, 4)
		arrayProp.TypeCode = typeCode
	case 'd':
		// p.Data = fr.readArray(r, 8,
		// 	func(len uint32) interface{} {
		// 		data := make([]float64, len)
		// 		return data
		// 	})
		arrayProp = fr.readArray(r, 8)
		arrayProp.TypeCode = typeCode
	case 'i':
		// p.Data = fr.readArray(r, 4,
		// 	func(len uint32) interface{} {
		// 		data := make([]int32, len)
		// 		return data
		// 	})
		arrayProp = fr.readArray(r, 4)
		arrayProp.TypeCode = typeCode
	case 'l':
		// p.Data = fr.readArray(r, 8,
		// 	func(len uint32) interface{} {
		// 		data := make([]int64, len)
		// 		return data
		// 	})
		arrayProp = fr.readArray(r, 8)
		arrayProp.TypeCode = typeCode
	case 'b':
		// var tmp []byte
		// array := fr.readArray(r, 1,
		// 	func(len uint32) interface{} {
		// 		tmp = make([]byte, len)
		// 		return tmp
		// 	})
		// data := make([]bool, len(tmp))
		// for i, b := range tmp {
		// 	data[i] = (b == 1)
		// }
		// array.Data = tmp
		// p.Data = array
		arrayProp = fr.readArray(r, 1)
		arrayProp.TypeCode = typeCode
	default:
		panic(fmt.Sprintf("unsupported type '%s'", string(typeCode)))
	}

	if prop != nil {
		node.Properties = append(node.Properties, prop)
	}

	if arrayProp != nil {
		node.ArrayProperties = append(node.ArrayProperties, arrayProp)
	}

	fr.Position += nn
	return
}

func (fr *FBXReader) readArrayHeader(r io.Reader, a *ArrayProperty) {
	a.ArrayLength = fr.readUint32(r)
	if fr.Error != nil {
		return
	}
	a.Encoding = fr.readUint32(r)
	if fr.Error != nil {
		return
	}

	a.CompressedLength = fr.readUint32(r)
	if fr.Error != nil {
		return
	}
	return
}

func (fr *FBXReader) readArray(r io.Reader, eleSize uint32) *ArrayProperty {
	a := &ArrayProperty{}
	fr.readArrayHeader(r, a)
	if fr.Error != nil {
		return nil
	}
	// data := slicer(a.ArrayLength)
	// var nn int64
	// nn, fr.Error = readArrayData(r, eleSize, a, data)
	// a.Data = data
	// fr.Position += nn

	var bufferLength int
	if a.Encoding == 0 {
		bufferLength = int(eleSize * a.ArrayLength)
	} else {
		bufferLength = int(a.CompressedLength)
	}
	a.Data = fr.read(r, int(bufferLength))

	return a
}

// func readArrayData(r io.Reader, size uint32, a *ArrayProperty, data interface{}) (n int64, err error) {
// 	if a.Encoding == 0 {
// 		err = binary.Read(r, binary.LittleEndian, data)
// 		if err != nil {
// 			return
// 		}
// 		n += int64(size * a.ArrayLength)
// 	} else {
// 		var compressedBytes = make([]byte, a.CompressedLength)
// 		err = binary.Read(r, binary.LittleEndian, &compressedBytes)
// 		if err != nil {
// 			return
// 		}
// 		n += int64(a.CompressedLength)
// 		err = uncompress(compressedBytes, data)
// 	}
// 	return
// }

func (fr *FBXReader) readUint64(r io.Reader) uint64 {
	var data uint64
	fr.Error = binary.Read(r, binary.LittleEndian, &data)
	fr.Position += 8
	return data
}

func (fr *FBXReader) readUint32(r io.Reader) uint32 {
	var data uint32
	fr.Error = binary.Read(r, binary.LittleEndian, &data)
	fr.Position += 4
	return data
}

func (fr *FBXReader) readUint8(r io.Reader) uint8 {
	var data uint8
	fr.Error = binary.Read(r, binary.LittleEndian, &data)
	fr.Position++
	return data
}

func (fr *FBXReader) readInt16(r io.Reader) []byte {
	return fr.read(r, 2)
}

func (fr *FBXReader) readInt32(r io.Reader) []byte {
	return fr.read(r, 4)
}

func (fr *FBXReader) readFloat32(r io.Reader) []byte {
	return fr.read(r, 4)
}

func (fr *FBXReader) readFloat64(r io.Reader) []byte {
	return fr.read(r, 8)
}

func (fr *FBXReader) readInt64(r io.Reader) []byte {
	return fr.read(r, 8)
}

func (fr *FBXReader) readInt8(r io.Reader) []byte {
	return fr.read(r, 1)
}

func (fr *FBXReader) read(r io.Reader, bytes int) []byte {
	b := make([]byte, bytes)
	var i int
	i, fr.Error = r.Read(b)
	fr.Position += int64(i)
	return b
}

func (fr *FBXReader) readString(r io.Reader) []byte {
	len := fr.readUint32(r)
	if fr.Error != nil {
		return nil
	}
	return fr.read(r, int(len))
}

func (fr *FBXReader) readBytes(r io.Reader) []byte {
	len := fr.readUint32(r)
	if fr.Error != nil {
		return nil
	}
	return fr.read(r, int(len))
}

func ReadFrom(r io.Reader) (*FBX, error) {
	reader := NewReader()
	reader.ReadFrom(r)
	return reader.FBX, reader.Error
}
