package cleardir_test

import (
	"fmt"
	"testing"

	"github.com/echocrow/cleardir/pkg/cleardir"
	"github.com/echocrow/fsnap/dirsnap"
	"github.com/echocrow/osa/vos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type rms []string

func TestRemove(t *testing.T) {
	os, reset := vos.Patch()
	defer reset()

	tests := []struct {
		fsd  fsd
		rms  rms
		want fsd
	}{
		{fsd{}, nil, fsd{}},
		{fsd{"f": nil}, nil, fsd{"f": nil}},
		{fsd{"d": fsd{}}, nil, fsd{"d": fsd{}}},

		{fsd{"d": fsd{}, "f": nil}, rms{"f"}, fsd{"d": fsd{}}},
		{fsd{"d": fsd{}, "f": nil}, rms{"d"}, fsd{"f": nil}},
		{fsd{"d": fsd{}, "f": nil}, rms{"d", "f"}, fsd{}},
		{fsd{"d": fsd{}, "f": nil}, rms{"f", "d"}, fsd{}},

		{fsd{"d": fsd{"s0": nil, "s1": nil}}, rms{"d/s0"}, fsd{"d": fsd{"s1": nil}}},
		{fsd{"d": fsd{"s0": nil, "s1": nil}}, rms{"d/s0", "d/s1"}, fsd{"d": fsd{}}},
		{fsd{"d": fsd{"s0": nil, "s1": nil}}, rms{"d/s0", "d/s1", "d"}, fsd{}},
	}
	for i, tc := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			tmpDir := vos.MkTempDir(os)
			err := tc.fsd.Write(tmpDir)
			require.NoError(t, err)

			rms := joinBaseDir(tmpDir, tc.rms)

			gotErr := cleardir.Remove(rms...)
			assert.NoError(t, gotErr)

			gotFsd, fsdErr := dirsnap.Read(tmpDir, -1)
			require.NoError(t, fsdErr)

			assert.Equal(t, tc.want, gotFsd)
		})
	}
}

func TestRemoveErr(t *testing.T) {
	os, reset := vos.Patch()
	defer reset()

	tests := []struct {
		fsd fsd
		rms rms
	}{
		{fsd{}, rms{"foo"}},
		{fsd{"d": fsd{"s": nil}}, rms{"d"}},
	}
	for i, tc := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			tmpDir := vos.MkTempDir(os)
			err := tc.fsd.Write(tmpDir)
			require.NoError(t, err)

			rms := joinBaseDir(tmpDir, tc.rms)

			gotErr := cleardir.Remove(rms...)
			assert.Error(t, gotErr)
		})
	}
}
