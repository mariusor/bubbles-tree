## Tree

This example uses a very basic filesystem abstraction to display a tree of directories similar to the traditional `tree` unix command.

The model suports out of the box navigating through the tree using the directional keys and also collapsing/expanding directory nodes using `o`.

Different symbols and styles can be configured for the basic elements of the tree.

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
