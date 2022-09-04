package main

import (
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	tree "github.com/mariusor/bubbles-tree"
)

type pathNode struct {
	parent *pathNode
	Path   string
	state  tree.NodeState
	Nodes  []*pathNode
}

func (n pathNode) GoString() string {
	s := strings.Builder{}
	nodeS := fmt.Sprintf("Path: %s [%d]", n.Path, len(n.Nodes))
	s.WriteString(nodeS)
	if len(n.Children()) > 0 {
		s.WriteString(fmt.Sprintf("%#v", n.Children()))
	}
	s.WriteString("\n")
	return s.String()
}

func (n pathNode) Parent() tree.Node {
	return n.parent
}

func (n pathNode) String() string {
	return n.Path
}

func (n pathNode) Children() tree.Nodes {
	return treeNodes(n.Nodes)
}

func (n pathNode) State() tree.NodeState {
	return n.state
}

func (n *pathNode) SetState(s tree.NodeState) {
	n.state = s
}

func buildNodeTree(root fs.FS) (tree.Nodes, error) {
	allNodes := make([]*pathNode, 0)
	fs.WalkDir(root, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return fs.SkipDir
		}
		log.Printf("path: %s", p)
		cnt := len(strings.Split(p, "/"))
		if cnt > 2 {
			return fs.SkipDir
		}
		st := tree.NodeVisible
		if d.IsDir() {
			st |= tree.NodeCollapsible
		}
		parent := findNodeByPath(allNodes, filepath.Dir(p))
		node := pathNode{
			parent: parent,
			Path:   p,
			state:  st,
			Nodes:  make([]*pathNode, 0),
		}
		if parent == nil {
			allNodes = append(allNodes, &node)
		} else {
			parent.Nodes = append(parent.Nodes, &node)
		}
		return nil
	})

	return treeNodes(allNodes), nil
}

func treeNodes(pathNodes []*pathNode) tree.Nodes {
	nodes := make(tree.Nodes, len(pathNodes))
	for i, n := range pathNodes {
		nodes[i] = n
	}
	return nodes
}

func findNodeByPath(nodes []*pathNode, path string) *pathNode {
	for _, node := range nodes {
		if filepath.Clean(node.String()) == filepath.Clean(path) {
			return node
		}
		if child := findNodeByPath(node.Nodes, path); child != nil {
			return child
		}
	}
	return nil
}
