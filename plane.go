package main

import (
	"github.com/EliCDavis/vector"
)

// Plane has an origin and a normal
type Plane struct {
	origin vector.Vector3
	normal vector.Vector3
}

// NewPlane creates a new plane
func NewPlane(origin, normal vector.Vector3) Plane {
	return Plane{origin, normal.Normalized()}
}
