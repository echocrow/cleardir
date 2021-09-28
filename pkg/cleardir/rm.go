package cleardir

import (
	os "github.com/echocrow/osa"
)

// Remove removes all listed files and directories.
func Remove(paths ...string) error {
	for _, p := range paths {
		if err := os.Remove(p); err != nil {
			return err
		}
	}
	return nil
}
