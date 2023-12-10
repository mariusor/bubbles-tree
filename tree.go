package tree

import (
	"math"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"
)

var (
	defaultStyle         = lipgloss.NewStyle()
	defaultSelectedStyle = defaultStyle.Reverse(true)
	defaultSymbolStyle   = defaultStyle
)

// New initializes a new Model
// It sets the default keymap, styles and symbols.
func New(t Nodes) Model {
	vp := viewport.New(0, 0)
	return Model{
		Model:   &vp,
		KeyMap:  DefaultKeyMap(),
		Styles:  DefaultStyles(),
		Symbols: DefaultSymbols(),

		tree: t,

		focus: true,
	}
}

// KeyMap defines keybindings.
// It satisfies the github.com/charm/bubbles/help.KeyMap interface.
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

type DepthStyler interface {
	Width(int) DepthStyler
	Render(depth int, strs ...string) string
}

type Style lipgloss.Style

func (s Style) Render(_ int, strs ...string) string {
	return lipgloss.Style(s).Render(strs...)
}

func (s Style) Width(w int) DepthStyler {
	ss := lipgloss.Style(s).Width(w)
	return Style(ss)
}

// Styles contains style definitions for this list component. By default, these
// values are generated by DefaultStyles.
type Styles struct {
	Line     lipgloss.Style
	Selected lipgloss.Style
	Symbol   DepthStyler
}

// DefaultStyles returns a set of default style definitions for this tree.
func DefaultStyles() Styles {
	return Styles{
		Line:     defaultStyle,
		Selected: defaultSelectedStyle,
		Symbol:   Style(defaultSymbolStyle),
	}
}

// SetStyles sets the tree Styles.
func (m *Model) SetStyles(s Styles) {
	m.Styles = s
}

func draw(style DepthStyler, s string, width, depth int) string {
	return style.Width(width).Render(depth, s)
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
	return noop
}

func (m *Model) currentNode() Node {
	if m.tree == nil || m.cursor < 0 {
		return nil
	}
	return m.tree.at(m.cursor)
}

// Model is the Bubble Tea model for this user interface.
type Model struct {
	*viewport.Model

	KeyMap  KeyMap
	Styles  Styles
	Symbols Symbols

	focus  bool
	cursor int

	tree Nodes
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
	return m.Model.PastBottom()
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
	m.Model.Width = w
}

// SetHeight sets the height of the viewport of the tree.
func (m *Model) SetHeight(h int) {
	m.Model.Height = h
	m.updateNodeVisibility(h)
}

// Height returns the viewport height of the tree.
func (m *Model) Height() int {
	return m.Model.Height
}

// Width returns the viewport width of the tree.
func (m *Model) Width() int {
	return m.Model.Width
}

// YOffset returns the viewport vertical scroll position of the tree.
func (m *Model) YOffset() int {
	return m.Model.YOffset
}

// SetYOffset sets Y offset of the tree's viewport.
func (m *Model) SetYOffset(n int) {
	m.Model.SetYOffset(n)
	m.updateNodeVisibility(m.Height())
}

// ScrollPercent returns the amount scrolled as a float between 0 and 1.
func (m *Model) ScrollPercent() float64 {
	if m.Model.Height >= len(m.tree.visibleNodes()) {
		return 1.0
	}
	y := float64(m.Model.YOffset)
	h := float64(m.Model.Height)
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
	if cursor == m.cursor {
		return noop
	}

	yOffset := -1
	if cursor < m.Model.YOffset {
		yOffset = cursor
	}
	if cursor > (m.Model.YOffset + (m.Model.Height - 1)) {
		yOffset = cursor - m.Model.Height + 1
	}
	if yOffset > -1 {
		m.Model.SetYOffset(yOffset)
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

var noop tea.Cmd = nil

func (m *Model) init() tea.Msg {
	return Msg("initialized")
}

func (m *Model) Init() tea.Cmd {
	return m.init
}

func (m *Model) updateNodeVisibility(height int) tea.Cmd {
	if height == 0 {
		return noop
	}
	start := m.YOffset()
	end := start + height

	cmds := make([]tea.Cmd, 0)
	for i, nn := range m.tree.visibleNodes() {
		if i >= start && i < end {
			continue
		}
		_, cmd := nn.Update(nn.State() | nodeSkipRender)
		cmds = append(cmds, cmd)
	}
	return tea.Batch()
}

// Update is the Tea update function which binds keystrokes to pagination.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.focus {
		return m, noop
	}

	var err error

	switch msg := msg.(type) {
	case Msg:
		return m, m.setCurrentNode(m.cursor)
	case tea.WindowSizeMsg:
		m.SetWidth(msg.Width)
		m.SetHeight(msg.Height)
		//cmd := m.tree.Update(tea.WindowSizeMsg{Width: msg.Width, Height: 1})
		return m, tea.Batch(m.setCurrentNode(m.cursor))
	case tea.KeyMsg:
		var cmd tea.Cmd
		switch {
		case key.Matches(msg, m.KeyMap.LineUp):
			cmd = m.MoveUp(1)
		case key.Matches(msg, m.KeyMap.LineDown):
			cmd = m.MoveDown(1)
		case key.Matches(msg, m.KeyMap.PageUp):
			cmd = m.MoveUp(m.Model.Height)
		case key.Matches(msg, m.KeyMap.PageDown):
			cmd = m.MoveDown(m.Model.Height)
		case key.Matches(msg, m.KeyMap.HalfPageUp):
			cmd = m.MoveUp(m.Model.Height / 2)
		case key.Matches(msg, m.KeyMap.HalfPageDown):
			cmd = m.MoveDown(m.Model.Height / 2)
		case key.Matches(msg, m.KeyMap.LineDown):
			cmd = m.MoveDown(1)
		case key.Matches(msg, m.KeyMap.GotoTop):
			cmd = m.GotoTop()
		case key.Matches(msg, m.KeyMap.GotoBottom):
			cmd = m.GotoBottom()
		case key.Matches(msg, m.KeyMap.Expand):
			m.ToggleExpand()
		}

		return m, tea.Batch(cmd, m.updateNodeVisibility(m.Height()))
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
	m.Model.SetContent(
		lipgloss.JoinVertical(lipgloss.Left, renderedRows...),
	)
	return m.Model.View()
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

// When we render the tree symbols we consider them as a grid of maxDepth width
// Each pos in the grid corresponds to a
func (m *Model) getTreeSymbolForPos(n Node, pos, maxDepth int) string {
	if n == nil {
		return ""
	}
	s := m.Styles.Symbol
	if renderPaddingAtPos(n, pos, maxDepth) {
		return Padding(s, m.Symbols, pos)
	}
	if pos < maxDepth {
		return RenderConnector(s, m.Symbols, pos)
	}
	if isLastNode(n) {
		return RenderTerminator(s, m.Symbols, pos)
	}
	return RenderStarter(s, m.Symbols, pos)
}

// renderPaddingAtPos computes the tree symbol for a Node for a specific depth
func renderPaddingAtPos(n Node, depth, maxDepth int) bool {
	if n == nil {
		return true
	}
	if depth > maxDepth {
		return true
	}
	if depth == maxDepth {
		return false
	}
	parentInPos := maxDepth - depth
	for i := 0; i < parentInPos; i++ {
		if n = n.Parent(); n == nil {
			return true
		}
	}
	return isLastNode(n)
}

func (m *Model) renderPrefixForSingleLineNode(t Node) string {
	maxDepth := getDepth(t)

	prefix := strings.Builder{}
	for pos := 0; pos <= maxDepth; pos++ {
		prefix.WriteString(m.getTreeSymbolForPos(t, pos, maxDepth))
	}
	return prefix.String()
}

func (m *Model) renderPrefixForMultiLineNode(t Node, lineCount int) string {
	maxDepth := getDepth(t)

	s := m.Styles.Symbol

	prefix := strings.Builder{}

	connectsBottom := isLastNode(t)
	for line := 0; line < lineCount; line++ {
		for lvl := 0; lvl <= maxDepth-1; lvl++ {
			prefix.WriteString(m.getTreeSymbolForPos(t, lvl, maxDepth))
		}
		if line == 0 {
			prefix.WriteString(RenderStarter(s, m.Symbols, maxDepth))
			if lineCount > 1 {
				prefix.WriteRune('\n')
			}
		} else if line == lineCount-1 {
			if !connectsBottom {
				prefix.WriteString(RenderTerminator(s, m.Symbols, maxDepth))
			} else {
				prefix.WriteString(RenderConnector(s, m.Symbols, maxDepth))
			}
		} else {
			prefix.WriteString(RenderConnector(s, m.Symbols, maxDepth))
			prefix.WriteRune('\n')
		}
	}

	return prefix.String()
}

const Ellipsis = "…"

func (m *Model) renderNode(t Node) string {
	if t == nil {
		return ""
	}

	prefix := ""

	name := t.View()

	style := m.Styles.Line
	if isSelected(t) {
		style = m.Styles.Selected
	}

	if lineCount := lipgloss.Height(name); lineCount > 1 {
		prefix = m.renderPrefixForMultiLineNode(t, lineCount)
	} else {
		prefix = m.renderPrefixForSingleLineNode(t)
	}

	pw := lipgloss.Width(prefix)
	nw := m.Width() - pw
	render := style.Width(nw).MaxWidth(nw - 1).Render
	if lipgloss.Width(name) > nw {
		name = truncate.StringWithTail(name, uint(nw-1), Ellipsis)
	}
	node := lipgloss.JoinHorizontal(lipgloss.Left, prefix, render(name))
	if isExpanded(t) && len(t.Children()) > 0 {
		renderedChildren := m.renderNodes(t.Children())
		node = lipgloss.JoinVertical(lipgloss.Top, node, lipgloss.JoinVertical(lipgloss.Left, renderedChildren...))
	}

	return node
}

func isHidden(n Node) bool {
	return n.State().Is(NodeHidden)
}

func skipRender(n Node) bool {
	return n.State().Is(nodeSkipRender)
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

func hasPreviousSibling(n Node) bool {
	return n.State().Is(nodeHasPreviousSibling)
}

func isMultiLine(n Node) bool {
	return n.State().Is(NodeIsMultiLine)
}

func (m *Model) renderNodes(nl Nodes) []string {
	if len(nl) == 0 {
		return nil
	}

	rendered := make([]string, 0)

	for i, n := range nl {
		if isHidden(n) || skipRender(n) {
			continue
		}
		var hints NodeState = 0

		if len(nl) > 0 && i > 0 {
			hints |= nodeHasPreviousSibling
		}
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
	if m.Model.Height+m.Model.Width == 0 {
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
