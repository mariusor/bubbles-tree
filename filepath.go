package tree

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
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
func (p Path) Advance(file string) Treeish {
	return Path(file)
}

// Walk will load cnt element from the current path
func (p Path) Walk(cnt int) ([]string, error) {
	all := make([]string, 0)
	pp := filepath.Clean(string(p))
	err := filepath.WalkDir(pp, func(file string, fi fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if file == pp || filepath.Dir(file) == pp {
			all = append(all, file)
		}
		return nil
	})
	sort.Slice(all, func(i, j int) bool {
		f1, _ := os.Stat(all[i])
		if f1 == nil {
			return false
		}
		f2, _ := os.Stat(all[j])
		if f2 == nil {
			return true
		}
		if f1.IsDir() {
			if f2.IsDir() {
				return f1.Name() <= f2.Name()
			}
			return true
		}
		return false
	})
	return all, err
}
