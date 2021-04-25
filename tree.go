package tree

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var Debug = true
var DebugLines = make([]string, 0)

type cursor struct {
	h, w      int
	top, left int
}

// Model is the Bubble Tea model for this user interface.
type Model struct{
	Err    error
	Root   string
	cur    cursor
	top    *os.File
	tree   []string
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
	m.Err = filepath.Walk(m.Root, func(p string, fi fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		//if p != root && fi.IsDir() {
		//	return fs.SkipDir
		//}
		m.tree = append(m.tree, p)
		return nil
	})
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

func (m Model) render() string {
	bot := clamp(m.cur.top+m.cur.h, m.cur.h, len(m.tree))
	top := clamp(m.cur.top, 0, max(len(m.tree)-m.cur.h, m.cur.h))
	if Debug {
		DebugLines = append(DebugLines, fmt.Sprintf("w:h %d:%d", m.cur.w, m.cur.h))
		DebugLines = append(DebugLines, fmt.Sprintf("t:b %d:%d", top, bot))
		DebugLines = append(DebugLines, fmt.Sprintf("lines: %d", len(m.tree)))
		if len(m.tree) > bot {
			bot -= len(DebugLines)
		}
	}
	var lines []string
	if len(m.tree) > bot {
		lines = m.tree[top:bot]
	}
	if m.Err != nil {
		lines = append([]string{m.Err.Error()}, lines...)
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
