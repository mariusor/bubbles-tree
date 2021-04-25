package tree

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

var Debug = true
var DebugLines = make([]string, 0)

type Treeish interface {
	Walk() ([]string, error)
}

type cursor struct {
	h, w      int
	top, left int
}

// Model is the Bubble Tea model for this user interface.
type Model struct{
	Err    error
	t      Treeish
	cur    cursor
	tree   []string
}

func New(t Treeish) *Model {
	m := new(Model)
	m.t = t
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
	m.tree, m.Err = m.t.Walk()
	return TreeMsg("inited")
}

func (m *Model) Init() tea.Cmd {
	return m.init
}

// Update is the Tea update function which binds keystrokes to pagination.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var err error
	DebugLines = DebugLines[:0]
	switch msg := msg.(type) {
	case TreeMsg:
		fmt.Printf("%s", string(msg))
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			err = m.Prev(1)
		case "down", "j":
			err = m.Next(1)
		case "pgup":
			err = m.Prev(m.cur.h)
		case "pgdown":
			err = m.Next(m.cur.h)
		case "q", "esc", "ctrl+q", "ctrl+c":
			return m, tea.Quit
		default:
			m.Err = fmt.Errorf("unknown key %s", msg.String())
		}
	case tea.WindowSizeMsg:
		m.cur.h = msg.Height
		m.cur.w = msg.Width
	}
	if err != nil {
		m.Err = err
	}
	return m, nil
}

// View renders the pagination to a string.
func (m Model) View() string {
	return m.render()
}

var (
	errForegroundColor = lipgloss.AdaptiveColor{ Light: "#E03F3F", Dark: "#F45B5B" }
	errBackgroundColor = lipgloss.AdaptiveColor{ Light: "#212121", Dark: "#4A4A4A" }
	errStyle = lipgloss.Style{}.Foreground(errForegroundColor).Background(errBackgroundColor)

	debugForegroundColor = lipgloss.AdaptiveColor{ Light: "#FFA348", Dark: "#FFBE6F" }
	debugBackgroundColor = lipgloss.AdaptiveColor{ Light: "#212121", Dark: "#4A4A4A" }
	debugStyle = lipgloss.Style{}.Foreground(debugForegroundColor).Background(debugBackgroundColor)
)

func (m Model) render() string {
	bot := clamp(m.cur.top+m.cur.h, m.cur.h, len(m.tree))
	top := clamp(m.cur.top, 0, max(len(m.tree)-m.cur.h, m.cur.h))
	if Debug {
		DebugLines = append(DebugLines, debugStyle.Width(m.cur.w).Render(fmt.Sprintf("w:h %d:%d", m.cur.w, m.cur.h)))
		DebugLines = append(DebugLines, debugStyle.Width(m.cur.w).Render(fmt.Sprintf("t:b %d:%d", top, bot)))
		DebugLines = append(DebugLines, debugStyle.Width(m.cur.w).Render(fmt.Sprintf("lines: %d", len(m.tree))))
		if len(m.tree) > bot {
			bot -= len(DebugLines)
		}
	}
	var lines []string
	if len(m.tree) > bot {
		lines = m.tree[top:bot]
	}
	if m.Err != nil {
		err := []string{errStyle.Width(m.cur.w).Render(m.Err.Error())}
		lines = append(err, lines...)
	}
	if Debug {
		lines = append(DebugLines, lines...)
		DebugLines = DebugLines[:0]
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
