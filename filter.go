package main

import "strings"

// NodeFilter returns true if a node should be kept
type NodeFilter func(*NodeStack) bool

// FilterName returns true if the node's name matches what's passed in
func FilterName(name string) NodeFilter {
	return func(n *NodeStack) bool {
		s := n.String()
		if len(s) < len(name) {
			return strings.Index(name, s) == 0
		}
		return strings.Index(s, name) == 0
	}
}

// EITHER filter returns true f any of the filters passed in return true
func EITHER(filters ...NodeFilter) NodeFilter {
	return func(n *NodeStack) bool {
		for _, f := range filters {
			if f(n) == true {
				return true
			}
		}
		return false
	}
}

func MatchStackAndSubNodes(name string, subNodes ...string) NodeFilter {
	return func(n *NodeStack) bool {
		if name != n.String() {
			return false
		}
		for _, subName := range subNodes {
			match := false
			for _, subNode := range n.data[n.position].NestedNodes {
				if subNode.Name == subName {
					match = true
				}
			}
			if match == false {
				return false
			}
		}

		return true
	}
}
