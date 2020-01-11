package main

import "time"

// CreateTimestampNode creates a node with multiple children with properties
// that put together
func CreateTimestampNode(t time.Time) *Node {
	return NewNodeParent(
		"CreationTimeStamp",
		[]*Node{
			NewNodeSingleProperty("Version", NewPropertyInt32(1000)),
			NewNodeSingleProperty("Year", NewPropertyInt32(int32(t.Year()))),
			NewNodeSingleProperty("Month", NewPropertyInt32(int32(t.Month()))),
			NewNodeSingleProperty("Day", NewPropertyInt32(int32(t.Day()))),
			NewNodeSingleProperty("Hour", NewPropertyInt32(int32(t.Hour()))),
			NewNodeSingleProperty("Minute", NewPropertyInt32(int32(t.Minute()))),
			NewNodeSingleProperty("Second", NewPropertyInt32(int32(t.Second()))),
			NewNodeSingleProperty("Millisecond", NewPropertyInt32(0)),
		},
	)
}
