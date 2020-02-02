package main

import (
	"encoding/binary"
	"io"
)

// PatchWriter takes an fbx and a list of patches and writes out the results
// the underlying writer
type PatchWriter struct {
	fbx       *FBX
	diffs     []Diff
	diffIndex int
	callback  func(int, error)
}

func NewPatchWriter(fbx *FBX, diffs []Diff, callback func(int, error)) *PatchWriter {
	return &PatchWriter{
		fbx:       fbx,
		diffs:     diffs,
		callback:  callback,
		diffIndex: 0,
	}
}

func (pw PatchWriter) Write(w io.Writer) (int, error) {
	currentOffset := 0
	bytesWritten, err := w.Write(pw.fbx.Header.data)
	currentOffset += bytesWritten

	if err != nil {
		return currentOffset, err
	}

	currentOffset, err = pw.writeNode(w, pw.fbx.Top, currentOffset, false)
	if err != nil {
		return currentOffset, err
	}

	for i, n := range pw.fbx.Nodes {
		currentOffset, err = pw.writeNode(w, n, currentOffset, len(pw.fbx.Nodes)-1 == i)
		if err != nil {
			return currentOffset, err
		}
	}

	err = binary.Write(w, binary.LittleEndian, uint64(7409768082772311290))
	if err != nil {
		return currentOffset, err
	}
	err = binary.Write(w, binary.LittleEndian, uint64(9090224599739365041))
	if err != nil {
		return currentOffset, err
	}

	fuckingEmpty := make([]byte, 120)
	n, err := w.Write(fuckingEmpty)

	if err != nil {
		return currentOffset + 16 + n, err
	}

	// this just appears at the end of every compliant file
	n, err = w.Write([]byte{0xF8, 0x5A, 0x8C, 0x6A, 0xDE, 0xF5, 0xD9, 0x7E, 0xEC, 0xE9, 0x0C, 0xE3, 0x75, 0x8F, 0x29, 0x0B})

	pw.callback(currentOffset+136+n, err)
	return currentOffset + 136 + n, err
}

// WriteNode writes a node to the writer, and returns true if you can continue writing
func (pw *PatchWriter) writeNode(w io.Writer, n *Node, currentOffset int, endOfList bool) (int, error) {
	var diffedNode *Node
	diffedNode, pw.diffIndex = n.ApplyDiffs(pw.diffs, pw.diffIndex)
	newOffset, err := diffedNode.Write(w, uint64(currentOffset), endOfList)
	// newOffset, err := n.Write(w, uint64(currentOffset), endOfList)
	return int(newOffset), err
}
