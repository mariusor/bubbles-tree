package tree

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// NodeState is used for passing information from a Treeish element to the view itself
type NodeState int

const (
	// NodeCollapsed hints that the current node is collapsed
	NodeCollapsed NodeState = 1 << iota
	// NodeCollapsible hints that the current node can be collapsed
	NodeCollapsible
	// NodeVisible hints that the current node is ready to be displayed
	NodeVisible
	NodeError

	NodeNone = 0
)

const (
	BoxDrawingsVerticalAndRight = "├"
	BoxDrawingsVertical         = "│"
	BoxDrawingsUpAndRight       = "└"
	BoxDrawingsDownAndRight     = "┌"
	BoxDrawingsHorizontal       = "─"

	SquaredPlus  = "⊞"
	SquaredMinus = "⊟"

	Ellipsis = "…"
)

var (
	defaultStyle         = lipgloss.Style{}
	defaultSelectedStyle = defaultStyle.Reverse(true)
)

type Treeish interface {
	// Advance moves the Treeish to a new received path,
	// this can return a new Treeish instance at the new path, or perform some other function
	// for the cases where the path doesn't correspond to a Treeish object.
	// Specifically in the case of the filepath Treeish:
	// If a passed path parameter corresponds to a folder, it will return a new Treeish object at the new path
	// If the passed path parameter corresponds to a file, it returns a nil Treeish, but it can execute something else.
	// Eg, When being passed a path that corresponds to a text file, another bubbletea function corresponding to a
	// viewer can be called from here.
	Advance(string) (Treeish, error)
	// State returns the NodeState for the received path parameter
	// This is used when rendering the path in the tree view
	State(string) (NodeState, error)

	// Walk loads the elements of current Treeish and returns them as a flat list of maximum elements
	Walk(int) ([]string, error)
}

func (m *Model) debug(s string, params ...interface{}) {
	m.LogFn(s, params...)
}

func (m *Model) err(err error) {
	if err != nil {
		m.LogFn("error: %s", err.Error())
	}
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

func (n *pathNode) SetState(s NodeState) {
	n.state = s
}

type Node interface {
	String() string
	Children() Nodes
	State() NodeState
	SetState(NodeState)
}

// MoveUp moves the selection up by any number of row.
// It can not go above the first row.
func (m *Model) MoveUp(n int) {
	m.cursor = clamp(m.cursor-n, 0, m.tree.Len()-1)
	m.debug("move %d, new pos: %d", n, m.cursor)

	if m.cursor < m.viewport.YOffset {
		m.debug("viewport adjustment %d", n)
		m.viewport.LineUp(n)
	}
	m.UpdateViewport()
}

// MoveDown moves the selection down by any number of row.
// It can not go below the last row.
func (m *Model) MoveDown(n int) {
	m.cursor = clamp(m.cursor+n, 0, m.tree.Len()-1)
	m.debug("move %d, new pos: %d", n, m.cursor)

	if m.cursor > (m.viewport.YOffset + (m.viewport.Height - 1)) {
		m.debug("viewport adjustment %d", n)
		m.viewport.LineDown(n)
	}
	m.UpdateViewport()
}

// GotoTop moves the selection to the first row.
func (m *Model) GotoTop() {
	m.MoveUp(m.cursor)
}

// GotoBottom moves the selection to the last row.
func (m *Model) GotoBottom() {
	m.MoveDown(m.tree.Len())
}

type Nodes []Node

func (n Nodes) Len() int {
	l := 0
	for _, node := range n {
		l++
		if node.Children() != nil {
			l += node.Children().Len()
		}
	}
	return l
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

// KeyMap defines keybindings.
// It satisfies to the github.com/charm/bubbles/help.KeyMap interface, which is used to render the menu.
type KeyMap struct {
	LineUp       key.Binding
	LineDown     key.Binding
	PageUp       key.Binding
	PageDown     key.Binding
	HalfPageUp   key.Binding
	HalfPageDown key.Binding
	GotoTop      key.Binding
	GotoBottom   key.Binding

	Expand  key.Binding
	Advance key.Binding
	Parent  key.Binding
}

// DefaultKeyMap returns a default set of keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		LineUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		LineDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("b", "pgup"),
			key.WithHelp("b/pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("f", "pgdown", " "),
			key.WithHelp("f/pgdn", "page down"),
		),
		HalfPageUp: key.NewBinding(
			key.WithKeys("u", "ctrl+u"),
			key.WithHelp("u", "½ page up"),
		),
		HalfPageDown: key.NewBinding(
			key.WithKeys("d", "ctrl+d"),
			key.WithHelp("d", "½ page down"),
		),
		GotoTop: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("g/home", "go to start"),
		),
		GotoBottom: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("G/end", "go to end"),
		),
		Expand: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "toggle expand for current node"),
		),
		Advance: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "open this node"),
		),
		Parent: key.NewBinding(
			key.WithKeys("backspace"),
			key.WithHelp("backspace", "go back to the parent node"),
		),
	}
}

// Styles contains style definitions for this list component. By default, these
// values are generated by DefaultStyles.
type Styles struct {
	Line     lipgloss.Style
	Ellipsis lipgloss.Style
	Selected lipgloss.Style
}

// DefaultStyles returns a set of default style definitions for this table.
func DefaultStyles() Styles {
	return Styles{
		Line:     defaultStyle,
		Ellipsis: defaultStyle,
		Selected: defaultSelectedStyle,
	}
}

// SetStyles sets the table styles.
func (m *Model) SetStyles(s Styles) {
	m.styles = s
	m.UpdateViewport()
}

// UpdateViewport updates the list content based on the previously defined
// columns and rows.
func (m *Model) UpdateViewport() {
	renderedRows := m.render()
	m.viewport.SetContent(
		lipgloss.JoinVertical(lipgloss.Left, renderedRows...),
	)
}

type logFn func(s string, args ...interface{})

func emptyLog(_ string, _ ...interface{}) {}

// Model is the Bubble Tea model for this user interface.
type Model struct {
	KeyMap   KeyMap
	viewport viewport.Model

	cursor int
	focus  bool
	styles Styles

	tree  Nodes
	LogFn logFn

	t Treeish
}

func New(t Treeish) Model {
	return Model{
		t: t,

		viewport: viewport.New(0, 1),
		focus:    true,

		KeyMap: DefaultKeyMap(),
		styles: DefaultStyles(),
		LogFn:  emptyLog,
	}
}

func (m *Model) Children() Nodes {
	return m.tree
}

// ToggleExpand toggles the expanded state of the node pointed at by m.cursor
func (m *Model) ToggleExpand() error {
	n := m.tree.at(m.cursor)
	m.LogFn("TODO: expanding: %s", n)
	return nil
}

// Parent moves the whole Treeish to the parent node
func (m *Model) Parent() error {
	n := m.tree.at(0)
	if n == nil {
		return fmt.Errorf("invalid node at pos %d", m.cursor)
	}
	parent := path.Dir(n.String())
	t, err := m.t.Advance(parent)
	if err != nil {
		return err
	}
	m.debug("Going to parent: %s", parent)
	m.t = t
	walk(m)
	m.GotoTop()

	n.SetState(n.State() | NodeCollapsed)
	m.UpdateViewport()
	return nil
}

// Advance moves the whole Treeish to the node m.cursor points at
func (m *Model) Advance() error {
	n := m.tree.at(m.cursor)
	if n == nil {
		return fmt.Errorf("invalid node at pos %d", m.cursor)
	}

	t, err := m.t.Advance(n.String())
	if err != nil {
		return err
	}
	m.debug("Advancing to: %s", n.String())
	m.t = t
	walk(m)
	m.GotoTop()

	n.SetState(n.State() ^ NodeCollapsed)
	m.UpdateViewport()
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

// SetWidth sets the width of the viewport of the table.
func (m *Model) SetWidth(w int) {
	m.viewport.Width = w
	m.UpdateViewport()
}

// SetHeight sets the height of the viewport of the table.
func (m *Model) SetHeight(h int) {
	m.viewport.Height = h
	m.UpdateViewport()
}

// Height returns the viewport height of the table.
func (m *Model) Height() int {
	return m.viewport.Height
}

// Width returns the viewport width of the table.
func (m *Model) Width() int {
	return m.viewport.Width
}

// Cursor returns the index of the selected row.
func (m *Model) Cursor() int {
	return m.cursor
}

type Msg string

func (m *Model) init() tea.Msg {
	m.err(walk(m))
	m.UpdateViewport()
	return Msg("initialized")
}

func (m *Model) Init() tea.Cmd {
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
		var ppath string
		if u, err := url.Parse(n.String()); err == nil {
			ppath, _ = path.Split(u.Path)
		} else {
			ppath, _ = path.Split(n.String())
		}
		if parent := findNodeByPath(flatNodes, ppath); parent != nil && ppath != parent.String() {
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
	//m.debug("walking treeish: %v", m.t)
	paths, err := m.t.Walk(m.viewport.Height)
	if err != nil {
		m.err(err)
	}
	//m.debug("loaded %d paths", len(paths))
	m.tree, err = buildNodeTree(m.t, paths)
	if err != nil {
		m.err(err)
	}
	//m.debug("built %d nodes", m.tree.Len())
	return nil
}

// Focused returns the focus state of the table.
func (m *Model) Focused() bool {
	return m.focus
}

// Focus focusses the table, allowing the user to move around the rows and
// interact.
func (m *Model) Focus() {
	m.focus = true
	m.UpdateViewport()
}

// Blur blurs the table, preventing selection or movement.
func (m *Model) Blur() {
	m.focus = false
	m.UpdateViewport()
}

// Update is the Tea update function which binds keystrokes to pagination.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.focus {
		return m, nil
	}

	var err error

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetHeight(msg.Height)
		m.SetWidth(msg.Width)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Expand):
			err = m.ToggleExpand()
		case key.Matches(msg, m.KeyMap.Advance):
			err = m.Advance()
		case key.Matches(msg, m.KeyMap.Parent):
			err = m.Parent()
		case key.Matches(msg, m.KeyMap.LineUp):
			m.MoveUp(1)
		case key.Matches(msg, m.KeyMap.LineDown):
			m.MoveDown(1)
		case key.Matches(msg, m.KeyMap.PageUp):
			m.MoveUp(m.viewport.Height)
		case key.Matches(msg, m.KeyMap.PageDown):
			m.MoveDown(m.viewport.Height)
		case key.Matches(msg, m.KeyMap.HalfPageUp):
			m.MoveUp(m.viewport.Height / 2)
		case key.Matches(msg, m.KeyMap.HalfPageDown):
			m.MoveDown(m.viewport.Height / 2)
		case key.Matches(msg, m.KeyMap.LineDown):
			m.MoveDown(1)
		case key.Matches(msg, m.KeyMap.GotoTop):
			m.GotoTop()
		case key.Matches(msg, m.KeyMap.GotoBottom):
			m.GotoBottom()
		}
	}
	if err != nil {
		m.err(err)
	}
	return m, nil
}

// View renders the pagination to a string.
func (m *Model) View() string {
	return m.viewport.View()
}

const (
	NodeFirstChild = 1 << iota
	NodeLastChild
)

func (m *Model) renderNode(t Node, nodeHints, depth int) string {
	style := defaultStyle
	annotation := ""
	padding := ""

	if t.State()&NodeCollapsible == NodeCollapsible {
		annotation = SquaredMinus
		if t.State()&NodeCollapsed == NodeCollapsed {
			annotation = SquaredPlus
		}
	}

	for i := 0; i < depth; i++ {
		padding += "  " // 2 is the length of a tree opener
	}
	if nodeHints&NodeLastChild == NodeLastChild {
		padding += BoxDrawingsUpAndRight
	} else if nodeHints&NodeFirstChild == NodeFirstChild {
		padding += BoxDrawingsUpAndRight
	} else {
		padding += BoxDrawingsVerticalAndRight
	}
	padding += BoxDrawingsHorizontal

	base, name := path.Split(t.String())
	if name == "" {
		name = base
	}
	prefix := fmt.Sprintf("%s%-2s", padding, annotation)
	name = ellipsize(name, m.viewport.Width-strings.Count(prefix, ""))
	return style.Width(m.viewport.Width).Render(fmt.Sprintf("%s%s", prefix, name))
}

func ellipsize(s string, w int) string {
	if w > len(s) || w < 0 {
		return s
	}
	b := strings.Builder{}
	b.WriteString(s[:w-1])
	b.WriteString(Ellipsis)
	return b.String()
}

func (m *Model) renderNodes(nl Nodes) []string {
	rendered := make([]string, 0)

	nlLen := len(nl)
	firstInTree := m.tree.at(0)
	startsWithRoot := false
	if firstInTree.String() == "/" {
		startsWithRoot = true
	}
	topDepth := len(strings.Split(firstInTree.String(), "/"))

	for i, n := range nl {
		visible := n.State()&NodeVisible == NodeVisible
		if !visible {
			continue
		}
		isFirst := firstInTree == n

		depth := len(strings.Split(n.String(), string(os.PathSeparator))) - topDepth

		hints := 0
		if i == 0 && isFirst {
			hints = NodeFirstChild
		} else if i+1 == nlLen {
			hints |= NodeLastChild
		}

		if startsWithRoot && !isFirst {
			depth += 1
		}
		out := m.renderNode(n, hints, depth)
		if len(out) > 0 {
			rendered = append(rendered, out)
		}

		if collapsed := n.State()&NodeCollapsed == NodeCollapsed; !collapsed {
			if childLen := visibleLines(n.Children()); childLen > 0 {
				renderedChildren := m.renderNodes(n.Children())
				rendered = append(rendered, renderedChildren...)
			}
		}
	}

	return rendered
}

func (m *Model) render() []string {
	if m.viewport.Height == 0 {
		return nil
	}
	cursor := m.Children()
	if cursor.Len() == 0 {
		return nil
	}

	rendered := m.renderNodes(cursor)
	lines := make([]string, 0)

	for i, l := range rendered {
		if i == m.cursor {
			l = m.styles.Selected.Render(l)
		} else {
			l = m.styles.Line.Render(l)
		}
		lines = append(lines, l)
	}

	return lines
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
