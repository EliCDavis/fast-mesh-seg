package main

// https://github.com/o5h/fbx/blob/master/header.go

import (
	"encoding/binary"
	"fmt"
)

// type Header [27]byte

// Header is the header to the fbx file
type Header struct {
	data    []byte
	version uint32
}

// NewHeader creates a new Header and computes file version
func NewHeader(data []byte) *Header {
	return &Header{
		data:    data,
		version: binary.LittleEndian.Uint32(data[23:27]),
	}
}

func (h Header) String() string {
	return fmt.Sprint(string(h.data[0:20]), h.Version())
}

// Version represents the FBX file version
func (h Header) Version() uint32 {
	return h.version
}
