package main

import (
	"io"
)

// PatchWriter takes an fbx and a list of patches and writes out the results
// the underlying writer
type PatchWriter struct {
	fbx            *FBX
	diffs          []Diff
	remainingDiffs []Diff
}

func NewPatchWriter(fbx *FBX, diffs []Diff) *PatchWriter {
	return &PatchWriter{
		fbx:   fbx,
		diffs: diffs,
	}
}

func (pw PatchWriter) Write(w io.Writer) (n int, err error) {
	currentOffset := 0
	bytesWritten, err := w.Write(pw.fbx.Header.data)
	currentOffset += bytesWritten

	if err != nil {
		return currentOffset, err
	}

	pw.remainingDiffs = make([]Diff, len(pw.diffs))
	copy(pw.remainingDiffs, pw.diffs)

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

	return currentOffset + n, err
}

// WriteNode writes a node to the writer, and returns true if you can continue writing
func (pw *PatchWriter) writeNode(w io.Writer, n *Node, currentOffset int, endOfList bool) (int, error) {
	// var diffedNode *Node
	// diffedNode, pw.remainingDiffs = n.ApplyDiffs(pw.remainingDiffs)
	// newOffset, err := diffedNode.Write(w, uint64(currentOffset), endOfList)
	newOffset, err := n.Write(w, uint64(currentOffset), endOfList)
	return int(newOffset), err
}
