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

func parent(p *n) func(*n) {
	return func(nn *n) {
		nn.p = p
	}
}
func state(st NodeState) func(*n) {
	return func(nn *n) {
		nn.s = st
	}
}
func children(c ...*n) func(*n) {
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
	n := &n{n: name}
	for _, fn := range fns {
		fn(n)
	}
	return n
}

var treeOne = tn("tmp",
	state(NodeLastChild|NodeSingleChild),
	children(
		tn("example1"),
		tn("test",
			children(
				tn("example",
					children(
						tn("file2"),
						tn("file4"),
						tn("lastchild", children(tn("file"))),
					),
				),
				tn("file1"),
				tn("file3"),
				tn("file5"),
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
