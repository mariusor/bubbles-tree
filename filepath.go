package tree

import (
	"io/fs"
	"path/filepath"
)

type Path string

func (p Path) Walk() ([]string, error) {
	all := make([]string, 0)
	err := filepath.Walk(string(p), func(p string, fi fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		//if p != root && fi.IsDir() {
		//	return fs.SkipDir
		//}
		all = append(all, p)
		return nil
	})
	return all, err
}
