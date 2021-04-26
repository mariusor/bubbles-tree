package tree

import (
	"fmt"
	"os"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type NodeState int

const (
	NodeCollapsed = 1 << iota
	NodeCollapsible
	NodeError
	NodeDebug
)

type Treeish interface {
	Advance(string) Treeish
	Walk(int) ([]string, error)
	State(string) (NodeState, error)
}

type viewport struct {
	h, w      int
	top, left int
}

// Model is the Bubble Tea model for this user interface.
type Model struct {
	// the index of the current element in the 'tree'
	pos  int
	tree []string

	t    Treeish
	view viewport
}

func New(t Treeish) *Model {
	m := new(Model)
	m.t = t
	return m
}

func (m *Model) Lines() []string {
	bot := min(m.view.top+m.view.h, len(m.tree))
	top := clamp(m.view.top, 0, len(m.tree)-bot)

	//fmt.Fprintf(os.Stderr, "h:%d p:%d :: t:%d b:%d l:%d\n", m.view.h, m.pos, m.view.top, bot, len(m.tree))
	return m.tree[top:bot]
}

func (m Model) nodeAt(i int) string {
	for j, p := range m.tree {
		if j == i {
			return p
		}
	}
	return ""
}

// ToggleExpand
func (m *Model) ToggleExpand() error {
	cur := m.nodeAt(m.pos)

	currentTree := m.tree
	m.t = m.t.Advance(cur)

	err := walk(m)
	if err != nil {
		return err
	}
	newTree := m.tree

	m.tree = append(currentTree[:m.pos+1], newTree...)
	m.tree = append(m.tree, currentTree[m.pos:]...)

	return nil
}

// Prev moves the current position to the previous 'i'th element in the tree.
// If it's above the viewport we need to recompute the top
func (m *Model) Prev(i int) error {
	m.pos = clamp(m.pos-i, 0, len(m.tree)-1)
	if m.pos < m.view.top {
		m.view.top = clamp(m.pos, 0, max(len(m.tree)-m.view.h, m.view.h))
	}
	return nil
}

// Next moves the current position to the next 'i'th element in the tree
// If it's below the viewport we need to recompute the top
func (m *Model) Next(i int) error {
	m.pos = clamp(m.pos+i, 0, len(m.tree)-1)
	bot := min(m.view.top+m.view.h-1, len(m.tree))
	if m.pos > bot {
		m.view.top = clamp(bot+i, 0, len(m.tree)-1)
	}
	return nil
}

type TreeMsg string

func (m *Model) init() tea.Msg {
	walk(m)
	return TreeMsg("inited")
}

func (m *Model) Init() tea.Cmd {
	return m.init
}

func walk(m *Model) error {
	paths, err := m.t.Walk(m.view.h - 5)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err: %s\n", err)
	}
	sort.Slice(paths, func(i, j int) bool {
		f1, _ := os.Stat(paths[i])
		if f1 == nil {
			return false
		}
		f2, _ := os.Stat(paths[j])
		if f2 == nil {
			return true
		}
		if f1.IsDir() {
			if f2.IsDir() {
				return f1.Name() < f2.Name()
			}
			return true
		}
		return false
	})
	m.tree = paths
	return nil
}

// Update is the Tea update function which binds keystrokes to pagination.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var err error
	needsWalk := false
	switch msg := msg.(type) {
	//case TreeMsg:
	//	m.tree = append(m.tree, DebugNode{string(msg)})
	case tea.KeyMsg:
		// TODO(marius): we can create a data type that can be passed to the model and would function as key mapping.
		//   So the dev can load the mapping from someplace.
		//   There can be one where we add all the Readline bindings for example.
		switch msg.String() {
		case "enter":
			err = m.ToggleExpand()
		case "up", "k":
			err = m.Prev(1)
		case "down", "j":
			err = m.Next(1)
			needsWalk = true
		case "pgup":
			err = m.Prev(m.view.h)
			needsWalk = true
		case "pgdown":
			err = m.Next(m.view.h)
			needsWalk = true
		case "q", "esc", "ctrl+q", "ctrl+c":
			return m, tea.Quit
		default:
			fmt.Fprintf(os.Stderr, "unknown key %s\n", msg.String())
		}
	case tea.WindowSizeMsg:
		m.view.h = msg.Height
		m.view.w = msg.Width
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err: %s\n", err)
	}
	if needsWalk {
		walk(m)
	}
	return m, nil
}

// View renders the pagination to a string.
func (m Model) View() string {
	return m.render()
}

func (m Model) renderLine(i int) string {
	style := defaultStyle
	annotation := ""
	t := m.tree[i]
	st, _ := m.t.State(t)
	if st&NodeCollapsed == NodeCollapsed {
		annotation = "-"
	}
	if st&NodeCollapsible == NodeCollapsible {
		annotation = "+"
	}
	if i == min(m.pos, m.pos+m.view.top) {
		style = highlightStyle
	}

	return style.Render(fmt.Sprintf("%4s %s", annotation, t))
}

func (m Model) render() string {
	//fmt.Fprintf(os.Stderr, "WxH %dx%d - t:b %d:%d nc: %d\n", m.view.w, m.view.h, top, bot, len(m.tree))
	cursor := m.Lines()
	lines := make([]string, len(cursor))
	for i := range cursor {
		lines[i] = m.renderLine(i)
	}
	return strings.Join(lines, "\n")
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
