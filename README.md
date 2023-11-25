# Tree Model for BubbleTea TUI framework

This is a package to be used in applications based on [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI framework.

It can be used to output an interactive tree model.

The nodes of the tree, need to conform to a simple interface.

```go

type Node interface {
	tea.Model
	Parent() Node
	Children() Nodes
	State() NodeState
}

```

## Instalation

```sh
go get github.com/mariusor/bubbles-tree
```

## Examples

To see two of the possible usages for this package check the `examples` folder
where we have a minimal filesystem tree utility and a threaded conversation.

### [Tree](./examples/tree)

```sh
$ go run main.go -depth 3 ../../
 └─ ⊟ /some/path/bubbles-tree
    ├─ ⊟ examples
    │  └─ ⊟ tree
    │     ├─   go.mod
    │     ├─   go.sum
    │     ├─   go.work
    │     └─   main.go
    ├─   go.mod
    ├─   go.sum
    ├─   tree.go
    └─   tree_test.go
```

### Conversation

```sh
 $ go run . -depth 3
╻
┃ Root node
┃ Sphinx of black quartz, judge my vow!
┃ The quick brown fox jumps over the lazy dog.
╹
┃ ╻
┃ ┃ Child node
┃ ┃ Sphinx of black quartz, judge my vow!
┃ ┃ The quick brown fox jumps over the lazy dog.
┃ ╹
┃ ┃ ╻
┃ ┃ ┃ Child node
┃ ┃ ┃ Sphinx of black quartz, judge my vow!
┃ ┃ ┃ The quick brown fox jumps over the lazy dog.
┃ ┃ ╹
```
