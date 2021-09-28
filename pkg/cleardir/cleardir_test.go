package cleardir_test

import (
	"path"

	"github.com/echocrow/fsnap/dirsnap"
)

type fsd = dirsnap.Dirs

func joinBaseDir(root string, names []string) []string {
	out := make([]string, len(names))
	for i, p := range names {
		out[i] = path.Join(root, p)
	}
	return out
}
