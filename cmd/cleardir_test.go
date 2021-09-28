package cmd_test

import (
	"fmt"
	"path"
	"regexp"
	"strings"
	"testing"

	"github.com/echocrow/cleardir/cmd"
	"github.com/echocrow/fsnap/dirsnap"
	"github.com/echocrow/osa/testos"
	"github.com/echocrow/osa/vos"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fsd = dirsnap.Dirs

var version = "0.0.0-test"

var emptyRe = regexp.MustCompile(`^$`)

func TestCmdBasicOut(t *testing.T) {
	v, reset := vos.Patch()
	defer reset()

	_, stdout, stderr := vos.GetStdio(v)

	cfgPath, cfgPathErr := v.UserConfigDir()
	require.NoError(t, cfgPathErr, "must be able to get config path")

	tests := []struct {
		name    string
		args    []string
		wantOut interface{}
	}{
		{
			"Help",
			[]string{"--help"},
			regexp.MustCompile(`(?s)^Cleardir.+Usage:.+Flags:`),
		},
		{
			"Version",
			[]string{"--version"},
			version,
		},
		{
			"Get config path",
			[]string{"--config", "?"},
			cfgPath,
		},
		{
			"Clear",
			[]string{vos.MkTempDir(v)},
			"All clear",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := execWithArgs(tc.args...)
			assert.NoError(t, err)
			assert.Regexp(t, tc.wantOut, stdout)
			assert.Empty(t, stderr)
		})
	}
}
func TestCmdErrInvalidDir(t *testing.T) {
	v, reset := vos.Patch()
	defer reset()

	dir := vos.MkTempDir(v)
	invalidDir := path.Join(dir, "missing")

	_, stdout, stderr := vos.GetStdio(v)

	err := execWithArgsInDir(invalidDir)
	assert.Error(t, err)
	assert.True(t, v.IsNotExist(err), "expected not-exists error")
	assert.Regexp(t, "Usage", stdout)
	assert.Regexp(t, "Error:", stderr)
}

func TestCmdExcute(t *testing.T) {
	v, reset := vos.Patch()
	defer reset()

	srcFsd := fsd{
		"f": nil,
		"d": fsd{"sd": fsd{}, "sf": nil},
	}
	noEmptyDirFsd := fsd{
		"f": nil,
		"d": fsd{"sf": nil},
	}

	tests := []struct {
		name    string
		args    []string
		sendIn  string
		wantFsd fsd
		wantOut interface{}
		wantErr interface{}
	}{
		{
			"Run",
			nil, "y\n",
			noEmptyDirFsd,
			"", nil,
		},
		{
			"Accept Yes Prompt",
			nil, "yes\n",
			noEmptyDirFsd,
			"", nil,
		},
		{
			"Resolve",
			[]string{"-y"}, "",
			noEmptyDirFsd,
			"", nil,
		},
		{
			"Silent",
			[]string{"-s"}, "",
			noEmptyDirFsd,
			nil, nil,
		},
		{
			"Dry",
			[]string{"--dry"}, "",
			srcFsd,
			"", nil,
		},
		{
			"Custom File",
			[]string{"-f", "sf"}, "y\n",
			fsd{"f": nil},
			"", nil,
		},
		{
			"Custom Files",
			[]string{"-f", "f", "-f", "sf"}, "y\n",
			fsd{},
			"", nil,
		},
		{
			"Custom Depth",
			[]string{"-d", "0"}, "y\n",
			srcFsd,
			"", nil,
		},
		{
			"Custom Depth with Files",
			[]string{"-d", "0", "-f", "f", "-f", "sf"}, "y\n",
			fsd{"d": fsd{"sd": fsd{}, "sf": nil}},
			"", nil,
		},
		{
			"Abort Prompt",
			nil, "n\n",
			srcFsd,
			"", "Aborted",
		},
		{
			"Abort Invalid Prompt",
			nil, "foobar\n",
			srcFsd,
			"", "Aborted",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := vos.MkTempDir(v)
			err := srcFsd.Write(dir)
			require.NoError(t, err)

			stdin, stdout, stderr := vos.GetStdio(v)
			stdin.Write([]byte(tc.sendIn))

			err = execWithArgsInDir(dir, tc.args...)

			if tc.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}

			gotFsd, fsdErr := dirsnap.Read(dir, -1)
			require.NoError(t, fsdErr)
			assert.Equal(t, tc.wantFsd, gotFsd)

			if tc.wantOut == nil {
				assert.Regexp(t, emptyRe, stdout)
			} else if tc.wantOut != "" {
				assert.Regexp(t, tc.wantOut, stdout)
			}

			if tc.wantErr == nil {
				assert.Regexp(t, emptyRe, stderr)
			} else if tc.wantErr != "" {
				assert.Regexp(t, tc.wantErr, stderr)
			}
		})

		vos.ClearStdio(v)
	}
}

func TestCmdListFiles(t *testing.T) {
	v, reset := vos.Patch()
	defer reset()

	srcFsd := fsd{
		"ad": fsd{},
		"aa": nil,
		"f":  nil,
		"d":  fsd{"sd": fsd{}, "sf": nil},
	}

	delList := []string{"aa", "ad", "d/sd", "d/sf", "d"}

	baseArgs := []string{"-f", "sf", "-f", "aa"}

	tests := []struct {
		name string
		args []string
	}{
		{"Regular", nil},
		{"Dry", []string{"--dry"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := vos.MkTempDir(v)
			err := srcFsd.Write(dir)
			require.NoError(t, err)

			stdin, stdout, stderr := vos.GetStdio(v)
			stdin.Write([]byte("y\n"))

			err = execWithArgsInDir(dir, append(baseArgs, tc.args...)...)
			require.NoError(t, err)
			require.Empty(t, stderr)

			wantLstRe := ""
			for _, p := range joinBaseDir(dir, delList) {
				wantLstRe += "- " + p + "\n"
			}
			assert.Regexp(t, wantLstRe, stdout, "expected output to list files")
		})

		vos.ClearStdio(v)
	}
}

func TestCmdConfig(t *testing.T) {
	v, reset := vos.Patch()
	defer reset()

	srcFsd := fsd{
		"a": nil,
		"b": nil,
		"c": nil,
		"d": nil,
	}

	baseArgs := []string{"-y"}

	tests := []struct {
		name     string
		argFiles []string
		cfgFiles []string
		wantLeft []string
	}{
		{"None", nil, nil, []string{"a", "b", "c", "d"}},
		{"Args", []string{"a", "c"}, nil, []string{"b", "d"}},
		{"Config", nil, []string{"c", "d"}, []string{"a", "b"}},
		{"Args & Config", []string{"a", "c"}, []string{"c", "d"}, []string{"b"}},
	}
	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := vos.MkTempDir(v)
			err := srcFsd.Write(dir)
			require.NoError(t, err)

			args := baseArgs

			if tc.cfgFiles != nil {
				cfgDir, err := v.UserConfigDir()
				require.NoError(t, err)
				cfgPath := path.Join(cfgDir, fmt.Sprintf("cfg%d", i))
				testos.RequireWrite(t, v, cfgPath, strings.Join(tc.cfgFiles, "\n"))
				args = append(args, "-c", cfgPath)
			}

			for _, f := range tc.argFiles {
				args = append(args, "-f", f)
			}

			err = execWithArgsInDir(dir, args...)
			require.NoError(t, err)

			wantFsd := fsd{}
			for _, f := range tc.wantLeft {
				wantFsd[f] = nil
			}

			gotFsd, fsdErr := dirsnap.Read(dir, -1)
			require.NoError(t, fsdErr)
			assert.Equal(t, wantFsd, gotFsd)
		})

		vos.ClearStdio(v)
	}
}

func TestExecute(t *testing.T) {
	v, reset := vos.Patch()
	defer reset()

	srcFsd := fsd{"f": nil, "d": fsd{}}
	wantFsd := fsd{"f": nil}

	dir := vos.MkTempDir(v)

	err := srcFsd.Write(dir)
	require.NoError(t, err)

	setArgs, resetArgs := vos.PatchArgs()
	defer resetArgs()
	setArgs([]string{"cleardir", "-y", dir})

	cmd.Execute(version)

	gotFsd, fsdErr := dirsnap.Read(dir, -1)
	require.NoError(t, fsdErr)
	assert.Equal(t, gotFsd, wantFsd)
}

func TestExecuteErr(t *testing.T) {
	v, reset := vos.Patch()
	defer reset()

	dir := vos.MkTempDir(v)
	invalidDir := path.Join(dir, "missing")

	setArgs, resetArgs := vos.PatchArgs()
	defer resetArgs()
	setArgs([]string{"cleardir", "-y", invalidDir})

	gotExits := make(chan int, 1)
	defer vos.CatchExit(func(got int) {
		gotExits <- got
	})

	cmd.Execute(version)

	gotExit := <-gotExits
	wantExit := 1
	assert.Equal(t, wantExit, gotExit)
}

func newCmd() *cobra.Command {
	return cmd.NewCmd(version)
}

func execWithArgs(args ...string) error {
	c := newCmd()
	c.SetArgs(args)
	return c.Execute()
}

func execWithArgsInDir(dir string, args ...string) error {
	c := newCmd()
	c.SetArgs(append(args, dir))
	return c.Execute()
}

func joinBaseDir(root string, names []string) []string {
	out := make([]string, len(names))
	for i, p := range names {
		out[i] = path.Join(root, p)
	}
	return out
}
