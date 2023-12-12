package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	tree "github.com/mariusor/bubbles-tree"
)

const RootPath = "/tmp"

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

func (n *pathNode) Init() tea.Cmd {
	return nil
}

const (
	Collapsed = "⊞"
	Expanded  = "⊟"
)

func (n *pathNode) View() string {
	name := filepath.Base(n.path)
	if n.parent == nil {
		name = n.path
	}

	hints := n.state
	annotation := ""
	s := strings.Builder{}
	if hints&tree.NodeCollapsible == tree.NodeCollapsible {
		annotation = Expanded
		if hints&tree.NodeCollapsed == tree.NodeCollapsed {
			annotation = Collapsed
		}
	}
	if len(annotation) > 0 {
		fmt.Fprintf(&s, "%-2s%s", annotation, name)
	} else {
		fmt.Fprintf(&s, "%s", name)
	}

	return s.String()
}

func (n *pathNode) Children() tree.Nodes {
	return treeNodes(n.children)
}

func (n *pathNode) State() tree.NodeState {
	return n.state
}

func (n *pathNode) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tree.NodeState:
		n.state = m
	case tree.Nodes:
		n.setChildren(m...)
	}

	return n, nil
}

func (n *pathNode) setChildren(nodes ...tree.Node) {
	n.children = n.children[:0]
	for _, nn := range nodes {
		if c, ok := nn.(*pathNode); ok {
			n.children = append(n.children, c)
		}
	}
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
	_ = fs.WalkDir(os.DirFS(root), ".", func(p string, d fs.DirEntry, err error) error {
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

		st := tree.NodeNone
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

type quittingTree struct {
	tree.Model
}

func (e *quittingTree) Update(m tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := m.(tea.KeyMsg); ok && key.Matches(msg, key.NewBinding(key.WithKeys("q"))) {
		return e, tea.Quit
	}
	mod, cmd := e.Model.Update(m)
	if mm, ok := mod.(*tree.Model); ok {
		e.Model = *mm
	}
	return e, cmd
}

func main() {
	var depth int
	var style string
	flag.IntVar(&depth, "depth", 2, "The maximum depth to read the directory structure")
	flag.StringVar(&style, "style", "normal", "The style to use when drawing the tree: double, thick, rounded, edge, normal")
	flag.Parse()

	symbols := tree.DefaultSymbols()
	switch style {
	case "thick":
		symbols = tree.ThickSymbols()
	case "rounded":
		symbols = tree.RoundedSymbols()
	case "double":
		symbols = tree.DoubleSymbols()
	case "edge":
		symbols = tree.NormalEdgeSymbols()
	case "thickedge":
		symbols = tree.ThickEdgeSymbols()
	case "", "normal":
	default:
		fmt.Fprintf(os.Stderr, "invalid style type, using default 'normal'\n")
	}

	path := RootPath
	if flag.NArg() > 0 {
		abs, err := filepath.Abs(flag.Arg(0))
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			os.Exit(1)
		}
		path = abs
	}

	t := tree.New(buildNodeTree(path, depth))
	t.Symbols = symbols
	m := quittingTree{Model: t}

	if _, err := tea.NewProgram(&m).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}
