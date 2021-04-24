package tree

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
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
	top    os.FileInfo
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
		m.tree = strings.Split(staticTree, "\n")
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
		fmt.Printf("WxH: %dx%d", m.cur.w, m.cur.h)
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

// SetCursor start moves the cursor to the given position. If the position is
// out of bounds the cursor will be moved to the start or end accordingly.
// Returns whether or nor the cursor timer should be reset.
func (m *Model) SetCursor(pos int) {
	m.cur.top = clamp(pos, 0, len(m.tree))
}

const staticTree = `
Documents/
├── 70-smartcard.rules
├── 70-u2f.rules
├── 70-wifi-powersave.rules
├── AP.postman_collection.json
├── bugs
│   ├── blackmesa
│   │   ├── crash.png
│   │   └── dump.txt
│   └── exherbo
│       └── sway
│           └── install.log
├── cv
│   ├── coverletter-soundcloud.txt
│   ├── coverletter-zatoo.txt
│   ├── en
│   │   ├── additional[en].tex
│   │   ├── education[en].tex
│   │   ├── experience[en].tex
│   │   └── skills[en].tex
├── ebooks
│   ├── Delve
│   │   ├── Delve - 001: Woodland.mobi
│   │   ├── Delve - 002: One on One.mobi
│   │   ├── Delve - 003: Pothole.mobi
│   │   ├── Delve - 004: Statistics.mobi
│   │   ├── Delve - 005: Alone.mobi
├── expenses
│   └── 2020-08.md
├── Light_over_Darkness.xcf
├── My Games
│   └── The Vanishing of Ethan Carter
│       ├── AstronautsGame
│       │   ├── Cloud
│       │   │   └── CloudStorage.ini
│       │   ├── Config
│       │   │   ├── AstronautsConfig.inb
│       │   │   ├── AstronautsEngine.ini
│       │   │   ├── AstronautsGame.ini
│       │   │   ├── AstronautsInput.ini
│       │   │   ├── AstronautsLightmass.ini
│       │   │   ├── AstronautsSystemSettings.ini
│       │   │   └── AstronautsUI.ini
│       │   ├── Logs
│       │   │   ├── Launch-backup-2017.09.13-20.14.42.log
│       │   │   └── Launch.log
│       │   └── SaveData
│       │       └── SaveGame_0.sav
│       ├── Binaries
│       │   └── Win64
│       └── Engine
│           └── Config
│               └── ConsoleVariables.ini
├── ragel-guide-6.9.pdf
├── rules-for-polite-online-discourse.md
├── stories
│   ├── activitypub-c2s.md
│   ├── bulk_amorphous_metal.pdf
│   ├── butterflies_from_leonard.md
│   ├── coverlet.md
│   ├── gs
│   │   ├── chapter-eight
│   │   │   └── main.md
│   │   ├── chapter-eighteen
│   │   │   └── main.md
│   │   ├── chapter-eleven
│   │   │   └── main.md
│   │   ├── chapter-fifteen
│   │   │   └── main.md
│   │   ├── chapter-five
│   │   │   └── main.md
│   │   ├── chapter-four
│   │   │   ├── journey-start.md
│   │   │   └── main.md
│   │   ├── chapter-fourteen
│   │   │   └── main.md
│   │   ├── chapter-nine
│   │   │   └── main.md
│   │   ├── chapter-nineteen
│   │   │   └── main.md
│   │   ├── chapter-one
│   │   │   ├── dream.md
│   │   │   ├── main.md
│   │   │   ├── village-description.md
│   │   │   └── wake-up.md
│   │   ├── chapter-seven
│   │   │   └── main.md
│   │   ├── chapter-seventeen
│   │   │   └── main.md
│   │   ├── chapter-six
│   │   │   └── main.md
│   │   ├── chapter-sixteen
│   │   │   └── main.md
│   │   ├── chapter-ten
│   │   │   └── main.md
│   │   ├── chapter-thirteen
│   │   │   └── main.md
│   │   ├── chapter-three
│   │   │   └── main.md
│   │   ├── chapter-twelve
│   │   │   └── main.md
│   │   ├── chapter-two
│   │   │   └── main.md
│   │   ├── chapter-zero
│   │   │   ├── a-new-again.md
│   │   │   ├── commet.md
│   │   │   ├── disaster.md
│   │   │   ├── dream.md
│   │   │   ├── edit.md
│   │   │   ├── empty-valley.md
│   │   │   ├── i-am-not-really-here.md
│   │   │   ├── main.md
│   │   │   ├── memory_1.md
│   │   │   ├── old-man.md
│   │   │   ├── schema.md
│   │   │   └── snow.md
│   │   ├── excerpts
│   │   │   ├── below.md
│   │   │   ├── butterflies.md
│   │   │   ├── elk.md
│   │   │   ├── fairy-songs.md
│   │   │   ├── mother.md
│   │   │   ├── nordic-funeral-inscription.md
│   │   │   ├── one-hand.md
│   │   │   ├── strangers.md
│   │   │   ├── the-big-fizz.md
│   │   │   ├── the-wave.md
│   │   │   └── traveller.md
│   │   ├── external-references
│   │   │   ├── archaic-words.md
│   │   │   ├── arctic-dreams.md
│   │   │   ├── characters
│   │   │   │   ├── dogs.md
│   │   │   │   ├── families.md
│   │   │   │   ├── father.md
│   │   │   │   ├── guul
│   │   │   │   │   └── story.md
│   │   │   │   ├── jarl
│   │   │   │   │   └── dialogue-jarl.md
│   │   │   │   ├── little-brother.md
│   │   │   │   ├── main-character
│   │   │   │   │   ├── backstory-from-father.md
│   │   │   │   │   └── options.md
│   │   │   │   ├── main.md
│   │   │   │   └── mother.md
│   │   │   ├── main.md
│   │   │   ├── plagues.md
│   │   │   ├── religion
│   │   │   │   └── main.md
│   │   │   ├── setting
│   │   │   │   └── main.md
│   │   │   ├── snippets.md
│   │   │   └── snow-racing.md
│   │   ├── full.md
│   │   ├── gss.msk
│   │   ├── LICENSE
│   │   ├── meta.yml
│   │   └── toc.md
│   ├── hole.md
│   ├── LiquidMetal_01.pdf
│   ├── spk
│   │   ├── excerpts
│   │   │   ├── begining.md
│   │   │   └── snow.md
│   │   ├── external-references
│   │   │   ├── main.md
│   │   │   └── snippets.md
│   │   ├── LICENSE
│   │   ├── meta.yml
│   │   └── toc.md
│   ├── the-missing.md
│   └── the-unlikeness.md
├── tax.md
├── tuxedo-laptop-20200718.pdf
└── work
`

func (m Model) render() string {
	if m.Err != nil {
		return m.Err.Error()
	}
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
