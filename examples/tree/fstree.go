package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	tree "github.com/mariusor/bubbles-tree"
)

type pathNode struct {
	parent   *pathNode
	path     string
	state    tree.NodeState
	children []*pathNode
}

func (n *pathNode) Parent() tree.Node {
	if n == nil || n.parent == nil {
		return nil
	}
	return n.parent
}

func (n *pathNode) Name() string {
	if n.parent == nil {
		return n.path
	}
	return filepath.Base(n.path)
}

func (n *pathNode) Children() tree.Nodes {
	return treeNodes(n.children)
}

func (n *pathNode) State() tree.NodeState {
	return n.state
}

func (n *pathNode) SetState(s tree.NodeState) {
	n.state = s
}

func isUnixHiddenFile(name string) bool {
	return len(name) > 2 && (name[0] == '.' || name[:2] == "..")
}

func buildNodeTree(root string, maxDepth int) tree.Nodes {
	allNodes := make([]*pathNode, 0)

	rootPath := func(p string) string {
		if p == "." {
			return root
		}
		return p
	}
	fs.WalkDir(os.DirFS(root), ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return fs.SkipDir
		}
		if isUnixHiddenFile(d.Name()) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		cnt := len(strings.Split(p, string(os.PathSeparator)))
		if maxDepth != -1 && cnt > maxDepth {
			return fs.SkipDir
		}

		st := tree.NodeVisible
		if d.IsDir() {
			st |= tree.NodeCollapsible
		}
		p = rootPath(p)
		parent := findNodeByPath(allNodes, rootPath(filepath.Dir(p)))

		node := new(pathNode)
		node.path = p
		node.state = st
		node.children = make([]*pathNode, 0)

		if parent == nil {
			allNodes = append(allNodes, node)
		} else {
			node.parent = parent
			node.state |= tree.NodeCollapsed
			parent.children = append(parent.children, node)
		}
		return nil
	})

	return treeNodes(allNodes)
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
		if filepath.Clean(node.path) == filepath.Clean(path) {
			return node
		}
		if child := findNodeByPath(node.children, path); child != nil {
			return child
		}
	}
	return nil
}
