package tree

import (
	"io/fs"
	"os"
	"path/filepath"
)

type Path string

func (p Path) State(file string) (NodeState, error) {
	f, err := os.Stat(file)
	if err != nil {
		return NodeError, err
	}
	var state NodeState = 0
	if f.IsDir() {
		state |= NodeCollapsible
	}
	return state, nil
}

// Advance returns a new Treeish object based on the new path
func (p Path) Advance(file string) (Treeish, error) {
	fi, err := os.Stat(file)
	if err != nil {
		return nil, err
	}
	if fi.IsDir() {
		return Path(file), nil
	}
	return nil, nil
}

// Walk will load cnt element from the current path
func (p Path) Walk(cnt int) ([]string, error) {
	all := make([]string, 0)
	pp := filepath.Clean(string(p))
	err := filepath.WalkDir(pp, func(file string, fi fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		parent := filepath.Dir(file)
		grandParent := filepath.Dir(parent)
		if file != pp && parent != pp && grandParent != pp {
			return filepath.SkipDir
		}
		if _, err := fi.Info(); err != nil {
			return err
		}
		all = append(all, file)
		return nil
	})
	return all, err
}
