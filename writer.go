package main

import (
	"io"
	"time"
)

// Writer is responsible for writing nodes to FBX
type Writer struct {
	w             io.Writer
	currentOffset uint64
	err           error
	complete      bool
}

// NewWriter creates a new writer and immediately writes the FBX header and
// top node
func NewWriter(w io.Writer) (Writer, error) {

	// Write header
	n, err := w.Write([]byte{
		75, 97, 121, 100,
		97, 114, 97, 32,
		70, 66, 88, 32,
		66, 105, 110, 97,
		114, 121, 32, 32,
		0, 26, 0, 76, 29,
		0, 0,
	})

	fbxWriter := Writer{
		w:             w,
		currentOffset: uint64(n),
		err:           nil,
		complete:      false,
	}

	if err != nil {
		return fbxWriter, err
	}

	creationTime := time.Now()

	fbxWriter.WriteNode(NewNodeParent(
		"FBXHeaderExtension",
		NewNodeInt32("FBXHeaderVersion", 1003),
		NewNodeInt32("FBXVersion", 7500),
		NewNodeInt32("EncryptionType", 0),
		CreateTimestampNode(creationTime),
		NewNodeString("Creator", "https://github.com/EliCDavis"),
	))

	fbxWriter.WriteNode(NewNodeString("CreationTime", creationTime.String()))
	fbxWriter.WriteNode(NewNodeString("Creator", "https://github.com/EliCDavis"))

	return fbxWriter, nil
}

// WriteNode writes a node to the writer, and returns true if you can continue writing
func (w *Writer) WriteNode(n *Node) bool {
	if w.err != nil || w.complete {
		return false
	}

	newOffset, err := n.Write(w.w, w.currentOffset)
	if err != nil {
		w.err = err
		return false
	}
	w.currentOffset = newOffset
	return true
}

func (w *Writer) Complete() error {
	if w.err != nil || w.complete {
		return w.err
	}
	n, err := w.w.Write([]byte{
		0, 0, 0, 0, 0,
		0, 0, 0, 0, 0,
		0, 0, 0, 0, 0,
		0, 0, 0, 0, 0,
		0, 0, 0, 0, 0,
	})

	w.currentOffset += uint64(n)
	w.err = err
	return w.err
}
