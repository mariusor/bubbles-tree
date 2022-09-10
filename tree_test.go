package tree

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type n struct {
	n string
	p *n
	c []*n
	s NodeState
}

func (n *n) Parent() Node {
	if n.p == nil {
		return nil
	}
	return n.p
}
func (n *n) Init() tea.Cmd {
	return nil
}
func (n *n) View() string {
	return n.n
}
func (n *n) Children() Nodes {
	nodes := make(Nodes, len(n.c))
	for i, nn := range n.c {
		nodes[i] = nn
	}
	return nodes
}

func (n *n) State() NodeState {
	return n.s
}

func (n *n) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if st, ok := msg.(NodeState); ok {
		n.s = st
	}
	return n, nil
}

func p(p *n) func(*n) {
	return func(nn *n) {
		nn.p = p
	}
}

func st(st NodeState) func(*n) {
	return func(nn *n) {
		nn.s = st
	}
}

func c(c ...*n) func(*n) {
	return func(nn *n) {
		for i, nnn := range c {
			if i == len(c)-1 {
				nnn.s |= NodeLastChild
			}
			nnn.p = nn
			nn.c = append(nn.c, nnn)
		}
	}
}
func tn(name string, fns ...func(*n)) *n {
	n := &n{n: name}
	for _, fn := range fns {
		fn(n)
	}
	if len(n.c) > 0 {
		n.s |= NodeCollapsible
	}
	return n
}

// We're building this mock tree:
// 0  1  2  3  4  <- these are the positions for tree Symbols
// └─ tmp
//    ├─ example1
//    └─ test
//       ├─ example
//       │  ├─ file2
//       │  ├─ file4
//       │  └─ lastchild
//       │     └─ file
//       ├─ file1
//       ├─ file3
//       └─ file5
//
// Generated by the following code:
//
// m := New(Nodes{treeOne})
// m.SetWidth(26)
// m.SetHeight(12)
// m.render()

var treeOne = tn("tmp",
	st(NodeLastChild),
	c(
		tn("example1"),
		tn("test",
			c(
				tn("example",
					c(
						tn("file2"),
						tn("file4"),
						tn("lastchild", st(NodeLastChild), c(tn("file", st(NodeLastChild)))),
					),
				),
				tn("file1"),
				tn("file3"),
				tn("file5", st(NodeLastChild)),
			),
		),
	),
)

func Test_showTreeSymbolAtPos(t *testing.T) {
	type args struct {
		n     Node
		depth int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "nil",
			args: args{},
			want: false,
		},
		{
			name: "no parent - position 0",
			args: args{tn("tmp"), 0},
			want: true,
		},
		{
			// this is irrelevant, as depth is greater than the actual tree depth
			name: "no parent - position 10",
			args: args{tn("tmp"), 10},
			want: false,
		},
		{
			name: "/tmp - position 0",
			args: args{treeOne, 0},
			want: true,
		},
		{
			// this is irrelevant, as depth is greater than the actual tree depth
			name: "/tmp - position 1",
			args: args{treeOne, 1},
			want: false,
		},
		{
			name: "/tmp/test - position 0",
			args: args{treeOne.c[1], 0},
			want: false,
		},
		{
			name: "/tmp/test - position 1",
			args: args{treeOne.c[1], 1},
			want: true,
		},
		{
			name: "/tmp/test/example - position 0",
			args: args{treeOne.c[1].c[0], 0},
			want: false,
		},
		{
			name: "/tmp/test/example - position 1",
			args: args{treeOne.c[1].c[0], 1},
			want: false,
		},
		{
			name: "/tmp/test/example - position 2",
			args: args{treeOne.c[1].c[0], 2},
			want: true,
		},
		{
			name: "/tmp/test/example/file2 - position 0",
			args: args{treeOne.c[1].c[0].c[0], 0},
			want: false,
		},
		{
			name: "/tmp/test/example/file2 - position 1",
			args: args{treeOne.c[1].c[0].c[0], 1},
			want: false,
		},
		{
			name: "/tmp/test/example/file2 - position 2",
			args: args{treeOne.c[1].c[0].c[0], 2},
			want: true,
		},
		{
			name: "/tmp/test/example/file2 - position 3",
			args: args{treeOne.c[1].c[0].c[0], 3},
			want: true,
		},

		{
			name: "/tmp/test/example/lastchild/file - position 0",
			args: args{treeOne.c[1].c[0].c[2].c[0], 0},
			want: false,
		},
		{
			name: "/tmp/test/example/lastchild/file - position 1",
			args: args{treeOne.c[1].c[0].c[2].c[0], 1},
			want: false,
		},
		{
			name: "/tmp/test/example/lastchild/file - position 2",
			args: args{treeOne.c[1].c[0].c[2].c[0], 2},
			want: true,
		},
		{
			name: "/tmp/test/example/lastchild/file - position 3",
			args: args{treeOne.c[1].c[0].c[2].c[0], 3},
			want: false,
		},
		{
			name: "/tmp/test/example/lastchild/file - position 4",
			args: args{treeOne.c[1].c[0].c[2].c[0], 4},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := showTreeSymbolAtPos(tt.args.n, tt.args.depth, getDepth(tt.args.n)); got != tt.want {
				t.Errorf("showTreeSymbolAtPos() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getDepth(t *testing.T) {
	type args struct {
		n Node
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "nil",
			args: args{},
			want: 0,
		},
		{
			name: "/tmp",
			args: args{treeOne},
			want: 0,
		},
		{
			name: "/tmp/example1",
			args: args{treeOne.c[0]},
			want: 1,
		},
		{
			name: "/tmp/test",
			args: args{treeOne.c[1]},
			want: 1,
		},
		{
			name: "/tmp/test/example",
			args: args{treeOne.c[1].c[0]},
			want: 2,
		},
		{
			name: "/tmp/test/example/file2",
			args: args{treeOne.c[1].c[0].c[0]},
			want: 3,
		},
		{
			name: "/tmp/test/example/file4",
			args: args{treeOne.c[1].c[0].c[1]},
			want: 3,
		},
		{
			name: "/tmp/test/example/lastchild",
			args: args{treeOne.c[1].c[0].c[2]},
			want: 3,
		},
		{
			name: "/tmp/test/example/lastchild/file",
			args: args{treeOne.c[1].c[0].c[2].c[0]},
			want: 4,
		},
		{
			name: "/tmp/test/file1",
			args: args{treeOne.c[1].c[1]},
			want: 2,
		},
		{
			name: "/tmp/test/file3",
			args: args{treeOne.c[1].c[2]},
			want: 2,
		},
		{
			name: "/tmp/test/file5",
			args: args{treeOne.c[1].c[3]},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getDepth(tt.args.n); got != tt.want {
				t.Errorf("getDepth() = %v, want %v", got, tt.want)
			}
		})
	}
}

var emptyPadding = DefaultSymbols().Padding()
var vertical = DefaultSymbols().Vertical.draw(DefaultSymbols().Width)
var verticalAndRight = DefaultSymbols().VerticalAndRight.draw(DefaultSymbols().Width)
var upAndRight = DefaultSymbols().UpAndRight.draw(DefaultSymbols().Width)
var squaredPlus = DefaultSymbols().Collapsed
var squaredMinus = DefaultSymbols().Expanded

func Test_getTreeSymbolForPos(t *testing.T) {
	type args struct {
		n   Node
		pos int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "nil",
			args: args{nil, 0},
			want: "",
		},
		{
			name: "/tmp",
			args: args{treeOne, 0},
			want: upAndRight,
		},
		// /tmp/example1
		{
			name: "/tmp/example1 - pos 0",
			args: args{treeOne.c[0], 0},
			want: emptyPadding,
		},
		{
			name: "/tmp/example1 - pos 1",
			args: args{treeOne.c[0], 1},
			want: verticalAndRight,
		},
		// /tmp/test
		{
			name: "/tmp/test - pos 0",
			args: args{treeOne.c[1], 0},
			want: emptyPadding,
		},
		{
			name: "/tmp/test - pos 1",
			args: args{treeOne.c[1], 1},
			want: upAndRight,
		},
		// /tmp/test/example
		{
			name: "/tmp/test/example - pos 0",
			args: args{treeOne.c[1].c[0], 0},
			want: emptyPadding,
		},
		{
			name: "/tmp/test/example - pos 1",
			args: args{treeOne.c[1].c[0], 1},
			want: emptyPadding,
		},
		{
			name: "/tmp/test/example - pos 2",
			args: args{treeOne.c[1].c[0], 2},
			want: verticalAndRight,
		},
		// /tmp/text/example/file2
		{
			name: "/tmp/test/example/file2 - pos 0",
			args: args{treeOne.c[1].c[0].c[0], 0},
			want: emptyPadding,
		},
		{
			name: "/tmp/test/example/file2 - pos 1",
			args: args{treeOne.c[1].c[0].c[0], 1},
			want: emptyPadding,
		},
		{
			name: "/tmp/test/example/file2 - pos 2",
			args: args{treeOne.c[1].c[0].c[0], 2},
			want: vertical,
		},
		// /tmp/test/example/file4
		{
			name: "/tmp/test/example/file4 - pos 0",
			args: args{treeOne.c[1].c[0].c[1], 0},
			want: emptyPadding,
		},
		{
			name: "/tmp/test/example/file4 - pos 1",
			args: args{treeOne.c[1].c[0].c[1], 1},
			want: emptyPadding,
		},
		{
			name: "/tmp/test/example/file4 - pos 2",
			args: args{treeOne.c[1].c[0].c[1], 2},
			want: vertical,
		},
		// /tmp/test/example/lastchild
		{
			name: "/tmp/test/example/lastchild - pos 0",
			args: args{treeOne.c[1].c[0].c[2], 0},
			want: emptyPadding,
		},
		{
			name: "/tmp/test/example/lastchild - pos 1",
			args: args{treeOne.c[1].c[0].c[2], 1},
			want: emptyPadding,
		},
		{
			name: "/tmp/test/example/lastchild - pos 2",
			args: args{treeOne.c[1].c[0].c[2], 2},
			want: vertical,
		},
		{
			name: "/tmp/test/example/lastchild - pos 3",
			args: args{treeOne.c[1].c[0].c[2], 3},
			want: upAndRight,
		},
		// /tmp/test/example/lastchild/file
		{
			name: "/tmp/test/example/lastchild/file - pos 0",
			args: args{treeOne.c[1].c[0].c[2].c[0], 0},
			want: emptyPadding,
		},
		{
			name: "/tmp/test/example/lastchild/file - pos 1",
			args: args{treeOne.c[1].c[0].c[2].c[0], 1},
			want: emptyPadding,
		},
		{
			name: "/tmp/test/example/lastchild/file - pos 2",
			args: args{treeOne.c[1].c[0].c[2].c[0], 2},
			want: vertical,
		},
		{
			name: "/tmp/test/example/lastchild/file - pos 3",
			args: args{treeOne.c[1].c[0].c[2].c[0], 3},
			want: emptyPadding,
		},
		{
			name: "/tmp/test/example/lastchild/file - pos 4",
			args: args{treeOne.c[1].c[0].c[2].c[0], 4},
			want: upAndRight,
		},
		// /tmp/test/file1
		{
			name: "/tmp/test/file1 - pos 0",
			args: args{treeOne.c[1].c[1], 0},
			want: emptyPadding,
		},
		{
			name: "/tmp/test/file1 - pos 1",
			args: args{treeOne.c[1].c[1], 1},
			want: emptyPadding,
		},
		{
			name: "/tmp/test/file1 - pos 2",
			args: args{treeOne.c[1].c[1], 2},
			want: verticalAndRight,
		},
		// /tmp/test/file3
		{
			name: "/tmp/test/file3 - pos 0",
			args: args{treeOne.c[1].c[2], 0},
			want: emptyPadding,
		},
		{
			name: "/tmp/test/file3 - pos 1",
			args: args{treeOne.c[1].c[2], 1},
			want: emptyPadding,
		},
		{
			name: "/tmp/test/file3 - pos 2",
			args: args{treeOne.c[1].c[2], 2},
			want: verticalAndRight,
		},
		// /tmp/test/file5
		{
			name: "/tmp/test/file5 - pos 0",
			args: args{treeOne.c[1].c[3], 0},
			want: emptyPadding,
		},
		{
			name: "/tmp/test/file5 - pos 1",
			args: args{treeOne.c[1].c[3], 1},
			want: emptyPadding,
		},
		{
			name: "/tmp/test/file5 - pos 2",
			args: args{treeOne.c[1].c[3], 2},
			want: upAndRight,
		},
	}
	m := mockModel()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			maxDepth := getDepth(tt.args.n)
			if got := m.getTreeSymbolForPos(tt.args.n, tt.args.pos, maxDepth); got != tt.want {
				t.Errorf("getTreeSymbolForPos() = %v, want %v", got, tt.want)
			}
		})
	}
}

var ellipsis = DefaultSymbols().Ellipsis

func Test_ellipsize(t *testing.T) {
	type args struct {
		s string
		w int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty",
			args: args{},
			want: "",
		},
		{
			name: "not ellipsized",
			args: args{"ana are mere", 20},
			want: "ana are mere",
		},
		{
			name: "ellipsized",
			args: args{"ana are mere", 10},
			want: "ana are m" + ellipsis,
		},
	}
	m := mockModel()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := m.ellipsize(tt.args.s, tt.args.w); got != tt.want {
				t.Errorf("ellipsize() = %q, want %q", got, tt.want)
			}
		})
	}
}

func Test_clamp(t *testing.T) {
	type args struct {
		v    int
		low  int
		high int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "all zeroes - want zero",
			args: args{0, 0, 0},
			want: 0,
		},
		{
			name: "min/max are zeroes - want zero",
			args: args{100, 0, 0},
			want: 0,
		},
		{
			name: "give greater than max - want max",
			args: args{101, 0, 100},
			want: 100,
		},
		{
			name: "give lower than min - want min",
			args: args{-101, 0, 100},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clamp(tt.args.v, tt.args.low, tt.args.high); got != tt.want {
				t.Errorf("clamp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_min(t *testing.T) {
	type args struct {
		a int
		b int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "equal values",
			args: args{a: 0, b: 0},
			want: 0,
		},
		{
			name: "min first",
			args: args{a: 0, b: 10},
			want: 0,
		},
		{
			name: "min last",
			args: args{a: 1000, b: -10},
			want: -10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := min(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("min() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_max(t *testing.T) {
	type args struct {
		a int
		b int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "equal values",
			args: args{a: 0, b: 0},
			want: 0,
		},
		{
			name: "max first",
			args: args{a: 10, b: 0},
			want: 10,
		},
		{
			name: "max last",
			args: args{a: -10, b: 1000},
			want: 1000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := max(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("max() = %v, want %v", got, tt.want)
			}
		})
	}
}

var child = tn("two")
var oneWithChild = tn("one", c(child))
var oneWithChildExpected = Nodes{oneWithChild, child}

var hiddenChild = tn("two", st(NodeHidden))
var oneWithHiddenChild = tn("one", c(hiddenChild))
var oneWithHiddenChildExpected = Nodes{oneWithHiddenChild}

var oneWithChildCollapsed = tn("one collapsed", st(NodeCollapsed), c(child))
var oneWithChildCollapsedExpected = Nodes{oneWithChildCollapsed}

func TestNodes_visibleNodes(t *testing.T) {
	tests := []struct {
		name string
		n    Nodes
		want Nodes
	}{
		{
			name: "empty",
			n:    Nodes{},
			want: Nodes{},
		},
		{
			name: "single node",
			n:    Nodes{tn("one")},
			want: Nodes{tn("one")},
		},
		{
			name: "two nodes",
			n:    Nodes{tn("one"), tn("two")},
			want: Nodes{tn("one"), tn("two")},
		},
		{
			name: "one node with visible child",
			n:    Nodes{oneWithChild},
			want: oneWithChildExpected,
		},
		{
			name: "one node with non visible child",
			n:    Nodes{oneWithHiddenChild},
			want: oneWithHiddenChildExpected,
		},
		{
			name: "one collapsed with visible child",
			n:    Nodes{oneWithChildCollapsed},
			want: oneWithChildCollapsedExpected,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.n.visibleNodes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("visibleNodes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodes_at(t *testing.T) {
	type args struct {
		i int
	}
	tests := []struct {
		name string
		n    Nodes
		args args
		want Node
	}{
		{
			name: "empty",
			n:    nil,
			args: args{0},
			want: nil,
		},
		{
			name: "empty: invalid index",
			n:    nil,
			args: args{1},
			want: nil,
		},
		{
			name: "first from one node",
			n:    Nodes{tn("one")},
			args: args{0},
			want: tn("one"),
		},
		{
			name: "invalid index from one node",
			n:    Nodes{tn("one")},
			args: args{1},
			want: nil,
		},
		{
			name: "second from two nodes",
			n:    Nodes{tn("one"), tn("two")},
			args: args{1},
			want: tn("two"),
		},
		{
			name: "second from node with child",
			n:    Nodes{tn("one", c(child))},
			args: args{1},
			want: child,
		},
		{
			name: "nil when getting hidden child",
			n:    Nodes{oneWithHiddenChild},
			args: args{1},
			want: nil,
		},
		{
			name: "parent when getting from collapsed parent",
			n:    Nodes{oneWithChildCollapsed},
			args: args{0},
			want: oneWithChildCollapsed,
		},
		{
			name: "nil when getting from collapsed parent",
			n:    Nodes{oneWithChildCollapsed},
			args: args{1},
			want: nil,
		},
		{
			name: "treeOne - pos 0",
			n:    Nodes{treeOne},
			args: args{0},
			want: treeOne,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.n.at(tt.args.i); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("at() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodes_len(t *testing.T) {
	tests := []struct {
		name string
		n    Nodes
		want int
	}{
		{
			name: "nil",
			n:    nil,
			want: 0,
		},
		{
			name: "empty",
			n:    Nodes{},
			want: 0,
		},
		{
			name: "one node",
			n:    Nodes{tn("one")},
			want: 1,
		},
		{
			name: "two nodes",
			n:    Nodes{tn("one"), tn("two")},
			want: 2,
		},
		{
			name: "one with one child",
			n:    Nodes{tn("one", c(tn("two")))},
			want: 2,
		},
		{
			name: "treeOne",
			n:    Nodes{treeOne},
			want: 11,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.n.len(); got != tt.want {
				t.Errorf("len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func mockModel(nn ...*n) Model {
	m := Model{
		viewport: viewport.New(0, 1),
		focus:    true,
		KeyMap:   DefaultKeyMap(),
		Styles:   DefaultStyles(),
		Symbols:  DefaultSymbols(),
	}
	if len(nn) == 0 {
		return m
	}
	m.tree = make(Nodes, len(nn))
	for i, n := range nn {
		m.tree[i] = n
	}
	return m
}

func TestNew(t *testing.T) {
	type args struct {
		t Nodes
	}
	tests := []struct {
		name string
		args args
		want Model
	}{
		{
			name: "no nodes",
			args: args{},
			want: mockModel(),
		},
		{
			name: "single node",
			args: args{Nodes{tn("test")}},
			want: mockModel(tn("test")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.t); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModel_Children(t *testing.T) {
	type fields struct {
		tree []*n
	}
	tests := []struct {
		name   string
		fields fields
		want   Nodes
	}{
		{
			name:   "nil",
			fields: fields{},
			want:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := mockModel(tt.fields.tree...)
			if got := m.Children(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Children() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModel_SetWidth(t *testing.T) {
	testValues := []int{
		0, 10, -100, 200,
	}
	for _, w := range testValues {
		t.Run(fmt.Sprintf("Width: %d", w), func(t *testing.T) {
			m := mockModel()
			if m.viewport.Width != 0 {
				t.Errorf("invalid width after initialization: %d, expected %d", m.viewport.Width, 0)
			}
			m.SetWidth(w)
			if m.viewport.Width != w {
				t.Errorf("invalid width after SetWidth(): %d, expected %d", m.viewport.Width, w)
			}
			if m.Width() != w {
				t.Errorf("invalid value returned by Width(): %d, expected %d", m.Width(), w)
			}
		})
	}
}

func TestModel_SetHeight(t *testing.T) {
	testValues := []int{
		0, 10, -100, 200, 666,
	}
	for _, w := range testValues {
		t.Run(fmt.Sprintf("Height: %d", w), func(t *testing.T) {
			m := mockModel()
			if m.viewport.Height != 1 {
				t.Errorf("invalid height after initialization: %d, expected %d", m.viewport.Height, 1)
			}
			m.SetHeight(w)
			if m.viewport.Height != w {
				t.Errorf("invalid width after SetHeight(): %d, expected %d", m.viewport.Height, w)
			}
			if m.Height() != w {
				t.Errorf("invalid value returned by Height(): %d, expected %d", m.Height(), w)
			}
		})
	}
}

func TestModel_Init(t *testing.T) {
	// NOTE(marius): having the init() function as a Cmd seems iffy and maybe pointless
	m := mockModel()
	want := m.init
	got := m.Init()
	if reflect.TypeOf(want).Kind() != reflect.Func {
		t.Errorf("Init() did not return a function")
	}
	if !reflect.DeepEqual(got(), want()) {
		t.Errorf("Init() = %v, want %v", got(), want())
	}
}

func TestModel_Focus(t *testing.T) {
	m := mockModel()
	if !m.focus {
		t.Errorf("invalid focus value after initialization: %t, expected %t", m.focus, true)
	}
	if !m.Focused() {
		t.Errorf("invalid Focused() value after initialization: %t, expected %t", m.Focused(), true)
	}
	m.focus = false
	m.Focus()
	if !m.focus {
		t.Errorf("invalid focus value after calling Focus(): %t, expected %t", m.focus, true)
	}
	if !m.Focused() {
		t.Errorf("invalid Focused() value after calling Focus(): %t, expected %t", m.Focused(), true)
	}
}

func TestModel_Blur(t *testing.T) {
	m := mockModel()
	if !m.focus {
		t.Errorf("invalid focus value after initialization: %t, expected %t", m.focus, true)
	}
	if !m.Focused() {
		t.Errorf("invalid Focused() value after initialization: %t, expected %t", m.Focused(), true)
	}
	m.Blur()
	if m.focus {
		t.Errorf("invalid focus value after calling Blur(): %t, expected %t", m.focus, false)
	}
	if m.Focused() {
		t.Errorf("invalid Focused() value after calling Blur(): %t, expected %t", m.Focused(), false)
	}
}

func TestModel_View(t *testing.T) {
	tests := []struct {
		name     string
		viewport viewport.Model
		want     string
	}{
		{
			name:     "empty",
			viewport: viewport.Model{},
			want:     (viewport.Model{}).View(),
		},
		{
			name:     "1x1",
			viewport: viewport.Model{Width: 1, Height: 1},
			want:     (viewport.Model{Width: 1, Height: 1}).View(),
		},
		{
			name:     "1x1 - using selectedStyle",
			viewport: viewport.Model{Width: 1, Height: 1, Style: defaultSelectedStyle},
			want:     (viewport.Model{Width: 1, Height: 1, Style: defaultSelectedStyle}).View(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := mockModel()
			m.viewport = tt.viewport
			if got := m.View(); got != tt.want {
				t.Errorf("View() = %v, want %v", got, tt.want)
			}
		})
	}
}

var treeOneRendered = ` └─ tmp               
    ├─ example1       
    └─ test           
       ├─ example     
       │  ├─ file2    
       │  ├─ file4    
       │  └─ lastchild
       │     └─ file  
       ├─ file1       
       ├─ file3       
       └─ file5       `

func TestModel_renderNode(t *testing.T) {
	tests := []struct {
		name string
		node Node
		want string
	}{
		{
			name: "empty",
			node: nil,
			want: "",
		},
		{
			name: "single node",
			node: tn("test", st(NodeLastChild)),
			want: upAndRight + " test",
		},
		{
			name: "single node with child collapsed",
			node: tn("one", st(NodeLastChild|NodeCollapsed), c(tn("two"))),
			want: upAndRight + " one",
		},
		{
			name: "single node with child",
			node: tn("one", st(NodeLastChild), c(tn("two"))),
			want: upAndRight + " one   \n" +
				"   " + upAndRight + " two",
		},
		{
			name: "treeOne",
			node: treeOne,
			want: treeOneRendered,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := mockModel()
			m.tree = Nodes{tt.node}

			got := m.renderNode(tt.node)
			linesGot := strings.Split(got, "\n")
			linesWant := strings.Split(tt.want, "\n")
			for i, lw := range linesWant {
				if lg := linesGot[i]; lw != lg {
					t.Errorf("%2d %s| %s", i, lw, lg)
				}
			}
		})
	}
}

func TestNodeState_Is(t *testing.T) {
	type args struct {
		st NodeState
	}
	tests := []struct {
		name string
		s    NodeState
		args args
		want bool
	}{
		{
			name: "nil",
			s:    0,
			args: args{},
			want: true,
		},
		{
			name: "Collapsible.Is_Collapsible",
			s:    NodeCollapsible,
			args: args{NodeCollapsible},
			want: true,
		},
		{
			name: "Collapsed.Is_Collapsed",
			s:    NodeCollapsed,
			args: args{NodeCollapsed},
			want: true,
		},
		{
			name: "Collapsed.IsNot_Collapsible",
			s:    NodeCollapsed,
			args: args{NodeCollapsible},
			want: false,
		},
		{
			name: "Collapsed|Collapsible.Is_Collapsible",
			s:    NodeCollapsed | NodeCollapsible,
			args: args{NodeCollapsible},
			want: true,
		},
		{
			name: "Collapsed.IsNot_Collapsible|Collapsed",
			s:    NodeCollapsed,
			args: args{NodeCollapsed | NodeCollapsible},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Is(tt.args.st); got != tt.want {
				t.Errorf("Is() = %v, want %v", got, tt.want)
			}
		})
	}
}
