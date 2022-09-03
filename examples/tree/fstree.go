package main

import (
	tree "github.com/mariusor/bubbles-tree"
	"io/fs"
	"log"
)

type dirFs struct {
	fs.FS
}

// Advance moves the Treeish to a new received path,
// this can return a new Treeish instance at the new path, or perform some other function
// for the cases where the path doesn't correspond to a Treeish object.
// Specifically in the case of the filepath Treeish:
// If a passed path parameter corresponds to a folder, it will return a new Treeish object at the new path
// If the passed path parameter corresponds to a file, it returns a nil Treeish, but it can execute something else.
// Eg, When being passed a path that corresponds to a text file, another bubbletea function corresponding to a
// viewer can be called from here.
func (d *dirFs) Advance(path string) (tree.Treeish, error) {
	fi, err := fs.Stat(d, path)
	if err == nil && fi.IsDir() {
		s, err := fs.Sub(d, path)
		log.Printf("subdir: %s", s)
		return &dirFs{s}, err
	}
	return d, nil
}

// State returns the NodeState for the received path parameter
// This is used when rendering the path in the tree view
func (d *dirFs) State(file string) (tree.NodeState, error) {
	f, err := fs.Stat(d, file)
	if err != nil {
		return tree.NodeError, err
	}
	var state tree.NodeState
	if f.IsDir() {
		state |= tree.NodeCollapsible
	}
	return state, nil
}

// Walk loads the elements of current dirFs and returns them as a flat list
func (d *dirFs) Walk(max int) ([]string, error) {
	all := make([]string, 0)

	children, err := fs.ReadDir(d, ".")
	if err != nil {
		return nil, err
	}

	for _, dir := range children {
		all = append(all, dir.Name())
	}
	log.Printf("read %d files", len(all))
	return all, nil
}
