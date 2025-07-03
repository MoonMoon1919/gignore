package gignore

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		ignore   IgnoreFile
		expected Ruler
	}{
		{
			name: "Pass-Extension",
			line: "*.txt",
			ignore: IgnoreFile{
				rules: []Ruler{},
			},
			expected: ExtensionRule{
				ext: "txt",
				act: INCLUDE,
			},
		},
		{
			name: "Pass-ExtensionExclude",
			line: "!*.txt",
			ignore: IgnoreFile{
				rules: []Ruler{},
			},
			expected: ExtensionRule{
				ext: "txt",
				act: EXCLUDE,
			},
		},
		{
			name: "Pass-DirectoryRecursive",
			line: "build/**",
			ignore: IgnoreFile{
				rules: []Ruler{},
			},
			expected: DirectoryRule{
				name: "build",
				mode: RECURSIVE,
				act:  INCLUDE,
			},
		},
		{
			name: "Pass-DirectoryChildren",
			line: "build/*",
			ignore: IgnoreFile{
				rules: []Ruler{},
			},
			expected: DirectoryRule{
				name: "build",
				mode: CHILDREN,
				act:  INCLUDE,
			},
		},
		{
			name: "Pass-Directory",
			line: "build/",
			ignore: IgnoreFile{
				rules: []Ruler{},
			},
			expected: DirectoryRule{
				name: "build",
				mode: DIRECTORY,
				act:  INCLUDE,
			},
		},
		{
			name: "Pass-DirectoryAnywhere",
			line: "**/build",
			ignore: IgnoreFile{
				rules: []Ruler{},
			},
			expected: DirectoryRule{
				name: "build",
				mode: ANYWHERE,
				act:  INCLUDE,
			},
		},
		{
			name: "Pass-Glob",
			line: "buildlogs*.txt",
			ignore: IgnoreFile{
				rules: []Ruler{},
			},
			expected: GlobRule{
				pattern: "buildlogs*.txt",
				act:     INCLUDE,
			},
		},
		{
			name: "Pass-File",
			line: "todo.md",
			ignore: IgnoreFile{
				rules: []Ruler{},
			},
			expected: FileRule{
				path: "todo.md",
				act:  INCLUDE,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			Parse(tc.line, &tc.ignore)

			var found bool

			for _, rule := range tc.ignore.Rules() {
				if rulesEqual(tc.expected, rule) {
					found = true
				}
			}

			if !found {
				t.Errorf("did not find expected rule %v", tc.expected)
			}
		})
	}
}

func TestParseMultiLine(t *testing.T) {
	content := `# This is a comment
	/.pnp
	*.log
	build/
	!build/important.txt
	node_modules/**
	temp*.backup

	src/main.go`

	expected := []Ruler{
		DirectoryRule{
			name: ".pnp",
			mode: ROOT_ONLY,
			act:  INCLUDE,
		},
		ExtensionRule{
			ext: "log",
			act: INCLUDE,
		},
		DirectoryRule{
			name: "build",
			mode: DIRECTORY,
			act:  INCLUDE,
		},
		FileRule{
			path: "build/important.txt",
			act:  EXCLUDE,
		},
		DirectoryRule{
			name: "node_modules",
			mode: RECURSIVE,
			act:  INCLUDE,
		},
		GlobRule{
			pattern: "temp*.backup",
			act:     INCLUDE,
		},
		FileRule{
			path: "src/main.go",
			act:  INCLUDE,
		},
	}

	ignore := IgnoreFile{rules: []Ruler{}}

	Parse(content, &ignore)

	for fIdx, foundRule := range ignore.Rules() {
		for eIdx, expectedRule := range expected {
			if fIdx != eIdx {
				continue
			}

			if foundRule != expectedRule {
				t.Errorf("Found rule %v, expected %v", foundRule, expectedRule)
			}
		}
	}
}
