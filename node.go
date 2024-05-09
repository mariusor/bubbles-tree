package tree

import tea "github.com/charmbracelet/bubbletea"

// NodeState is used for passing information from a Treeish element to the view itself
type NodeState uint16

// Node represents the base model for the elements of the Treeish implementation
type Node interface {
	tea.Model
	// Parent should return the parent of the current node, or nil if a root node.
	Parent() Node
	// Children should return a list of Nodes which represent the children of the current node.
	Children() Nodes
	// State should return the annotation for the current node, which are used for computing various display states.
	State() NodeState
}

// Nodes is a slice of Node elements, usually representing the children of a Node.
type Nodes []Node

const (
	NodeNone NodeState = 0

	// NodeCollapsed hints that the current node is collapsed
	NodeCollapsed NodeState = 1 << iota
	// NodeSelected hints that the current node should be rendered as selected
	NodeSelected
	// NodeCollapsible hints that the current node can be collapsed
	NodeCollapsible
	// NodeHidden hints that the current node is not going to be displayed
	NodeHidden
	// NodeLastChild shows the node to be the last in the children list
	NodeLastChild
	// nodeHasPreviousSibling shows if the node has siblings
	nodeHasPreviousSibling
	// NodeIsMultiLine shows if the node should not be truncated to the viewport's max width
	NodeIsMultiLine
	// nodeSkipRender shows if the node will not be rendered
	// NOTE(marius): this might overlap with NodeHidden
	nodeSkipRender

	// NodeMaxState serves no other purpose than as a sentinel value for outside Node interface
	// implementations to append their own states.
	NodeMaxState
)

func (n Nodes) at(i int) Node {
	j := 0
	for _, p := range n {
		if isHidden(p) {
			continue
		}
		if j == i {
			return p
		}
		if isExpanded(p) && p.Children() != nil {
			if nn := p.Children().at(i - j - 1); nn != nil {
				return nn
			}
			j += len(p.Children().sequentialNodes())
		}
		j++
	}
	return nil
}

func (n Nodes) len() int {
	l := 0
	for _, node := range n {
		l++
		if node.Children() != nil {
			l += node.Children().len()
		}
	}
	return l
}

// sequentialNodes returns a slice of Nodes as
func (n Nodes) sequentialNodes() Nodes {
	seq := make(Nodes, 0)
	for _, nn := range n {
		if nn == nil || isHidden(nn) {
			continue
		}
		seq = append(seq, nn)
		if isCollapsible(nn) && isExpanded(nn) {
			seq = append(seq, nn.Children().sequentialNodes()...)
		}
	}
	return seq
}

func (n Nodes) Update(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, 0)
	for i, nn := range n {
		ne, cmd := nn.Update(msg)
		if nnn, ok := ne.(Node); ok {
			nn = nnn
		}
		cmds = append(cmds, cmd)
		if isCollapsible(nn) && isExpanded(nn) {
			cmds = append(cmds, nn.Children().Update(msg))
		}
		n[i] = nn
	}
	return tea.Batch(cmds...)
}

func (s NodeState) Is(st NodeState) bool {
	return s&st == st
}
