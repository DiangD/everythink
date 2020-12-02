package egg

import "strings"

type node struct {
	pattern string
	part    string
	child   []*node
	isWild  bool //是否为精确匹配 :golang || *filepath 为true
}

//插入节点
func (n *node) insert(pattern string, parts []string, height int) {
	//递归终点
	if height == len(parts) {
		//只有在最后一层才设置pattern 如：/hello/:name/search 当part=search时才设置pattern
		n.pattern = pattern
		return
	}

	part := parts[height]
	//匹配孩子
	child := n.matchChild(part)
	//第一次插入时为nil
	if child == nil {
		//生成孩子
		child = &node{
			part:   part,
			isWild: part[0] == ':' || part[0] == '*',
		}
		n.child = append(n.child, child)
	}
	//递归
	child.insert(pattern, parts, height+1)
}

//查找节点
func (n *node) search(parts []string, height int) *node {
	//递归出口
	if height == len(parts) || strings.HasSuffix(n.part, "*") {
		//匹配失败
		if n.pattern == "" {
			return nil
		}
		return n
	}

	part := parts[height]
	//匹配所有孩子
	child := n.matchAllChild(part)
	//循环遍历所有孩子节点
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
