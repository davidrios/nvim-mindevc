package utils

import (
	"strings"
	"testing"
)

func TestFileContainsLine(t *testing.T) {
	const CONTENT = `the quick
fox jumps
hello world
test test`

	testTable := []struct {
		name       string
		content    string
		lineToFind string
		want       bool
		wantErr    bool
	}{
		{
			name:       "test1",
			content:    CONTENT,
			lineToFind: "hello world",
			want:       true,
		},
		{
			name:       "test1",
			content:    CONTENT,
			lineToFind: "hell world",
			want:       false,
		},
	}

	for _, tv := range testTable {
		t.Run(tv.name, func(t *testing.T) {
			tmpFile := strings.NewReader(tv.content)

			got, gotErr := FileContainsLine(tmpFile, tv.lineToFind)
			if gotErr != nil {
				if !tv.wantErr {
					t.Errorf("FileContainsLine() failed: %v", gotErr)
				}
				return
			}

			if tv.wantErr {
				t.Fatal("FileContainsLine() succeeded unexpectedly")
			}

			if got != tv.want {
				t.Errorf("FileContainsLine() = %v, want %v", got, tv.want)
			}
		})
	}
}
