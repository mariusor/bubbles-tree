package tree

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type cursor struct {
	h, w      int
	top, left int
}

// Model is the Bubble Tea model for this user interface.
type Model struct{
	Err    error
	cur    cursor
	top    *os.File
	tree   []string
}

func (m *Model) Prev() error {
	m.cur.top = clamp(m.cur.top-1, 0, len(m.tree)-m.cur.h)
	return nil
}

func (m *Model) Next() error {
	m.cur.top = clamp(m.cur.top+1, 0, len(m.tree)-m.cur.h)
	return nil
}

func (m *Model) Init() tea.Cmd {
	return func() tea.Msg {
		//m.tree = strings.Split(staticTree, "\n")
		m.top, m.Err = os.Open("/tmp")
		fmt.Printf("inited: %s", m.top.Name())
		return nil
	}
}

// Update is the Tea update function which binds keystrokes to pagination.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var err error
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			err = m.Prev()
		case "down", "j":
			err = m.Next()
		case "q", "esc", "ctrl+q", "ctrl+c":
			return m, tea.Quit
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
	if m.Err != nil {
		return m.Err.Error()
	}
	if m.top == nil {
		return "waiting for init message"
	}
	root := m.top.Name()
	filepath.Walk(root, func(p string, fi fs.FileInfo, err error) error {
		if path.Dir(p) == root {
			m.tree = append(m.tree, p)
		}
		return nil
	})
	if m.tree == nil {
		return "waiting for init message"
	}
	bot := clamp(m.cur.top+ m.cur.h, 0, len(m.tree))
	return strings.Join(m.tree[m.cur.top:bot], "\n") + fmt.Sprintf("\ntop:h %d:%d", m.cur.top, bot)
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
