package tree

import (
	"testing"
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
func (n *n) Name() string {
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

func (n *n) SetState(st NodeState) {
	n.s = st
}

func p(p *n) func(*n) {
	return func(nn *n) {
		nn.p = p
	}
}

func st(st NodeState) func(*n) {
	return func(nn *n) {
		nn.s |= st
	}
}

func c(c ...*n) func(*n) {
	return func(nn *n) {
		if len(c) > 0 {
			nn.s |= NodeCollapsible
		}
		for i, nnn := range c {
			if len(c) == 1 {
				nnn.s |= NodeSingleChild
			}
			if i == len(c)-1 {
				nnn.s |= NodeLastChild
			}
			nnn.p = nn
			nn.c = append(nn.c, nnn)
		}
	}
}
func tn(name string, fns ...func(*n)) *n {
	n := &n{n: name, s: NodeVisible}
	for _, fn := range fns {
		fn(n)
	}
	return n
}

// We're building this mock tree:
// 0  1  2  3  4  <- these are the positions for tree symbols
// └─ ⊟ tmp
//    ├─   example1
//    └─ ⊟ test
//       ├─ ⊟ example
//       │  ├─   file2
//       │  ├─   file4
//       │  └─ ⊟ lastchild
//       │     └─   file
//       ├─   file1
//       ├─   file3
//       └─   file5
//
// Generated by the following code:
//
// m := New(Nodes{treeOne})
// m.SetWidth(26)
// m.SetHeight(12)
// m.render()

var treeOne = tn("tmp",
	st(NodeLastChild|NodeSingleChild),
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
			want: BoxDrawingsUpAndRight,
		},
		// /tmp/example1
		{
			name: "/tmp/example1 - pos 0",
			args: args{treeOne.c[0], 0},
			want: EmptyPadding,
		},
		{
			name: "/tmp/example1 - pos 1",
			args: args{treeOne.c[0], 1},
			want: BoxDrawingsVerticalAndRight,
		},
		// /tmp/test
		{
			name: "/tmp/test - pos 0",
			args: args{treeOne.c[1], 0},
			want: EmptyPadding,
		},
		{
			name: "/tmp/test - pos 1",
			args: args{treeOne.c[1], 1},
			want: BoxDrawingsUpAndRight,
		},
		// /tmp/test/example
		{
			name: "/tmp/test/example - pos 0",
			args: args{treeOne.c[1].c[0], 0},
			want: EmptyPadding,
		},
		{
			name: "/tmp/test/example - pos 1",
			args: args{treeOne.c[1].c[0], 1},
			want: EmptyPadding,
		},
		{
			name: "/tmp/test/example - pos 2",
			args: args{treeOne.c[1].c[0], 2},
			want: BoxDrawingsVerticalAndRight,
		},
		// /tmp/text/example/file2
		{
			name: "/tmp/test/example/file2 - pos 0",
			args: args{treeOne.c[1].c[0].c[0], 0},
			want: EmptyPadding,
		},
		{
			name: "/tmp/test/example/file2 - pos 1",
			args: args{treeOne.c[1].c[0].c[0], 1},
			want: EmptyPadding,
		},
		{
			name: "/tmp/test/example/file2 - pos 2",
			args: args{treeOne.c[1].c[0].c[0], 2},
			want: BoxDrawingsVertical,
		},
		// /tmp/test/example/file4
		{
			name: "/tmp/test/example/file4 - pos 0",
			args: args{treeOne.c[1].c[0].c[1], 0},
			want: EmptyPadding,
		},
		{
			name: "/tmp/test/example/file4 - pos 1",
			args: args{treeOne.c[1].c[0].c[1], 1},
			want: EmptyPadding,
		},
		{
			name: "/tmp/test/example/file4 - pos 2",
			args: args{treeOne.c[1].c[0].c[1], 2},
			want: BoxDrawingsVertical,
		},
		// /tmp/test/example/lastchild
		{
			name: "/tmp/test/example/lastchild - pos 0",
			args: args{treeOne.c[1].c[0].c[2], 0},
			want: EmptyPadding,
		},
		{
			name: "/tmp/test/example/lastchild - pos 1",
			args: args{treeOne.c[1].c[0].c[2], 1},
			want: EmptyPadding,
		},
		{
			name: "/tmp/test/example/lastchild - pos 2",
			args: args{treeOne.c[1].c[0].c[2], 2},
			want: BoxDrawingsVertical,
		},
		{
			name: "/tmp/test/example/lastchild - pos 3",
			args: args{treeOne.c[1].c[0].c[2], 3},
			want: BoxDrawingsUpAndRight,
		},
		// /tmp/test/example/lastchild/file
		{
			name: "/tmp/test/example/lastchild/file - pos 0",
			args: args{treeOne.c[1].c[0].c[2].c[0], 0},
			want: EmptyPadding,
		},
		{
			name: "/tmp/test/example/lastchild/file - pos 1",
			args: args{treeOne.c[1].c[0].c[2].c[0], 1},
			want: EmptyPadding,
		},
		{
			name: "/tmp/test/example/lastchild/file - pos 2",
			args: args{treeOne.c[1].c[0].c[2].c[0], 2},
			want: BoxDrawingsVertical,
		},
		{
			name: "/tmp/test/example/lastchild/file - pos 3",
			args: args{treeOne.c[1].c[0].c[2].c[0], 3},
			want: EmptyPadding,
		},
		{
			name: "/tmp/test/example/lastchild/file - pos 4",
			args: args{treeOne.c[1].c[0].c[2].c[0], 4},
			want: BoxDrawingsUpAndRight,
		},
		// /tmp/test/file1
		{
			name: "/tmp/test/file1 - pos 0",
			args: args{treeOne.c[1].c[1], 0},
			want: EmptyPadding,
		},
		{
			name: "/tmp/test/file1 - pos 1",
			args: args{treeOne.c[1].c[1], 1},
			want: EmptyPadding,
		},
		{
			name: "/tmp/test/file1 - pos 2",
			args: args{treeOne.c[1].c[1], 2},
			want: BoxDrawingsVerticalAndRight,
		},
		// /tmp/test/file3
		{
			name: "/tmp/test/file3 - pos 0",
			args: args{treeOne.c[1].c[2], 0},
			want: EmptyPadding,
		},
		{
			name: "/tmp/test/file3 - pos 1",
			args: args{treeOne.c[1].c[2], 1},
			want: EmptyPadding,
		},
		{
			name: "/tmp/test/file3 - pos 2",
			args: args{treeOne.c[1].c[2], 2},
			want: BoxDrawingsVerticalAndRight,
		},
		// /tmp/test/file5
		{
			name: "/tmp/test/file5 - pos 0",
			args: args{treeOne.c[1].c[3], 0},
			want: EmptyPadding,
		},
		{
			name: "/tmp/test/file5 - pos 1",
			args: args{treeOne.c[1].c[3], 1},
			want: EmptyPadding,
		},
		{
			name: "/tmp/test/file5 - pos 2",
			args: args{treeOne.c[1].c[3], 2},
			want: BoxDrawingsUpAndRight,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			maxDepth := getDepth(tt.args.n)
			if got := getTreeSymbolForPos(tt.args.n, tt.args.pos, maxDepth); got != tt.want {
				t.Errorf("getTreeSymbolForPos() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
			want: "ana are m" + Ellipsis,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ellipsize(tt.args.s, tt.args.w); got != tt.want {
				t.Errorf("ellipsize() = %v, want %v", got, tt.want)
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
