package tree

import (
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// NodeState is used for passing information from a Treeish element to the view itself
type NodeState int

//func init() {
//	fmt.Fprintf(os.Stderr, "Collapsed %d\n", NodeCollapsed)
//	fmt.Fprintf(os.Stderr, "Collapsible %d\n", NodeCollapsible)
//	fmt.Fprintf(os.Stderr, "Visible %d\n", NodeVisible)
//}

const (
	// NodeCollapsed hints that the current node is collapsed
	NodeCollapsed = 1 << iota
	// NodeCollapsible hints that the current node can be collapsed
	NodeCollapsible
	// NodeVisible hints that the current node is ready to be displayed
	NodeVisible
	NodeError
	NodeDebug

	NodeNone = 0
)

type Treeish interface {
	// Advance moves the Treeish to a new received path,
	// this can return a new Treeish instance at the new path, or perform some other function
	// for the cases where the path doesn't correspond to a Treeish object.
	// Specifically in the case of the filepath Treeish:
	// If a passed path parameter corresponds to a folder, it will return a new Treeish object at the new path
	// If the passed path parameter corresponds to a file, it returns a nil Treeish but can execute something else.
	Advance(string) (Treeish, error)
	// State returns the NodeState for the received path parameter
	// This is used when rendering the path in the tree view
	State(string) (NodeState, error)

	// Walk loads the elements of current Treeish and returns them as a flat list
	Walk(int) ([]string, error)
}

type debugNode struct {
	Content interface{}
	state   NodeState
}

func debug(m interface{}) debugNode {
	return debugNode{Content: m}
}

func (m *Model) debug(s string, params ...interface{}) {
	if m.debugNodes == nil {
		m.debugNodes = make([]Node, 0)
	}
	m.debugNodes = append(m.debugNodes, debug(fmt.Sprintf(s, params...)))
}

func (m *Model) err(err error) {
	if m.debugNodes == nil {
		m.debugNodes = make([]Node, 0)
	}
	m.debugNodes = append(m.debugNodes, debug(err))
}

func (d debugNode) String() string {
	switch n := d.Content.(type) {
	case error:
		return n.Error()
	case string:
		return n
	}
	return fmt.Sprintf("unknown type %T", d)
}

func (d debugNode) Children() Nodes {
	return nil
}

func (d debugNode) State() NodeState {
	if _, ok := d.Content.(error); ok {
		return NodeError
	}
	return NodeDebug
}

type pathNode struct {
	Path  string
	state NodeState
	Nodes Nodes
}

func (n pathNode) GoString() string {
	s := strings.Builder{}
	nodeS := fmt.Sprintf("Path: %s [%d]\n", n.Path, len(n.Children()))
	s.WriteString(nodeS)
	if len(n.Children()) > 0 {
		s.WriteString(fmt.Sprintf("%#v\n", n.Children()))
	}
	return s.String()
}

func (n pathNode) String() string {
	return n.Path
}

func (n pathNode) Children() Nodes {
	return n.Nodes
}

func (n pathNode) State() NodeState {
	return n.state
}

type Node interface {
	String() string
	Children() Nodes
	State() NodeState
}

type viewport struct {
	// h, w hold the screen dimensions in lines and columns
	h, w int
	// top holds the index of the rendered lines that is displayed at the top of the screen
	top, left int
	// pos should represend the index of the rendered line that the cursor is on
	pos   int
	lines []string
}

func (v viewport) bottom() int {
	return min(v.top+v.h, v.h)
}

func (v *viewport) setPos(pos, maxL, bot int) {
	v.pos = clamp(pos, 0, maxL-1)

	if v.pos < v.top {
		v.top = v.pos
	}
	if v.pos > v.top+bot-1 {
		v.top = clamp(v.pos-v.h, 0, maxL) + 1
	}
}

type Nodes []Node

func (n Nodes) Len() int {
	len := 0
	for _, node := range n {
		len++
		if node.Children() != nil {
			len += node.Children().Len()
		}
	}
	return len
}

func (n Nodes) at(i int) Node {
	for j, p := range n {
		if j == i && p.State()&NodeVisible == NodeVisible {
			return p
		}
		if p.Children() != nil {
			if nn := p.Children().at(i - j - 1); nn != nil {
				return nn
			}
		}
	}
	return nil
}

func (n Nodes) GoString() string {
	s := strings.Builder{}
	for i, nn := range n {
		s.WriteString(fmt.Sprintf(" %d => %#v\n", i, nn))
	}
	return s.String()
}

// Model is the Bubble Tea model for this user interface.
type Model struct {
	tree       Nodes
	debugNodes Nodes
	Debug      bool

	t    Treeish
	view viewport
}

func New(t Treeish) Model {
	return Model{t: t}
}

func (m *Model) bottom() int {
	bot := m.view.bottom()
	if m.Debug {
		return bot - len(m.debugNodes)
	}
	return bot
}

func (m *Model) Children() Nodes {
	return m.tree
}

// ToggleExpand toggles the expanded state of the node pointed at by m.view.pos
func (m *Model) ToggleExpand() error {
	return nil
}

// Parent moves the whole Treeish to the parent node
func (m *Model) Parent() error {
	n := m.tree.at(0)
	if n == nil {
		return fmt.Errorf("invalid node at pos %d", m.view.pos)
	}
	parent := path.Dir(n.String())
	t, err := m.t.Advance(parent)
	if err != nil {
		return err
	}
	if m.t != nil {
		m.debug("Going to parent: %s", parent)
		m.t = t
		m.view.setPos(0, visibleLines(m.tree), m.bottom())
	}
	return nil
}

// Advance moves the whole Treeish to the node m.view.pos points at
func (m *Model) Advance() error {
	n := m.tree.at(m.view.pos)
	if n == nil {
		return fmt.Errorf("invalid node at pos %d", m.view.pos)
	}
	// TODO(marius): this behaviour needs to be moved to the Treeish interface, as all implementations
	//   will need to know that a node is being collapsed or expanded.
	if pn, ok := n.(*pathNode); ok {
		if pn.state&NodeCollapsed == NodeCollapsed {
			t, err := m.t.Advance(n.String())
			if err != nil {
				return err
			}
			if m.t != nil {
				m.debug("Advancing to: %s", n.String())
				m.t = t
				m.view.setPos(0, visibleLines(m.tree), m.bottom())
			}
		}
		pn.state ^= NodeCollapsed
	}
	return nil
}

// Top moves the current position to the first element
func (m *Model) Top() error {
	m.view.setPos(0, visibleLines(m.tree), m.bottom())
	m.debug("Top: top %d, pos: %d", m.view.top, m.view.pos)
	return nil
}

// Bottom moves the current position to the last element
func (m *Model) Bottom() error {
	m.view.setPos(visibleLines(m.tree)-1, visibleLines(m.tree), m.bottom())
	m.debug("Bottom: top %d, pos: %d", m.view.top, m.view.pos)
	return nil
}

func visibleLines(n Nodes) int {
	count := 0
	for _, nn := range n {
		visible := nn.State()&NodeVisible == NodeVisible
		if visible {
			count++
		}
		count += visibleLines(nn.Children())
	}
	return count
}

// Prev moves the current position to the previous 'i'th element in the tree.
// If it's above the viewport we need to recompute the top
func (m *Model) Prev(i int) error {
	m.view.setPos(m.view.pos-i, visibleLines(m.tree), m.bottom())
	m.debug("Prev(%d): pos: %d top %d height: %d", i, m.view.pos, m.view.top, m.bottom())
	return nil
}

// Next moves the current position to the next 'i'th element in the tree
// If it's below the viewport we need to recompute the top
func (m *Model) Next(i int) error {
	m.view.setPos(m.view.pos+i, visibleLines(m.tree), m.bottom())
	m.debug("Next(%d): pos: %d top %d height: %d", i, m.view.pos, m.view.top, m.bottom())
	return nil
}

type Msg string

func (m Model) init() tea.Msg {
	return Msg("initialized")
}

func (m Model) Init() tea.Cmd {
	return m.init
}

func findNodeByPath(nodes Nodes, path string) Node {
	for _, node := range nodes {
		if filepath.Clean(node.String()) == filepath.Clean(path) {
			return node
		}
		if child := findNodeByPath(node.Children(), path); child != nil {
			return child
		}
	}
	return nil
}

func buildNodeTree(t Treeish, paths []string) (Nodes, error) {
	if len(paths) == 0 {
		return nil, nil
	}
	flatNodes := make(Nodes, len(paths))
	top := paths[0]
	topCnt := len(strings.Split(top, "/"))
	for i, p := range paths {
		st, _ := t.State(p)
		cnt := len(strings.Split(p, "/"))
		if st&NodeCollapsible == NodeCollapsible && i != 0 {
			st |= NodeCollapsed
		}
		if cnt-topCnt <= 1 {
			st |= NodeVisible
		}
		flatNodes[i] = &pathNode{
			Path:  p,
			state: st,
		}
	}
	sort.Slice(flatNodes, func(i, j int) bool {
		n1 := flatNodes[i]
		n2 := flatNodes[j]
		v1 := n1.State()&NodeCollapsible == NodeCollapsible
		v2 := n2.State()&NodeCollapsible == NodeCollapsible
		if v1 == v2 {
			return n1.String() < n2.String()
		}
		return v1 && !v2
	})

	nodes := make(Nodes, 0)
	for _, n := range flatNodes {
		ppath, _ := path.Split(n.String())
		if parent := findNodeByPath(flatNodes, ppath); parent != nil && parent != n {
			if p, ok := parent.(*pathNode); ok {
				p.Nodes = append(p.Nodes, n)
			}
		} else {
			nodes = append(nodes, n)
		}
	}
	return nodes, nil
}

func walk(m *Model) error {
	paths, err := m.t.Walk(m.view.h)
	if err != nil {
		m.err(err)
	}
	m.tree, err = buildNodeTree(m.t, paths)
	if err != nil {
		m.err(err)
	}
	return nil
}

// Update is the Tea update function which binds keystrokes to pagination.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.debugNodes = m.debugNodes[:0]
	var err error
	needsWalk := false
	switch msg := msg.(type) {
	case Msg:
		needsWalk = true
		m.debug(string(msg))
	case tea.KeyMsg:
		// TODO(marius): we can create a data type that can be passed to the model and would function as key mapping.
		//   So the dev can load the mapping from someplace.
		//   There can be one where we add all the Readline bindings for example.
		switch msg.String() {
		case "`":
			m.Debug = !m.Debug
		case "enter":
			err = m.Advance()
			needsWalk = true
		case "backspace":
			err = m.Parent()
			needsWalk = true
		case "home":
			err = m.Top()
			needsWalk = true
		case "end":
			err = m.Bottom()
			needsWalk = true
		case "k", "kk", "kkk", "kkkk":
			err = m.Prev(len(msg.String()))
			needsWalk = true
		case "up", "upup", "upupup", "upupupup":
			err = m.Prev(len(msg.String()) / 2)
			needsWalk = true
		case "j", "jj", "jjj", "jjjj":
			err = m.Next(len(msg.String()))
			needsWalk = true
		case "down", "downdown", "downdowndown", "downdowndowndown":
			err = m.Next(len(msg.String()) / 4)
			needsWalk = true
		case "pgup":
			err = m.Prev(m.view.h - 1)
			needsWalk = true
		case "pgdown":
			err = m.Next(m.view.h - 1)
			needsWalk = true
		case "o":
			err = m.ToggleExpand()
			needsWalk = true
		case "q", "esc", "ctrl+q", "ctrl+c":
			return m, tea.Quit
		default:
			m.err(fmt.Errorf("unknown key %q", msg))
		}
	case tea.WindowSizeMsg:
		m.view.h = msg.Height
		m.view.w = msg.Width
		m.view.lines = make([]string, m.view.h)
	}
	if err != nil {
		m.err(err)
	}
	if needsWalk {
		walk(&m)
		for i := range m.view.lines {
			m.view.lines[i] = ""
		}
	}
	return m, nil
}

// View renders the pagination to a string.
func (m Model) View() string {
	return m.render()
}

const (
	NodeFirstChild = 1 << iota
	NodeLastChild
)

func (m Model) renderDebugNode(t Node) string {
	style := DefaultStyle
	annotation := ""

	if t.State()&NodeDebug == NodeDebug {
		style = DebugStyle
		annotation = ">"
	}
	if t.State()&NodeError == NodeError {
		style = ErrStyle
		annotation = "!"
	}

	return style.Width(m.view.w).Render(fmt.Sprintf("%2s %s", annotation, t.String()))
}

func (m Model) renderNode(t Node, cur int, nodeHints, depth int) string {
	style := DefaultStyle
	annotation := ""
	padding := ""

	if t.State()&NodeCollapsible == NodeCollapsible {
		annotation = SquaredMinus
		if t.State()&NodeCollapsed == NodeCollapsed {
			annotation = SquaredPlus
		}
	}

	for i := 0; i <= depth; i++ {
		padding += "   " // 3 is the length of a tree opener
	}
	if nodeHints&NodeLastChild == NodeLastChild {
		padding += BoxDrawingsUpAndRight
	} else if nodeHints&NodeFirstChild == NodeFirstChild {
		padding += BoxDrawingsUpAndRight
	} else {
		padding += BoxDrawingsVerticalAndRight
	}
	padding += BoxDrawingsHorizontal

	st := t.State()
	if st&NodeDebug == NodeDebug {
		style = DebugStyle
	}
	if st&NodeError == NodeError {
		style = ErrStyle
	}

	_, name := path.Split(t.String())
	if name == "" {
		return ""
	}
	return style.Width(m.view.w).Render(fmt.Sprintf("%s%2s %s", padding, annotation, name))
}

func renderNodes(m Model, nl Nodes) []string {
	rendered := make([]string, 0)

	nlLen := len(nl)
	firstInTree := m.tree.at(0)
	topDepth := len(strings.Split(firstInTree.String(), "/"))

	for i, n := range nl {
		visible := n.State()&NodeVisible == NodeVisible
		if !visible {
			continue
		}
		isFirst := firstInTree == n

		hints := 0
		if i == 0 && isFirst {
			hints = NodeFirstChild
		} else if i+1 == nlLen {
			hints |= NodeLastChild
		}

		depth := len(strings.Split(n.String(), "/")) - topDepth
		out := m.renderNode(n, 0, hints, depth)
		if len(out) > 0 {
			rendered = append(rendered, out)
		}

		if collapsed := n.State()&NodeCollapsed == NodeCollapsed; !collapsed {
			if childLen := visibleLines(n.Children()); childLen > 0 {
				renderedChildren := renderNodes(m, n.Children())
				rendered = append(rendered, renderedChildren...)
			}
		}
	}

	return rendered
}

func (m Model) render() string {
	if m.view.h == 0 {
		return ""
	}
	cursor := m.Children()
	if cursor.Len() == 0 {
		return ""
	}

	maxLines := m.bottom()
	if m.Debug {
		maxLines -= m.debugNodes.Len()
	}
	// NOTE(marius): here we're rendering more lines than we strictly need
	rendered := renderNodes(m, cursor)

	top := clamp(m.view.top, 0, len(rendered))
	end := clamp(maxLines+top, 0, len(rendered))
	cropped := rendered[top:end]
	m.debug("Displaying: pos:%d ren:%d vis:%d/%d[%d:%d]", m.view.pos, len(rendered), len(cropped), visibleLines(cursor), top, end)
	for i, l := range cropped {
		if i == m.view.pos-top {
			l = lipgloss.Style{}.Reverse(true).Render(l)
		}
		m.view.lines[i] = l
	}

	debStart := len(m.view.lines) - len(m.debugNodes)
	if m.Debug && debStart >= 0 {
		for i, n := range m.debugNodes {
			lineIndx := debStart + i
			m.view.lines[lineIndx] = m.renderDebugNode(n)
		}
	}
	return strings.Join(m.view.lines, "\n")
}

func clamp(v, low, high int) int {
	return min(high, max(low, v))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
