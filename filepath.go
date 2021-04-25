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

// Walk will load cnt element from the current path
func (p Path) Walk(cnt int) ([]string, error) {
	all := make([]string, 0)
	pp := filepath.Clean(string(p))
	err := filepath.WalkDir(pp, func(file string, fi fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if filepath.Dir(file) == pp {
			//fmt.Fprintf(os.Stderr, "%s == %s : %t\n", filepath.Dir(file), pp, filepath.Dir(file) == pp )
			all = append(all, file)
		} else {
			//fmt.Fprintf(os.Stderr, "skipped %s\n", file)
		}
		return nil
	})
	return all, err
}
