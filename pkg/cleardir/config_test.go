package cleardir_test

import (
	"path"
	"testing"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/echocrow/cleardir/pkg/cleardir"
	"github.com/echocrow/osa"
	"github.com/echocrow/osa/testos"
	"github.com/echocrow/osa/vos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseClearables(t *testing.T) {
	os, reset := vos.Patch()
	defer reset()

	tests := []struct {
		name     string
		contents string
		want     []string
	}{
		{"empty-file", "", []string{}},
		{"single-file", "someFile", []string{"someFile"}},
		{"spaced-file", heredoc.Doc(`

			firstFile
			sandwhich

			lastFile

		`), []string{"firstFile", "sandwhich", "lastFile"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := vos.MkTempDir(os)

			cfgPath := path.Join(tmpDir, "cfg")
			testos.RequireWrite(t, os, cfgPath, tc.contents)

			fileNames, _, err := cleardir.ParseClearables(cfgPath)
			assert.Equal(t, tc.want, fileNames)
			assert.NoError(t, err)
		})
	}
}

func TestParseClearablesErrNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := path.Join(tmpDir, "missing")
	fileNames, _, err := cleardir.ParseClearables(cfgPath)
	assert.Nil(t, fileNames)
	assert.Error(t, err)
}

func TestParseClearablesDefCfg(t *testing.T) {
	os, reset := vos.Patch()
	defer reset()

	contents := heredoc.Doc(`
		aDefaultFile
		anotherDefaultFile
	`)
	want := []string{"aDefaultFile", "anotherDefaultFile"}

	defCfgPath := requireDefaultCfgPath(t, os)
	testos.RequireMkdirAll(t, os, path.Dir(defCfgPath))
	testos.RequireWrite(t, os, defCfgPath, contents)

	fileNames, _, err := cleardir.ParseClearables("")
	assert.Equal(t, want, fileNames)
	assert.NoError(t, err)
}

func TestParseClearablesDefCfgMissing(t *testing.T) {
	_, reset := vos.Patch()
	defer reset()

	want := []string{}

	fileNames, _, err := cleardir.ParseClearables("")
	assert.Equal(t, want, fileNames)
	assert.NoError(t, err)
}
func TestParseClearablesCfgPath(t *testing.T) {
	os, reset := vos.Patch()
	defer reset()

	t.Run("default path", func(t *testing.T) {
		want := requireDefaultCfgPath(t, os)
		_, got, _ := cleardir.ParseClearables("")
		assert.Equal(t, want, got)
	})

	t.Run("custom path", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgPath := path.Join(tmpDir, "my-config")
		_, got, _ := cleardir.ParseClearables(cfgPath)
		assert.Equal(t, cfgPath, got)
	})
}

func requireDefaultCfgPath(t *testing.T, os osa.I) string {
	cfgRoot, err := os.UserConfigDir()
	require.NoError(t, err)
	return path.Join(cfgRoot, "cleardir", "clearignore")
}
