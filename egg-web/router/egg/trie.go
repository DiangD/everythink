package egg

import "strings"

type node struct {
	pattern string
	part    string
	child   []*node
	isWild  bool
}

func (n *node) insert(pattern string, parts []string, height int) {
	if height == len(parts) {
		n.pattern = pattern
		return
	}

	part := parts[height]
	child := n.matchChild(part)
	if child == nil {
		child = &node{
			part:   part,
			isWild: part[0] == ':' || part[0] == '*',
		}
		n.child = append(n.child, child)
	}
	child.insert(pattern, parts, height+1)
}

func (n *node) search(parts []string, height int) *node {
	if height == len(parts) || strings.HasSuffix(n.part, "*") {
		if n.pattern == "" {
			return nil
		}
		return n
	}

	part := parts[height]
	child := n.matchAllChild(part)
	for _, node := range child {
		res := node.search(parts, height+1)
		if res != nil {
			return res
		}
	}
	return nil
}

func (n *node) matchChild(part string) *node {
	for _, node := range n.child {
		if node.part == part || node.isWild {
			return node
		}
	}
	return nil
}

func (n *node) matchAllChild(part string) []*node {
	nodes := make([]*node, 0)
	for _, node := range n.child {
		if node.part == part || node.isWild {
			nodes = append(nodes, node)
		}
	}
	return nodes
}
