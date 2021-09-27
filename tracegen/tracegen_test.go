package tracegen

import (
	"bytes"
	"embed"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed testdata
var testdata embed.FS

func TestRun(t *testing.T) {
	type args struct {
		source string
	}
	type want struct {
		file string
	}

	test := func(args args, want want) func(t *testing.T) {
		return func(t *testing.T) {
			t.Helper()

			data, err := testdata.ReadFile(want.file)
			require.NoError(t, err)

			got := &bytes.Buffer{}

			err = Run(args.source, got)
			require.NoError(t, err)
			require.Equal(t, string(data), got.String())
		}
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "empty",
			args: args{source: "testdata/empty.go"},
			want: want{file: "testdata/empty_tracegen.go"},
		},
		{
			name: "full",
			args: args{source: "testdata/full.go"},
			want: want{file: "testdata/full_tracegen.go"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, test(tt.args, tt.want))
	}
}
