package gignore

import (
	"testing"
)

// MARK: File
func TestNewFileRule(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		act          Action
		errorMessage string
	}{
		{
			name:         "Pass",
			path:         "todo.md",
			act:          INCLUDE,
			errorMessage: "",
		},
		{
			name:         "Fail-InvalidAction",
			path:         "todo.md",
			act:          Action(99),
			errorMessage: invalidActionError.Error(),
		},
		{
			name:         "Fail-InvalidPath",
			path:         "",
			act:          INCLUDE,
			errorMessage: emptyPathError.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := NewFileRule(tc.path, tc.act)

			var errMessage string
			if err != nil {
				errMessage = err.Error()
			}

			if errMessage != tc.errorMessage {
				t.Errorf("expected error %s, got %s", tc.errorMessage, errMessage)
			}

			if tc.errorMessage != "" {
				return
			}

			if out.path != tc.path {
				t.Errorf("expected path %s, got %s", tc.path, out.path)
			}

			if out.act != tc.act {
				t.Errorf("expected action %d, got %d", tc.act, out.act)
			}
		})
	}
}

// MARK: Extension
func TestNewExtensionRule(t *testing.T) {
	tests := []struct {
		name         string
		ext          string
		expected     string
		act          Action
		errorMessage string
	}{
		{
			name:         "Pass-WithDot",
			ext:          "*.exe",
			expected:     "exe",
			act:          INCLUDE,
			errorMessage: "",
		},
		{
			name:         "Pass-NoDot",
			ext:          "exe",
			expected:     "exe",
			act:          INCLUDE,
			errorMessage: "",
		},
		{
			name:         "Fail-InvalidAction",
			ext:          "*.exe",
			act:          Action(99),
			errorMessage: invalidActionError.Error(),
		},
		{
			name:         "Fail-InvalidPath",
			ext:          "",
			act:          INCLUDE,
			errorMessage: emptyExtensionError.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := NewExtensionRule(tc.ext, tc.act)

			var errMessage string
			if err != nil {
				errMessage = err.Error()
			}

			if errMessage != tc.errorMessage {
				t.Errorf("expected error %s, got %s", tc.errorMessage, errMessage)
			}

			if tc.errorMessage != "" {
				return
			}

			if out.ext != tc.expected {
				t.Errorf("expected path %s, got %s", tc.expected, out.ext)
			}

			if out.act != tc.act {
				t.Errorf("expected action %d, got %d", tc.act, out.act)
			}
		})
	}
}

// MARK: Directories
func TestNewDirectoryRule(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		act          Action
		mode         DirectoryMode
		expected     string
		errorMessage string
	}{
		{
			name:         "Pass-Basic",
			path:         "node_modules/",
			act:          INCLUDE,
			mode:         DIRECTORY,
			expected:     "node_modules",
			errorMessage: "",
		},
		{
			name:         "Pass-Basic",
			path:         "/node_modules/",
			act:          INCLUDE,
			mode:         ROOT_ONLY,
			expected:     "node_modules",
			errorMessage: "",
		},
		{
			name:         "Pass-Children",
			path:         ".yarn/",
			act:          INCLUDE,
			mode:         CHILDREN,
			expected:     ".yarn",
			errorMessage: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := NewDirectoryRule(tc.path, tc.mode, tc.act)

			var errMessage string
			if err != nil {
				errMessage = err.Error()
			}

			if errMessage != tc.errorMessage {
				t.Errorf("expected error %s, got %s", tc.errorMessage, errMessage)
			}

			if tc.errorMessage != "" {
				return
			}

			if out.name != tc.expected {
				t.Errorf("expected path %s, got %s", tc.expected, out.name)
			}

			if out.act != tc.act {
				t.Errorf("expected action %d, got %d", tc.act, out.act)
			}
		})
	}
}

// MARK: Glob
func TestNewGlobRule(t *testing.T) {
	tests := []struct {
		name         string
		pattern      string
		act          Action
		errorMessage string
	}{
		{
			name:         "Pass",
			pattern:      "npm-debug.log*",
			act:          INCLUDE,
			errorMessage: "",
		},
		{
			name:         "Fail-InvalidAction",
			pattern:      "todo.md",
			act:          Action(99),
			errorMessage: invalidActionError.Error(),
		},
		{
			name:         "Fail-InvalidPattern",
			pattern:      "",
			act:          INCLUDE,
			errorMessage: emptyGlobPatternError.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := NewGlobRule(tc.pattern, tc.act)

			var errMessage string
			if err != nil {
				errMessage = err.Error()
			}

			if errMessage != tc.errorMessage {
				t.Errorf("expected error %s, got %s", tc.errorMessage, errMessage)
			}

			if tc.errorMessage != "" {
				return
			}

			if out.pattern != tc.pattern {
				t.Errorf("expected pattern %s, got %s", tc.pattern, out.pattern)
			}

			if out.act != tc.act {
				t.Errorf("expected action %d, got %d", tc.act, out.act)
			}
		})
	}
}

// MARK: Render
func TestRendererFuncs(t *testing.T) {
	tests := []struct {
		name           string
		rule           Ruler
		expectedOutput string
	}{
		{
			name: "Pass-File",
			rule: FileRule{
				path: "todo.md",
				act:  INCLUDE,
			},
			expectedOutput: "todo.md",
		},
		{
			name: "Pass-File",
			rule: FileRule{
				path: "foo",
				act:  EXCLUDE,
			},
			expectedOutput: "!foo",
		},
		{
			name: "Pass-Extension",
			rule: ExtensionRule{
				ext: "exe",
				act: INCLUDE,
			},
			expectedOutput: "*.exe",
		},
		{
			name: "Pass-Extension",
			rule: ExtensionRule{
				ext: "exe",
				act: EXCLUDE,
			},
			expectedOutput: "!*.exe",
		},
		{
			name: "Pass-DirectoryBasic",
			rule: DirectoryRule{
				name: "node_modules",
				mode: DIRECTORY,
				act:  INCLUDE,
			},
			expectedOutput: "node_modules/",
		},
		{
			name: "Pass-DirectoryChildren",
			rule: DirectoryRule{
				name: ".yarn",
				mode: CHILDREN,
				act:  INCLUDE,
			},
			expectedOutput: ".yarn/*",
		},
		{
			name: "Pass-DirectoryRecursive",
			rule: DirectoryRule{
				name: "database",
				mode: RECURSIVE,
				act:  INCLUDE,
			},
			expectedOutput: "database/**",
		},
		{
			name: "Pass-DirectoryRecursive",
			rule: DirectoryRule{
				name: ".vitepress",
				mode: ANYWHERE,
				act:  INCLUDE,
			},
			expectedOutput: "**/.vitepress",
		},
		{
			name: "Pass-GlobInclude",
			rule: GlobRule{
				pattern: "todo.*",
				act:     INCLUDE,
			},
			expectedOutput: "todo.*",
		},
		{
			name: "Pass-GlobExclude",
			rule: GlobRule{
				pattern: "msg/*Result.msg",
				act:     EXCLUDE,
			},
			expectedOutput: "!msg/*Result.msg",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output := tc.rule.Render()

			if output != tc.expectedOutput {
				t.Errorf("expected %s, got %s", tc.expectedOutput, output)
			}
		})
	}
}

// MARK: IgnoreFile
func TestAddFile(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		action       Action
		ignore       IgnoreFile
		errorMessage string
	}{
		{
			name:   "Pass",
			path:   "todo.md",
			action: INCLUDE,
			ignore: IgnoreFile{
				rules: []Ruler{},
			},
			errorMessage: "",
		},
		{
			name:   "Fail-Conflict",
			path:   "todo.txt",
			action: INCLUDE,
			ignore: IgnoreFile{
				rules: []Ruler{
					ExtensionRule{
						ext: "txt",
						act: INCLUDE,
					},
				},
			},
			errorMessage: unreachableRuleError.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.ignore.AddFile(tc.path, tc.action)

			var errMessage string
			if err != nil {
				errMessage = err.Error()
			}

			if errMessage != tc.errorMessage {
				t.Errorf("expected error %s, got %s", tc.errorMessage, errMessage)
			}

			if tc.errorMessage != "" {
				return
			}

			var found bool = false
			for _, rule := range tc.ignore.rules {
				if rulesEqual(rule, FileRule{
					path: tc.path,
					act:  tc.action,
				}) {
					found = true
				}
			}

			if !found {
				t.Errorf("expected to find matching rule")
			}
		})
	}
}

func TestAddExtension(t *testing.T) {
	tests := []struct {
		name         string
		ext          string
		action       Action
		ignore       IgnoreFile
		errorMessage string
	}{
		{
			name:   "Pass",
			ext:    "txt",
			action: INCLUDE,
			ignore: IgnoreFile{
				rules: []Ruler{},
			},
			errorMessage: "",
		},
		{
			name:   "Fail-Conflict",
			ext:    "txt",
			action: INCLUDE,
			ignore: IgnoreFile{
				rules: []Ruler{
					GlobRule{
						pattern: "*.txt",
						act:     INCLUDE,
					},
				},
			},
			errorMessage: redundantRuleError.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.ignore.AddExtension(tc.ext, tc.action)

			var errMessage string
			if err != nil {
				errMessage = err.Error()
			}

			if errMessage != tc.errorMessage {
				t.Errorf("expected error %s, got %s", tc.errorMessage, errMessage)
			}

			if tc.errorMessage != "" {
				return
			}

			var found bool = false
			for _, rule := range tc.ignore.rules {
				if rulesEqual(rule, ExtensionRule{
					ext: tc.ext,
					act: tc.action,
				}) {
					found = true
				}
			}

			if !found {
				t.Errorf("expected to find matching rule")
			}
		})
	}
}

func TestAddDirectory(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		mode         DirectoryMode
		action       Action
		ignore       IgnoreFile
		errorMessage string
	}{
		{
			name:   "Pass",
			path:   "build",
			mode:   CHILDREN,
			action: INCLUDE,
			ignore: IgnoreFile{
				rules: []Ruler{},
			},
			errorMessage: "",
		},
		{
			name:   "Fail-Conflict",
			path:   "build",
			mode:   CHILDREN,
			action: INCLUDE,
			ignore: IgnoreFile{
				rules: []Ruler{
					DirectoryRule{
						name: "build",
						mode: RECURSIVE,
						act:  INCLUDE,
					},
				},
			},
			errorMessage: unreachableRuleError.Error(),
		},
		{
			name:   "Pass-FixIneffective",
			path:   "build",
			mode:   RECURSIVE,
			action: INCLUDE,
			ignore: IgnoreFile{
				rules: []Ruler{
					FileRule{
						path: "build/important.txt",
						act:  EXCLUDE,
					},
				},
			},
			errorMessage: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.ignore.AddDirectory(tc.path, tc.mode, tc.action)

			var errMessage string
			if err != nil {
				errMessage = err.Error()
			}

			if errMessage != tc.errorMessage {
				t.Errorf("expected error %s, got %s", tc.errorMessage, errMessage)
			}

			if tc.errorMessage != "" {
				return
			}

			var found bool = false
			for _, rule := range tc.ignore.rules {
				expected := DirectoryRule{
					name: tc.path,
					mode: tc.mode,
					act:  tc.action,
				}
				if rulesEqual(rule, expected) {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("expected to find matching rule")
			}
		})
	}
}

func TestAddGlob(t *testing.T) {
	tests := []struct {
		name         string
		pattern      string
		action       Action
		ignore       IgnoreFile
		errorMessage string
	}{
		{
			name:    "Pass",
			pattern: "buildlog*.txt",
			action:  INCLUDE,
			ignore: IgnoreFile{
				rules: []Ruler{},
			},
			errorMessage: "",
		},
		{
			name:    "Fail-Conflict",
			pattern: "buildlog*.txt",
			action:  INCLUDE,
			ignore: IgnoreFile{
				rules: []Ruler{
					ExtensionRule{
						ext: "txt",
						act: INCLUDE,
					},
				},
			},
			errorMessage: unreachableRuleError.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.ignore.AddGlob(tc.pattern, tc.action)

			var errMessage string
			if err != nil {
				errMessage = err.Error()
			}

			if errMessage != tc.errorMessage {
				t.Errorf("expected error %s, got %s", tc.errorMessage, errMessage)
			}

			if tc.errorMessage != "" {
				return
			}

			var found bool = false
			for _, rule := range tc.ignore.rules {
				if rulesEqual(rule, GlobRule{
					pattern: tc.pattern,
					act:     tc.action,
				}) {
					found = true
				}
			}

			if !found {
				t.Errorf("expected to find matching rule")
			}
		})
	}
}

func TestDeleteFile(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		action       Action
		ignore       IgnoreFile
		errorMessage string
	}{
		{
			name:   "Pass",
			path:   "todo.md",
			action: INCLUDE,
			ignore: IgnoreFile{
				rules: []Ruler{
					FileRule{
						path: "todo.md",
						act:  INCLUDE,
					},
				},
			},
			errorMessage: "",
		},
		{
			name:   "Fail-NotFound",
			path:   "todo.md",
			action: INCLUDE,
			ignore: IgnoreFile{
				rules: []Ruler{},
			},
			errorMessage: ruleNotFoundError.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.ignore.DeleteFile(tc.path, tc.action)

			var errMessage string
			if err != nil {
				errMessage = err.Error()
			}

			if errMessage != tc.errorMessage {
				t.Errorf("expected error %s, got %s", tc.errorMessage, errMessage)
			}

			if tc.errorMessage != "" {
				return
			}

			var found bool
			for _, rule := range tc.ignore.rules {
				if rulesEqual(rule, FileRule{
					path: tc.path,
					act:  tc.action,
				}) {
					found = true
				}
			}

			if found {
				t.Errorf("expected to not find matching rule")
			}
		})
	}
}

func TestDeleteExtension(t *testing.T) {
	tests := []struct {
		name         string
		ext          string
		action       Action
		ignore       IgnoreFile
		errorMessage string
	}{
		{
			name:   "Pass",
			ext:    "ext",
			action: INCLUDE,
			ignore: IgnoreFile{
				rules: []Ruler{
					ExtensionRule{
						ext: "ext",
						act: INCLUDE,
					},
				},
			},
			errorMessage: "",
		},
		{
			name:   "Fail-NotFound",
			ext:    "ext",
			action: INCLUDE,
			ignore: IgnoreFile{
				rules: []Ruler{},
			},
			errorMessage: ruleNotFoundError.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.ignore.DeleteExtension(tc.ext, tc.action)

			var errMessage string
			if err != nil {
				errMessage = err.Error()
			}

			if errMessage != tc.errorMessage {
				t.Errorf("expected error %s, got %s", tc.errorMessage, errMessage)
			}

			if tc.errorMessage != "" {
				return
			}

			var found bool
			for _, rule := range tc.ignore.rules {
				if rulesEqual(rule, ExtensionRule{
					ext: tc.ext,
					act: tc.action,
				}) {
					found = true
				}
			}

			if found {
				t.Errorf("expected to not find matching rule")
			}
		})
	}
}

func TestDeleteDirectory(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		mode         DirectoryMode
		action       Action
		ignore       IgnoreFile
		errorMessage string
	}{
		{
			name:   "Pass",
			path:   "build",
			mode:   DIRECTORY,
			action: INCLUDE,
			ignore: IgnoreFile{
				rules: []Ruler{
					DirectoryRule{
						name: "build",
						mode: DIRECTORY,
						act:  INCLUDE,
					},
				},
			},
			errorMessage: "",
		},
		{
			name:   "Fail-NotFound",
			path:   "build",
			mode:   DIRECTORY,
			action: INCLUDE,
			ignore: IgnoreFile{
				rules: []Ruler{},
			},
			errorMessage: ruleNotFoundError.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.ignore.DeleteDirectory(tc.path, tc.mode, tc.action)

			var errMessage string
			if err != nil {
				errMessage = err.Error()
			}

			if errMessage != tc.errorMessage {
				t.Errorf("expected error %s, got %s", tc.errorMessage, errMessage)
			}

			if tc.errorMessage != "" {
				return
			}

			var found bool
			for _, rule := range tc.ignore.rules {
				if rulesEqual(rule, DirectoryRule{
					name: tc.path,
					mode: tc.mode,
					act:  tc.action,
				}) {
					found = true
				}
			}

			if found {
				t.Errorf("expected to not find matching rule")
			}
		})
	}
}

func TestDeleteGlob(t *testing.T) {
	tests := []struct {
		name         string
		pattern      string
		action       Action
		ignore       IgnoreFile
		errorMessage string
	}{
		{
			name:    "Pass",
			pattern: "buildlog*.txt",
			action:  INCLUDE,
			ignore: IgnoreFile{
				rules: []Ruler{
					GlobRule{
						pattern: "buildlog*.txt",
						act:     INCLUDE,
					},
				},
			},
			errorMessage: "",
		},
		{
			name:    "Fail-NotFound",
			pattern: "buildlog*.txt",
			action:  INCLUDE,
			ignore: IgnoreFile{
				rules: []Ruler{},
			},
			errorMessage: ruleNotFoundError.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.ignore.DeleteGlob(tc.pattern, tc.action)

			var errMessage string
			if err != nil {
				errMessage = err.Error()
			}

			if errMessage != tc.errorMessage {
				t.Errorf("expected error %s, got %s", tc.errorMessage, errMessage)
			}

			if tc.errorMessage != "" {
				return
			}

			var found bool
			for _, rule := range tc.ignore.rules {
				if rulesEqual(rule, GlobRule{
					pattern: tc.pattern,
					act:     tc.action,
				}) {
					found = true
				}
			}

			if found {
				t.Errorf("expected to not find matching rule")
			}
		})
	}
}

// MARK: Read methods
func TestFindConflicts(t *testing.T) {
	tests := []struct {
		name          string
		ignore        IgnoreFile
		conflictCount int
	}{
		{
			name: "Pass-NoConflicts",
			ignore: IgnoreFile{
				rules: []Ruler{
					ExtensionRule{
						ext: "txt",
						act: INCLUDE,
					},
					FileRule{
						path: "todo.md",
						act:  INCLUDE,
					},
				},
			},
			conflictCount: 0,
		},
		{
			name: "Fail-SomeConflicts",
			ignore: IgnoreFile{
				rules: []Ruler{
					ExtensionRule{
						ext: "txt",
						act: INCLUDE,
					},
					GlobRule{
						pattern: "buildlogs*.txt",
						act:     INCLUDE,
					},
				},
			},
			conflictCount: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			conflicts := tc.ignore.FindConflicts()

			if len(conflicts) != tc.conflictCount {
				t.Errorf("expected %d conflicts, found %d", tc.conflictCount, len(conflicts))
			}
		})
	}
}

// MARK: Direction
func TestMoveRule(t *testing.T) {
	tests := []struct {
		name         string
		ignore       IgnoreFile
		ruleToMove   Ruler
		targetRule   Ruler
		direction    MoveDirection
		errorMessage string
	}{
		{
			name: "Pass-After-Simple",
			ignore: IgnoreFile{
				rules: []Ruler{
					FileRule{
						path: "build/important.txt",
						act:  EXCLUDE,
					},
					DirectoryRule{
						name: "build",
						mode: RECURSIVE,
						act:  INCLUDE,
					},
				},
			},
			ruleToMove: FileRule{
				path: "build/important.txt",
				act:  EXCLUDE,
			},
			targetRule: DirectoryRule{
				name: "build",
				mode: RECURSIVE,
				act:  INCLUDE,
			},
			direction:    AFTER,
			errorMessage: "",
		},
		{
			name: "Pass-After-Complicated",
			ignore: IgnoreFile{
				rules: []Ruler{
					ExtensionRule{
						ext: "txt",
						act: INCLUDE,
					},
					FileRule{
						path: "build/important.txt",
						act:  EXCLUDE,
					},
					DirectoryRule{
						name: "docs/generated",
						mode: CHILDREN,
						act:  INCLUDE,
					},
					DirectoryRule{
						name: "build",
						mode: RECURSIVE,
						act:  INCLUDE,
					},
					FileRule{
						path: "todo.md",
						act:  INCLUDE,
					},
				},
			},
			ruleToMove: FileRule{
				path: "build/important.txt",
				act:  EXCLUDE,
			},
			targetRule: DirectoryRule{
				name: "build",
				mode: RECURSIVE,
				act:  INCLUDE,
			},
			direction:    AFTER,
			errorMessage: "",
		},
		{
			name: "Pass-Before-Simple",
			ignore: IgnoreFile{
				rules: []Ruler{
					FileRule{
						path: "build/important.txt",
						act:  EXCLUDE,
					},
					DirectoryRule{
						name: "build",
						mode: RECURSIVE,
						act:  INCLUDE,
					},
				},
			},
			ruleToMove: DirectoryRule{
				name: "build",
				mode: RECURSIVE,
				act:  INCLUDE,
			},
			targetRule: FileRule{
				path: "build/important.txt",
				act:  EXCLUDE,
			},
			direction:    BEFORE,
			errorMessage: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.ignore.MoveRule(
				tc.ruleToMove,
				tc.targetRule,
				tc.direction,
				REQUESTED,
			)

			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}

			if errMsg != tc.errorMessage {
				t.Errorf("expected error message %s, received %s", tc.errorMessage, errMsg)
				return
			}

			// Comparisons
			moveIdx := tc.ignore.findRuleIndex(tc.ruleToMove)
			afterIdx := tc.ignore.findRuleIndex(tc.targetRule)

			var expectedIdx int
			switch tc.direction {
			case BEFORE:
				// Check if the index of ruleToMove is directly before index of targetRule
				expectedIdx = afterIdx - 1
			case AFTER:
				// Check if the index of ruleToMove is directly after index of targetRule
				expectedIdx = afterIdx + 1
			}

			if moveIdx != expectedIdx {
				t.Errorf("expected moved rule index to be %d, got %d", expectedIdx, moveIdx)
			}
		})
	}
}

func TestFixConflict(t *testing.T) {
	tests := []struct {
		name         string
		ignore       IgnoreFile
		errorMessage string
		result       []Ruler
	}{
		{
			name: "Pass-Complicated",
			ignore: IgnoreFile{
				rules: []Ruler{
					FileRule{
						path: "gignore-cli",
						act:  INCLUDE,
					},
					DirectoryRule{
						name: "build",
						mode: RECURSIVE,
						act:  INCLUDE,
					},
					FileRule{
						path: "build/important.txt",
						act:  EXCLUDE,
					},
					FileRule{ // Should be removed in favor of *.log
						path: "debug.log",
						act:  INCLUDE,
					},
					ExtensionRule{
						ext: "log",
						act: INCLUDE,
					},
					FileRule{ // Correct!
						path: "important.log",
						act:  EXCLUDE,
					},
					ExtensionRule{ // Duplicate - should be removed
						ext: "log",
						act: INCLUDE,
					},
				},
			},
			result: []Ruler{
				FileRule{
					path: "gignore-cli",
					act:  INCLUDE,
				},
				DirectoryRule{
					name: "build",
					mode: RECURSIVE,
					act:  INCLUDE,
				},
				FileRule{
					path: "build/important.txt",
					act:  EXCLUDE,
				},
				ExtensionRule{
					ext: "log",
					act: INCLUDE,
				},
				FileRule{
					path: "important.log",
					act:  EXCLUDE,
				},
			},
			errorMessage: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.ignore.FixConflicts(20)

			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}

			if errMsg != tc.errorMessage {
				t.Errorf("Expected err %s, got %s", tc.errorMessage, errMsg)
				return
			}

			if len(tc.ignore.rules) != len(tc.result) {
				t.Errorf("expected %d rules, found %d", len(tc.result), len(tc.ignore.rules))
				return
			}

			// Rule ordering matters, so we can directly compare
			for idx, rule := range tc.ignore.rules {
				expectedRule := tc.result[idx]

				if !rulesEqual(expectedRule, rule) {
					t.Errorf("Expected rule %s, found %s", expectedRule.Render(), rule.Render())
				}
			}
		})
	}
}
