package gi

import (
	"fmt"
	"strings"
)

type Node struct {
	level    int
	part     string
	pattern  string
	isWild   bool
	children []*Node
}

func parseParts(str string) []string {
	parts := strings.Split(str, "/")
	res := make([]string, 0)
	for _, part := range parts {
		if part != "" {
			res = append(res, part)
			// *xxx 特殊通配符处理，只取 *(xx) 后得 xx 作为 ParamKey
			// /*/xx 正常插入节点
			if len(part) > 1 && part[0] == '*' {
				break
			}
		}
	}
	return res
}

func (n *Node) matchChild(part string) *Node {
	for _, node := range n.children {
		if part == node.part {
			return node
		}
	}

	return nil
}

func (n *Node) matchChidren(part string) []*Node {
	var nodes []*Node
	for _, node := range n.children {
		nodePart := node.part
		if part == nodePart || nodePart[0] == '*' || nodePart[0] == ':' {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func (n *Node) findChild(parts []string, level int) *Node {
	if level == len(parts) {
		if n.pattern != "" {
			return n
		}
		return nil
	}

	part := parts[level]
	matchChildren := n.matchChidren(part)
	for _, c := range matchChildren {
		res := c.findChild(parts, level+1)
		if res != nil {
			return res
		}
	}

	return nil
}

func (n *Node) insertChild(pattern string, parts []string, level int) {
	part := parts[level]
	child := n.matchChild(part)
	if child == nil {
		//将通配符标识继承给子节点 （方便查找）
		isWild := n.isWild || strings.HasPrefix(part, "*") || strings.HasPrefix(part, ":")
		child = &Node{
			level:  level,
			part:   parts[level],
			isWild: isWild,
		}
		n.children = append(n.children, child)
	}

	if level+1 == len(parts) {
		if child.pattern != "" {
			fmt.Printf("插入节点出现覆盖:\n[%v] <=== [%v]\n", n.pattern, pattern)
		}

		child.pattern = pattern
	} else {
		child.insertChild(pattern, parts, level+1)
	}
}

// 插入节点返回 pattern 对应的 *node
func (n *Node) Insert(pattern string) {
	parts := parseParts(pattern)
	n.insertChild(pattern, parts, 0)
}

func (n *Node) Search(path string) *Node {
	parts := parseParts(path)
	return n.findChild(parts, 0)
}
