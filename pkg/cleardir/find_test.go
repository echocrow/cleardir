package cleardir_test

import (
	"fmt"
	"path"
	"testing"

	"github.com/echocrow/cleardir/pkg/cleardir"
	"github.com/echocrow/osa/vos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindClearables(t *testing.T) {
	v, reset := vos.Patch()
	defer reset()

	sampleFSD := fsd{
		"d0": {},
		"d1": {
			"f0":  nil,
			"f1":  nil,
			"sd0": {"f0": nil},
			"sd1": {},
		},
		"d2": {"sd3": {}},
		"f1": nil,
	}

	tests := []struct {
		fsd    fsd
		mDepth int
		trvs   []string
		want   []string
	}{
		{fsd{}, -1, nil, nil},

		{fsd{"f0": nil}, -1, nil, []string{}},
		{fsd{"f0": nil}, -1, []string{"f0"}, []string{"f0"}},

		{fsd{"d0": {"d1": {"f0": nil}}}, -1, nil, []string{}},

		{sampleFSD, -1, nil,
			[]string{"d0", "d1/sd1", "d2/sd3", "d2"},
		},
		{sampleFSD, -1, []string{"f0"},
			[]string{"d0", "d1/f0", "d1/sd0/f0", "d1/sd0", "d1/sd1", "d2/sd3", "d2"},
		},
		{sampleFSD, -1, []string{"d0", "d1", "d2"},
			[]string{"d0", "d1/sd1", "d2/sd3", "d2"},
		},

		{sampleFSD, 0, []string{"f0"}, nil},
		{sampleFSD, 1, []string{"f0"}, []string{"d0", "d1/f0"}},
		{sampleFSD, 0, []string{"f1"}, []string{"f1"}},
		{sampleFSD, 1, []string{"f0"}, []string{"d0", "d1/f0"}},
		{sampleFSD, 1, []string{"f1"}, []string{"d0", "d1/f1", "f1"}},
		{sampleFSD, 2, []string{"f0"},
			[]string{"d0", "d1/f0", "d1/sd0/f0", "d1/sd0", "d1/sd1", "d2/sd3", "d2"},
		},
		{sampleFSD, 2, []string{"f1"},
			[]string{"d0", "d1/f1", "d1/sd1", "d2/sd3", "d2", "f1"},
		},

		{sampleFSD, 100, []string{"f0"},
			[]string{"d0", "d1/f0", "d1/sd0/f0", "d1/sd0", "d1/sd1", "d2/sd3", "d2"},
		},

		{sampleFSD, -1, []string{"f0", "f1"},
			[]string{"d0", "d1/f0", "d1/f1", "d1/sd0/f0", "d1/sd0", "d1/sd1", "d1", "d2/sd3", "d2", "f1"},
		},
	}
	for i, tc := range tests {
		tc := tc
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			dir := vos.MkTempDir(v)
			err := tc.fsd.Write(dir)
			require.NoError(t, err)

			matches := make(chan string)
			go func() {
				err := cleardir.FindClearables(
					matches,
					dir,
					tc.trvs,
					tc.mDepth,
				)
				assert.NoError(t, err)
				close(matches)
			}()
			gotMatches := []string{}
			for m := range matches {
				gotMatches = append(gotMatches, m)
			}

			wantMatches := joinBaseDir(dir, tc.want)
			assert.Equal(t, wantMatches, gotMatches)
		})
	}
}

func TestFindClearablesErrInvalidDir(t *testing.T) {
	os, reset := vos.Patch()
	defer reset()

	dir := vos.MkTempDir(os)
	invalidDir := path.Join(dir, "missing")

	matches := make(chan string)
	var err error
	go func() {
		err = cleardir.FindClearables(matches, invalidDir, nil, -1)
		close(matches)
	}()
	gotMatches := []string{}
	for m := range matches {
		gotMatches = append(gotMatches, m)
	}
	assert.Error(t, err)
	assert.Empty(t, gotMatches)
}
