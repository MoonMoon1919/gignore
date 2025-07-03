package gignore

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestRoundTripFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name: "Pass-Simple",
			content: `*.log
			build/
			!build/important.txt
			node_modules/**
			temp*.backup
			`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reader := strings.NewReader(tc.content)
			var ignoreFile IgnoreFile

			err := LoadFile(reader, &ignoreFile)
			if err != nil {
				t.Errorf("unexpected error reading file: %s", err.Error())
			}

			var buf bytes.Buffer
			err = WriteFile(&buf, &ignoreFile, RenderOptions{})
			if err != nil {
				t.Errorf("unexpected error writing file: %s", err.Error())
			}

			expectedLines := strings.Fields(strings.ReplaceAll(tc.content, "\n", " "))
			actualLines := strings.Fields(strings.ReplaceAll(buf.String(), "\n", " "))

			if !reflect.DeepEqual(expectedLines, actualLines) {
				t.Errorf("content mismatch:\nexpected: %v\nactual: %v", expectedLines, actualLines)
			}
		})
	}
}
