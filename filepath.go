package tree

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
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

func fileDetails(file string) string {
	fi, _ := os.Lstat(file)

	s := strings.Builder{}
	switch mode := fi.Mode(); {
	case mode.IsRegular():
		fmt.Fprintf(&s, "Regular file %s", file)
	case mode&fs.ModeSymlink != 0:
		fmt.Fprintf(&s, "Symbolic link %s", file)
	case mode&fs.ModeNamedPipe != 0:
		fmt.Fprintf(&s, "Named pipe %s", file)
	}
	fmt.Fprintf(&s, "\nSize: %d bytes\nMode: %o", fi.Size(), fi.Mode().Perm())
	return s.String()
}

// Advance returns a new Treeish object based on the new path
// If the path corresponds to a folder it will return a new
// Path instance based on that.
// If it represents a file it will return an error which
// contains some details about it.
// For a real life implementation the treeish struct must be
// plugged to some function that can use said path when Advance
// is being called.
func (p Path) Advance(path string) (Treeish, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if fi.IsDir() {
		return Path(path), nil
	} else {
		// This is just an example that works with the error reporting
		// in the tree.Model and displays some basic information about
		// the file path under cursor.
		return nil, fmt.Errorf(fileDetails(path))
	}
	return nil, nil
}

// Walk will load cnt element from the current path
func (p Path) Walk(cnt int) ([]string, error) {
	all := make([]string, 0)
	pp := filepath.Clean(string(p))
	err := filepath.WalkDir(pp, func(file string, fi fs.DirEntry, err error) error {
		if err != nil {
			return filepath.SkipDir
		}
		parent := filepath.Dir(file)
		if file != pp && parent != pp {
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
