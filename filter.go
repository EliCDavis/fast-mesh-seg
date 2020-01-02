package main

import "strings"

type NodeFilter func(*NodeStack) bool

func FilterName(name string) NodeFilter {
	return func(n *NodeStack) bool {
		s := n.String()
		if len(s) < len(name) {
			return strings.Index(name, s) == 0
		}
		return strings.Index(s, name) == 0
	}
}
