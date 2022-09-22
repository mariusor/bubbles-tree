package tree

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// NodeState is used for passing information from a Treeish element to the view itself
type NodeState uint16

func (s NodeState) Is(st NodeState) bool {
	return s&st == st
}

const (
	NodeNone NodeState = 0

	// NodeCollapsed hints that the current node is collapsed
	NodeCollapsed NodeState = 1 << iota
	NodeSelected
	// NodeCollapsible hints that the current node can be collapsed
	NodeCollapsible
	// NodeHidden hints that the current node is not going to be displayed
	NodeHidden
	// NodeLastChild shows the node to be the last in the children list
	NodeLastChild
)

var (
	defaultStyle         = lipgloss.Style{}
	defaultSelectedStyle = defaultStyle.Reverse(true)
)

type Node interface {
	tea.Model
	Parent() Node
	Children() Nodes
	State() NodeState
}

func New(t Nodes) Model {
	return Model{
		KeyMap:  DefaultKeyMap(),
		Styles:  DefaultStyles(),
		Symbols: DefaultSymbols(),

		tree: t,

		viewport: viewport.New(0, 1),
		focus:    true,
	}
}

type Nodes []Node

func (n Nodes) at(i int) Node {
	j := 0
	for _, p := range n {
		if isHidden(p) {
			continue
		}
		if j == i {
			return p
		}
		if isExpanded(p) && p.Children() != nil {
			if nn := p.Children().at(i - j - 1); nn != nil {
				return nn
			}
			j += len(p.Children().visibleNodes())
		}
		j++
	}
	return nil
}

func (n Nodes) len() int {
	l := 0
	for _, node := range n {
		l++
		if node.Children() != nil {
			l += node.Children().len()
		}
	}
	return l
}

func (n Nodes) visibleNodes() Nodes {
	visible := make(Nodes, 0)
	for _, nn := range n {
		if isHidden(nn) {
			continue
		}
		visible = append(visible, nn)
		if isCollapsible(nn) && isExpanded(nn) {
			visible = append(visible, nn.Children().visibleNodes()...)
		}
	}
	return visible
}

// KeyMap defines keybindings.
// It satisfies to the github.com/charm/bubbles/help.KeyMap interface.
type KeyMap struct {
	LineUp       key.Binding
	LineDown     key.Binding
	PageUp       key.Binding
	PageDown     key.Binding
	HalfPageUp   key.Binding
	HalfPageDown key.Binding
	GotoTop      key.Binding
	GotoBottom   key.Binding

	Expand key.Binding
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
	}
}

// Styles contains style definitions for this list component. By default, these
// values are generated by DefaultStyles.
type Styles struct {
	Line     lipgloss.Style
	Ellipsis lipgloss.Style
	Selected lipgloss.Style
}

// DefaultStyles returns a set of default style definitions for this tree.
func DefaultStyles() Styles {
	return Styles{
		Line:     defaultStyle,
		Ellipsis: defaultStyle,
		Selected: defaultSelectedStyle,
	}
}

// SetStyles sets the tree Styles.
func (m *Model) SetStyles(s Styles) {
	m.Styles = s
}

type Symbol string

func (s Symbol) draw(p int) string {
	if len(s) == 0 {
		return strings.Repeat(" ", p)
	}
	sl := utf8.RuneCount([]byte(s))
	if p < sl {
		return string(s)
	}
	return strings.Repeat(" ", p-sl) + string(s)

}

type Symbols struct {
	Width int

	Vertical         Symbol
	VerticalAndRight Symbol
	UpAndRight       Symbol
	Horizontal       Symbol

	Collapsed string
	Expanded  string
	Ellipsis  string
}

func (s Symbols) Padding() string {
	return strings.Repeat(" ", s.Width)
}

// DefaultSymbols returns a set of default Symbols for drawing the tree.
func DefaultSymbols() Symbols {
	return Symbols{
		Width:            3,
		Vertical:         "│ ",
		VerticalAndRight: "├─",
		UpAndRight:       "└─",

		Ellipsis: "…",
	}
}

func (m *Model) setCurrentNode(cursor int) tea.Cmd {
	if cursor != m.cursor {
		if previous := m.currentNode(); previous != nil {
			previous.Update(previous.State() ^ NodeSelected)
		}

		m.cursor = cursor
	}
	if current := m.currentNode(); current != nil {
		current.Update(current.State() | NodeSelected)
		return m.positionChanged
	}
	return nil
}

func (m *Model) currentNode() Node {
	if m.tree == nil || m.cursor < 0 {
		return nil
	}
	return m.tree.at(m.cursor)
}

// Model is the Bubble Tea model for this user interface.
type Model struct {
	KeyMap  KeyMap
	Styles  Styles
	Symbols Symbols

	focus  bool
	cursor int

	tree Nodes

	viewport viewport.Model
}

func (m *Model) Children() Nodes {
	return m.tree
}

// MoveUp moves the selection up by any number of row.
// It can not go above the first row.
func (m *Model) MoveUp(n int) tea.Cmd {
	return m.SetCursor(m.cursor - n)
}

// MoveDown moves the selection down by any number of row.
// It can not go below the last row.
func (m *Model) MoveDown(n int) tea.Cmd {
	return m.SetCursor(m.cursor + n)
}

// GotoTop moves the selection to the first row.
func (m *Model) GotoTop() tea.Cmd {
	return m.SetCursor(0)
}

// PastBottom returns whether the viewport is scrolled beyond the last
// line. This can happen when adjusting the viewport height.
func (m *Model) PastBottom() bool {
	return m.viewport.PastBottom()
}

// GotoBottom moves the selection to the last row.
func (m *Model) GotoBottom() tea.Cmd {
	return m.SetCursor(m.tree.len() - 1)
}

// ToggleExpand toggles the expanded state of the node pointed at by m.cursor
func (m *Model) ToggleExpand() {
	n := m.currentNode()
	if n == nil {
		return
	}
	if !isCollapsible(n) {
		return
	}
	n.Update(n.State() ^ NodeCollapsed)
}

// SetWidth sets the width of the viewport of the tree.
func (m *Model) SetWidth(w int) {
	m.viewport.Width = w
}

// SetHeight sets the height of the viewport of the tree.
func (m *Model) SetHeight(h int) {
	m.viewport.Height = h
}

// Height returns the viewport height of the tree.
func (m *Model) Height() int {
	return m.viewport.Height
}

// Width returns the viewport width of the tree.
func (m *Model) Width() int {
	return m.viewport.Width
}

// YOffset returns the viewport vertical scroll position of the tree.
func (m *Model) YOffset() int {
	return m.viewport.YOffset
}

// SetYOffset sets Y offset of the tree's viewport.
func (m *Model) SetYOffset(n int) {
	m.viewport.SetYOffset(n)
}

// ScrollPercent returns the amount scrolled as a float between 0 and 1.
func (m *Model) ScrollPercent() float64 {
	if m.viewport.Height >= len(m.tree.visibleNodes()) {
		return 1.0
	}
	y := float64(m.viewport.YOffset)
	h := float64(m.viewport.Height)
	t := float64(len(m.tree.visibleNodes()))
	v := y / (t - h)
	return math.Max(0.0, math.Min(1.0, v))
}

// Cursor returns the index of the selected row.
func (m *Model) Cursor() int {
	return m.cursor
}

// SetCursor returns the index of the selected row.
func (m *Model) SetCursor(pos int) tea.Cmd {
	cursor := clamp(pos, 0, len(m.tree.visibleNodes())-1)

	yOffset := -1
	if cursor < m.viewport.YOffset {
		yOffset = cursor
	}
	if cursor > (m.viewport.YOffset + (m.viewport.Height - 1)) {
		yOffset = cursor - m.viewport.Height + 1
	}
	if yOffset > -1 {
		m.viewport.SetYOffset(yOffset)
	}
	return m.setCurrentNode(cursor)
}

type Msg string

type ErrorMsg error

func erred(err error) func() tea.Msg {
	return func() tea.Msg {
		return err
	}
}

func (m *Model) init() tea.Msg {
	return Msg("initialized")
}

func (m *Model) Init() tea.Cmd {
	return m.init
}

// Update is the Tea update function which binds keystrokes to pagination.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.focus {
		return m, nil
	}

	var err error

	switch msg := msg.(type) {
	case Msg:
		return m, m.setCurrentNode(m.cursor)
	case tea.WindowSizeMsg:
		m.SetHeight(msg.Height)
		m.SetWidth(msg.Width)
		return m, m.setCurrentNode(m.cursor)
	case tea.KeyMsg:
		var cmd tea.Cmd
		switch {
		case key.Matches(msg, m.KeyMap.LineUp):
			cmd = m.MoveUp(1)
		case key.Matches(msg, m.KeyMap.LineDown):
			cmd = m.MoveDown(1)
		case key.Matches(msg, m.KeyMap.PageUp):
			cmd = m.MoveUp(m.viewport.Height)
		case key.Matches(msg, m.KeyMap.PageDown):
			cmd = m.MoveDown(m.viewport.Height)
		case key.Matches(msg, m.KeyMap.HalfPageUp):
			cmd = m.MoveUp(m.viewport.Height / 2)
		case key.Matches(msg, m.KeyMap.HalfPageDown):
			cmd = m.MoveDown(m.viewport.Height / 2)
		case key.Matches(msg, m.KeyMap.LineDown):
			cmd = m.MoveDown(1)
		case key.Matches(msg, m.KeyMap.GotoTop):
			cmd = m.GotoTop()
		case key.Matches(msg, m.KeyMap.GotoBottom):
			cmd = m.GotoBottom()
		case key.Matches(msg, m.KeyMap.Expand):
			m.ToggleExpand()
		}
		return m, cmd
	}

	if err != nil {
		// TODO(marius): add a way to flash the model here?
		return m, erred(err)
	}
	return m, nil
}

// View renders the pagination to a string.
func (m *Model) View() string {
	renderedRows := m.render()
	m.viewport.SetContent(
		lipgloss.JoinVertical(lipgloss.Left, renderedRows...),
	)
	return m.viewport.View()
}

// Focused returns the focus state of the tree.
func (m *Model) Focused() bool {
	return m.focus
}

// Focus focuses the tree, allowing the user to move around the tree nodes.
// interact.
func (m *Model) Focus() {
	m.focus = true
}

// Blur blurs the tree, preventing selection or movement.
func (m *Model) Blur() {
	m.cursor = -1
	m.focus = false
}

func (m *Model) positionChanged() tea.Msg {
	return m.currentNode()
}

func getDepth(n Node) int {
	d := 0
	for {
		if n == nil || n.Parent() == nil {
			break
		}
		d++
		n = n.Parent()
	}
	return d
}

func (m *Model) getTreeSymbolForPos(n Node, pos, maxDepth int) string {
	if n == nil {
		return ""
	}
	if !showTreeSymbolAtPos(n, pos, maxDepth) {
		return m.Symbols.Padding()
	}
	if pos < maxDepth {
		return m.Symbols.Vertical.draw(m.Symbols.Width)
	}
	if isLastNode(n) {
		return m.Symbols.UpAndRight.draw(m.Symbols.Width)
	}
	return m.Symbols.VerticalAndRight.draw(m.Symbols.Width)
}

func showTreeSymbolAtPos(n Node, pos, maxDepth int) bool {
	if n == nil {
		return false
	}
	if pos > maxDepth {
		// NOTE(marius): We shouldn't try to compute tree Symbols for a position larger
		//  than the current node's parent depth
		return false
	}
	if maxDepth == pos {
		return true
	}
	parentInPos := maxDepth - pos
	for i := 0; i < parentInPos; i++ {
		if n = n.Parent(); n == nil {
			return false
		}
	}
	return !isLastNode(n)
}

func (m *Model) drawTreeElementsForNode(t Node) string {
	maxDepth := getDepth(t)

	treeSymbolsPrefix := strings.Builder{}
	for i := 0; i <= maxDepth; i++ {
		treeSymbolsPrefix.WriteString(m.getTreeSymbolForPos(t, i, maxDepth))
	}
	return treeSymbolsPrefix.String()
}

func (m *Model) renderNode(t Node) string {
	if t == nil {
		return ""
	}

	prefix := ""
	annotation := ""

	name := t.View()
	hints := t.State()

	prefix = fmt.Sprintf("%s%-1s", m.drawTreeElementsForNode(t), annotation)

	name = m.ellipsize(name, m.viewport.Width-strings.Count(prefix, ""))
	t.Update(hints)

	render := m.Styles.Line.Width(m.Width()).Render
	if isSelected(t) {
		render = m.Styles.Selected.Width(m.Width()).Render
	}
	node := render(fmt.Sprintf("%s%s", prefix, name))

	if isExpanded(t) && len(t.Children()) > 0 {
		renderedChildren := m.renderNodes(t.Children())
		node = lipgloss.JoinVertical(
			lipgloss.Left,
			node,
			lipgloss.JoinVertical(lipgloss.Left, renderedChildren...),
		)
	}
	return node
}

func (m *Model) ellipsize(s string, w int) string {
	if w > len(s) || w < 1 {
		return s
	}
	b := strings.Builder{}
	b.WriteString(s[:w-1])
	b.WriteString(m.Symbols.Ellipsis)
	return b.String()
}

func isHidden(n Node) bool {
	return n.State().Is(NodeHidden)
}

func isExpanded(n Node) bool {
	return !n.State().Is(NodeCollapsed)
}

func isCollapsible(n Node) bool {
	return n.State().Is(NodeCollapsible)
}

func isLastNode(n Node) bool {
	return n.State().Is(NodeLastChild)
}

func isSelected(n Node) bool {
	return n.State().Is(NodeSelected)
}

func (m *Model) renderNodes(nl Nodes) []string {
	if len(nl) == 0 || len(m.tree) == 0 {
		return nil
	}

	rendered := make([]string, 0)

	for i, n := range nl {
		if isHidden(n) {
			continue
		}

		var hints NodeState = 0
		if len(n.Children()) > 0 {
			hints |= NodeCollapsible
		}
		if i == len(nl)-1 {
			hints |= NodeLastChild
		}
		n.Update(n.State() | hints)
		if out := m.renderNode(n); len(out) > 0 {
			rendered = append(rendered, out)
		}
	}

	return rendered
}

func (m *Model) render() []string {
	if m.viewport.Height == 0 {
		return nil
	}

	return m.renderNodes(m.Children())
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
