package main

//https://github.com/o5h/fbx/blob/master/reader.go

import (
	"encoding/binary"
	"fmt"
	"io"
)

// FBXReader builds an FBX file from a reader
type FBXReader struct {
	FBX                      *FBX
	Position                 int64
	Error                    error
	Filters                  []NodeFilter
	stack                    *NodeStack
	results                  chan<- []*Node
	matcher                  NodeFilter
	currentResultsBuffer     []*Node
	currentResultsBufferSize int64
	nodeHeader               []byte
	curNodeCount             uint64
}

// NewReader creates a new reader
func NewReader() *FBXReader {
	return &FBXReader{
		FBX:      &FBX{},
		Position: 0,
		Error:    nil,
		Filters:  nil,
		stack:    NewNodeStack(),
		results:  nil,
		matcher:  nil,
	}
}

func NewReaderWithFilters(matcher NodeFilter, results chan<- []*Node, filters ...NodeFilter) *FBXReader {
	return &FBXReader{
		FBX:      &FBX{},
		Position: 0,
		Error:    nil,
		Filters:  filters,
		stack:    NewNodeStack(),
		results:  results,
		matcher:  matcher,
	}
}

func (fr FBXReader) filter() bool {
	for _, filter := range fr.Filters {
		if filter(fr.stack) == false {
			return false
		}
	}
	return true
}

func (fr *FBXReader) ReadFrom(r io.ReadSeeker) (n int64, err error) {

	fr.FBX.Header = fr.ReadHeaderFrom(r)
	if fr.Error != nil {
		return
	}

	nodeHeaderSize := 25 // 8 + 8 + 8 + 1
	if fr.FBX.Header.Version() < 7500 {
		nodeHeaderSize = 13 // 4 + 4 + 4 + 1
	}
	fr.nodeHeader = make([]byte, nodeHeaderSize)

	fr.FBX.Top, _ = fr.ReadNodeFrom(r)
	if fr.Error != nil {
		return
	}

	for {
		node, empty := fr.ReadNodeFrom(r)
		if fr.Error != nil {
			break
		}
		fr.FBX.Nodes = append(fr.FBX.Nodes, node)
		if empty {
			break
		}
	}

	// if fr.currentResultsBufferSize > 0 {
	// 	fr.results <- fr.currentResultsBuffer
	// }

	if fr.results != nil {
		close(fr.results)
	}

	return
}

func (fr *FBXReader) ReadHeaderFrom(r io.Reader) *Header {
	return NewHeader(fr.read(r, 27))
}

// ReadNodeFrom builds a node from the reader and returns true if the node was empty
func (fr *FBXReader) ReadNodeFrom(r io.ReadSeeker) (*Node, bool) {
	node := &Node{}
	node.id = fr.curNodeCount
	node.endingID = node.id
	fr.curNodeCount++
	fr.stack.push(node)
	defer fr.stack.pop()

	fr.readInto(r, fr.nodeHeader)
	if fr.Error != nil {
		return nil, true
	}

	var endOffset uint64
	if fr.FBX.Header.Version() >= 7500 {
		endOffset = uint64(fr.nodeHeader[0]) | uint64(fr.nodeHeader[1])<<8 | uint64(fr.nodeHeader[2])<<16 |
			uint64(fr.nodeHeader[3])<<24 | uint64(fr.nodeHeader[4])<<32 | uint64(fr.nodeHeader[5])<<40 |
			uint64(fr.nodeHeader[6])<<48 | uint64(fr.nodeHeader[7])<<56

		node.NumProperties = uint64(fr.nodeHeader[8]) | uint64(fr.nodeHeader[9])<<8 | uint64(fr.nodeHeader[10])<<16 |
			uint64(fr.nodeHeader[11])<<24 | uint64(fr.nodeHeader[12])<<32 | uint64(fr.nodeHeader[13])<<40 |
			uint64(fr.nodeHeader[14])<<48 | uint64(fr.nodeHeader[15])<<56

		node.PropertyListLen = uint64(fr.nodeHeader[16]) | uint64(fr.nodeHeader[17])<<8 | uint64(fr.nodeHeader[18])<<16 |
			uint64(fr.nodeHeader[19])<<24 | uint64(fr.nodeHeader[20])<<32 | uint64(fr.nodeHeader[21])<<40 |
			uint64(fr.nodeHeader[22])<<48 | uint64(fr.nodeHeader[23])<<56

		node.NameLen = fr.nodeHeader[24]
	} else {
		endOffset = uint64(uint32(fr.nodeHeader[0]) | uint32(fr.nodeHeader[1])<<8 | uint32(fr.nodeHeader[2])<<16 | uint32(fr.nodeHeader[3])<<24)
		node.NumProperties = uint64(uint32(fr.nodeHeader[4]) | uint32(fr.nodeHeader[5])<<8 | uint32(fr.nodeHeader[6])<<16 | uint32(fr.nodeHeader[7])<<24)
		node.PropertyListLen = uint64(uint32(fr.nodeHeader[8]) | uint32(fr.nodeHeader[9])<<8 | uint32(fr.nodeHeader[10])<<16 | uint32(fr.nodeHeader[11])<<24)
		node.NameLen = fr.nodeHeader[12]
	}

	size := endOffset - uint64(fr.Position)

	if fr.FBX.Header.Version() >= 7500 {
		node.Length = size + 25
	} else {
		node.Length = size + 13
	}

	if node.NameLen > 0 {
		bb := fr.read(r, int(node.NameLen))
		if fr.Error != nil {
			return nil, false
		}
		node.Name = string(bb)
	}

	if endOffset == 0 {
		node.Length = 0
		return node, true
	}

	if fr.filter() == false {
		leftToRead := endOffset - uint64(fr.Position)

		r.Seek(int64(leftToRead), io.SeekCurrent)

		fr.Position += int64(leftToRead)
		return node, false
	}

	for np := uint64(0); np < node.NumProperties; np++ {
		fr.ReadPropertyFrom(r, node)
		if fr.Error != nil {
			return node, false
		}
	}

	for {
		if fr.Position >= int64(endOffset) {
			break
		}

		subNode, empty := fr.ReadNodeFrom(r)
		if fr.Error != nil {
			break
		}
		node.endingID = subNode.endingID

		node.NestedNodes = append(node.NestedNodes, subNode)
		if empty {
			break
		}
	}

	if fr.matcher != nil && fr.results != nil {
		if fr.matcher(fr.stack) {
			fr.addNodeToResultsChannel(node, int64(size))
		}
	}

	return node, false
}

func (fr *FBXReader) addNodeToResultsChannel(n *Node, size int64) {
	fr.currentResultsBufferSize += size
	fr.currentResultsBuffer = append(fr.currentResultsBuffer, n)

	if fr.currentResultsBufferSize > 1000000 {
		fr.results <- fr.currentResultsBuffer
		fr.currentResultsBufferSize = 0
		fr.currentResultsBuffer = make([]*Node, 0)
	}

}

func (fr *FBXReader) ReadPropertyFrom(r io.Reader, node *Node) {
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
		arrayProp = fr.readArray(r, 4)
		arrayProp.TypeCode = typeCode
	case 'd':
		arrayProp = fr.readArray(r, 8)
		arrayProp.TypeCode = typeCode
	case 'i':
		arrayProp = fr.readArray(r, 4)
		arrayProp.TypeCode = typeCode
	case 'l':
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

	var bufferLength int
	if a.Encoding == 0 {
		bufferLength = int(eleSize * a.ArrayLength)
	} else {
		bufferLength = int(a.CompressedLength)
	}
	a.Data = fr.read(r, int(bufferLength))

	return a
}

func (fr *FBXReader) readUint64(r io.Reader) uint64 {
	b := fr.read(r, 8)
	return binary.LittleEndian.Uint64(b)
}

func (fr *FBXReader) readUint32(r io.Reader) uint32 {
	b := fr.read(r, 4)
	return binary.LittleEndian.Uint32(b)
}

func (fr *FBXReader) readUint8(r io.Reader) uint8 {
	return fr.read(r, 1)[0]
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

func (fr *FBXReader) readInto(r io.Reader, b []byte) {
	var i int
	i, fr.Error = r.Read(b)
	fr.Position += int64(i)
}

var stringReadReuse = []byte{0, 0, 0, 0}

func (fr *FBXReader) readString(r io.Reader) []byte {
	// len := fr.readUint32(r)
	// if fr.Error != nil {
	// 	return nil
	// }

	// b := fr.read(r, 4)
	// return binary.LittleEndian.Uint32(b)

	fr.readInto(r, stringReadReuse)
	return fr.read(r, int(binary.LittleEndian.Uint32(stringReadReuse)))
}

func (fr *FBXReader) readBytes(r io.Reader) []byte {
	len := fr.readUint32(r)
	if fr.Error != nil {
		return nil
	}
	return fr.read(r, int(len))
}

func ReadFrom(r io.ReadSeeker) (*FBX, error) {
	reader := NewReader()
	reader.ReadFrom(r)
	return reader.FBX, reader.Error
}
