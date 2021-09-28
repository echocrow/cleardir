package cleardir

import (
	"io"
	"path/filepath"

	os "github.com/echocrow/osa"
	"github.com/scylladb/go-set/strset"
)

// FindClearables finds files and directories that can be safely deleted.
func FindClearables(
	matches chan<- string,
	dir string,
	trivials []string,
	maxDepth int,
) error {
	_, err := find(matches, strset.New(trivials...), dir, maxDepth)
	return err
}

func find(
	matches chan<- string,
	trivials *strset.Set,
	path string,
	depth int,
) (
	canDel bool,
	err error,
) {
	entries, dirErr := os.ReadDir(path)
	if dirErr != nil && dirErr != io.EOF {
		return false, dirErr
	}

	canDel = true
	for _, e := range entries {
		n := e.Name()
		ep := filepath.Join(path, n)
		del := false
		if !e.IsDir() {
			del = trivials.Has(n)
		} else if depth != 0 {
			del, err = find(matches, trivials, ep, depth-1)
		}
		if err != nil {
			return false, err
		}
		if del {
			matches <- ep
		} else {
			canDel = false
		}
	}

	return
}
