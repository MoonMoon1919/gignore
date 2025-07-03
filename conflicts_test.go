package gignore

import "testing"

func TestCheckConflict(t *testing.T) {
	type output struct {
		conflictType ConflictType
		has          bool
	}

	tests := []struct {
		name        string
		left        Ruler
		right       Ruler
		intervening []Ruler
		output      output
	}{
		{
			name: "Pass",
			left: FileRule{
				path: "todo.md",
				act:  INCLUDE,
			},
			right: DirectoryRule{
				name: "build",
				act:  INCLUDE,
			},
			intervening: []Ruler{},
			output: output{
				conflictType: ConflictType(""),
				has:          false,
			},
		},
		{
			name: "Pass-InterveniningException",
			left: DirectoryRule{
				name: "build",
				mode: DIRECTORY,
				act:  INCLUDE,
			},
			right: DirectoryRule{
				name: "build",
				mode: RECURSIVE,
				act:  INCLUDE,
			},
			intervening: []Ruler{
				FileRule{
					path: "build/important.txt",
					act:  EXCLUDE,
				},
			},
			output: output{
				conflictType: ConflictType(""),
				has:          false,
			},
		},
		{
			name: "Fail-InvalidOrdering",
			left: FileRule{
				path: "build/important.txt",
				act:  EXCLUDE,
			},
			right: DirectoryRule{
				name: "build",
				mode: RECURSIVE,
				act:  INCLUDE,
			},
			intervening: []Ruler{},
			output: output{
				conflictType: INEFFECTIVE_RULE,
				has:          true,
			},
		},
		{
			name: "Fail-Semantic",
			left: FileRule{
				path: "todo.md",
				act:  INCLUDE,
			},
			right: FileRule{
				path: "todo.md",
				act:  EXCLUDE,
			},
			intervening: []Ruler{},
			output: output{
				conflictType: SEMANTIC_CONFLICT,
				has:          true,
			},
		},
		{
			name: "Fail-Redunant",
			left: FileRule{
				path: "todo.md",
				act:  INCLUDE,
			},
			right: FileRule{
				path: "todo.md",
				act:  INCLUDE,
			},
			intervening: []Ruler{},
			output: output{
				conflictType: REDUNDANT_RULE,
				has:          true,
			},
		},
		{
			name: "Fail-DirectoryModeSubsumes",
			left: DirectoryRule{
				name: "build",
				act:  INCLUDE,
				mode: CHILDREN,
			},
			right: DirectoryRule{
				name: "build",
				act:  INCLUDE,
				mode: DIRECTORY,
			},
			intervening: []Ruler{},
			output: output{
				conflictType: UNREACHABLE_RULE,
				has:          true,
			},
		},
		{
			name: "Fail-DirectoryPathSubsumes",
			left: DirectoryRule{
				name: "build",
				act:  INCLUDE,
				mode: DIRECTORY,
			},
			right: DirectoryRule{
				name: "build/logs",
				act:  INCLUDE,
				mode: DIRECTORY,
			},
			intervening: []Ruler{},
			output: output{
				conflictType: UNREACHABLE_RULE,
				has:          true,
			},
		},
		{
			name: "Fail-GlobSubsumes",
			left: GlobRule{
				pattern: "*.txt",
				act:     INCLUDE,
			},
			right: FileRule{
				path: "build/todo.txt",
				act:  INCLUDE,
			},
			intervening: []Ruler{},
			output: output{
				conflictType: UNREACHABLE_RULE,
				has:          true,
			},
		},
		{
			name: "Fail-ExtensionSubsumesFile",
			left: ExtensionRule{
				ext: "txt",
				act: INCLUDE,
			},
			right: FileRule{
				path: "todo.txt",
				act:  INCLUDE,
			},
			intervening: []Ruler{},
			output: output{
				conflictType: UNREACHABLE_RULE,
				has:          true,
			},
		},
		{
			name: "Fail-ExtensionSubsumesGlob",
			left: ExtensionRule{
				ext: "txt",
				act: INCLUDE,
			},
			right: GlobRule{
				pattern: "logs*.txt",
				act:     INCLUDE,
			},
			intervening: []Ruler{},
			output: output{
				conflictType: UNREACHABLE_RULE,
				has:          true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			conflict, ok := checkConflict(
				tc.left,
				tc.right,
				tc.intervening,
			)

			if ok != tc.output.has {
				t.Errorf("expected conflict to be %t, got %t", tc.output.has, ok)
			}

			if conflict.ConflictType != tc.output.conflictType {
				t.Errorf("expected conflict type to be %s, got %s", tc.output.conflictType, conflict.ConflictType)
			}
		})
	}
}
