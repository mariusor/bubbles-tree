package tree

import tea "github.com/charmbracelet/bubbletea"

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

func OnNodes(nodes Nodes, fn func(n *Node) error) {
	for _, node := range nodes {
		if err := fn(&node); err != nil {
			//
		}
	}
}

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
			j += len(p.Children().visibleNodes())
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

func (n Nodes) visibleNodes() Nodes {
	visible := make(Nodes, 0)
	for _, nn := range n {
		if nn == nil || isHidden(nn) {
			continue
		}
		visible = append(visible, nn)
		if isCollapsible(nn) && isExpanded(nn) {
			visible = append(visible, nn.Children().visibleNodes()...)
		}
	}
	return visible
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
