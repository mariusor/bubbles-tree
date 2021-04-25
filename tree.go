package tree

import (
	"fmt"
	"io"
	"os"
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
	Walk(int) ([]string, error)
	State(s string) (NodeState, error)
}

type cursor struct {
	h, w      int
	top, left int
}

type errModel struct {
	Debug bool
	tree  map[NodeState][]string
}

// Model is the Bubble Tea model for this user interface.
type Model struct{
	t      Treeish
	cur    cursor
	tree   []string
}

func New(t Treeish) *Model {
	m := new(Model)
	m.t = t
	m.tree = make([]string, 0)
	return m
}

func (m *Model) Prev(i int) error {
	m.cur.top = clamp(m.cur.top-i, 0, max(len(m.tree)-m.cur.h, m.cur.h))
	return nil
}

func (m *Model) Next(i int) error {
	m.cur.top = clamp(m.cur.top+i, 0, max(len(m.tree)-m.cur.h, m.cur.h))
	return nil
}

type TreeMsg string

func (m* Model) init() tea.Msg {
	walk(m)
	return TreeMsg("inited")
}

func (m *Model) Init() tea.Cmd {
	return m.init
}

func walk(m *Model) error {
	var err error
	if m.tree, err = m.t.Walk(m.cur.h); err != nil {
		fmt.Fprintf(os.Stderr, "Err: %s\n", err)
	}
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
		case "up", "k":
			err = m.Prev(1)
		case "down", "j":
			err = m.Next(1)
			needsWalk = true
		case "pgup":
			err = m.Prev(m.cur.h)
			needsWalk = true
		case "pgdown":
			err = m.Next(m.cur.h)
			needsWalk = true
		case "q", "esc", "ctrl+q", "ctrl+c":
			return m, tea.Quit
		default:
			fmt.Fprintf(os.Stderr, "unknown key %s\n", msg.String())
		}
	case tea.WindowSizeMsg:
		m.cur.h = msg.Height
		m.cur.w = msg.Width
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

func (m Model) renderNode (t string, builder io.StringWriter) error {
	style := defaultStyle
	padding := 5
	annotation := ""
	st, err := m.t.State(t)
	if st & NodeCollapsed == NodeCollapsed {
		annotation = SquaredMinus
	}
	if st & NodeCollapsible == NodeCollapsible {
		annotation = SquaredPlus
	}

	//_, file := path.Split(t)
	builder.WriteString(fmt.Sprintf("%4s ", annotation))
	builder.WriteString(style.Width(m.cur.w-padding).Render(t))
	builder.WriteString("\n")
	return err
}

func (m Model) render() string {
	top := clamp(m.cur.top, 0, max(len(m.tree)-m.cur.h, m.cur.h))
	bot := clamp(top+m.cur.h, m.cur.h, len(m.tree))

	builder := strings.Builder{}

	var nodes []string
	if len(m.tree) > bot {
		nodes = m.tree[top:bot]
	}
	for _, n := range nodes {
		m.renderNode(n, &builder)
	}
	return builder.String()
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
