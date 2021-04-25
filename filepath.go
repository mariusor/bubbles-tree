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

func (p Path) Walk(cnt int) ([]string, error) {
	all := make([]string, 0)
	err := filepath.Walk(string(p), func(file string, fi fs.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		all = append(all, file)
		return nil
	})
	return all, err
}
