package gignore

import (
	"strings"
	"testing"
)

type FakeRepository struct {
	files map[string]string
}

func (f *FakeRepository) Load(path string, ignoreFile *IgnoreFile) error {
	content, ok := f.files[path]
	if !ok {
		return fileReadError
	}

	return LoadFile(strings.NewReader(content), ignoreFile)
}

func (f *FakeRepository) Save(path string, ignoreFile *IgnoreFile) error {
	content := Render(ignoreFile, RenderOptions{})
	f.files[path] = content

	return nil
}

func NewFakeRepository() FakeRepository {
	return FakeRepository{
		files: make(map[string]string),
	}
}

// MARK: Helpers
func ruleExists(rules IgnoreFile, toCheck Ruler) bool {
	for _, rule := range rules.rules {
		if rulesEqual(rule, toCheck) {
			return true
		}
	}

	return false
}

func checkRuleExists(repo Repository, path string, expected Ruler, t *testing.T) {
	var loaded IgnoreFile
	err := repo.Load(path, &loaded)
	if err != nil {
		t.Errorf("unexpected error loading file for comparison %s", err.Error())
		return
	}

	if !ruleExists(loaded, expected) {
		t.Errorf("expected to find matching rule")
	}
}

func checkRuleDoesNotExist(repo Repository, path string, notExpected Ruler, t *testing.T) {
	var loaded IgnoreFile
	err := repo.Load(path, &loaded)
	if err != nil {
		t.Errorf("unexpected error loading file for comparison %s", err.Error())
		return
	}

	if ruleExists(loaded, notExpected) {
		t.Errorf("expected to not find matching rule")
	}
}

func checkRuleOrdering(repo Repository, path string, source, target Ruler, direction MoveDirection, t *testing.T) {
	var loaded IgnoreFile
	err := repo.Load(path, &loaded)
	if err != nil {
		t.Errorf("unexpected error loading file for comparison %s", err.Error())
		return
	}

	moveIdx := loaded.findRuleIndex(source)
	afterIdx := loaded.findRuleIndex(target)

	var expectedIdx int
	switch direction {
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
}

func checkErrors(expectedErrorMsg string, err error, t *testing.T) {
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	}

	if errMsg != expectedErrorMsg {
		t.Errorf("expected error %s, got %s", expectedErrorMsg, errMsg)
	}
}

func testServiceOperation(
	t *testing.T,
	initial IgnoreFile,
	path string, // "file" path
	operation func(*Service, string) error,
	errorMessage string,
	initRepo bool,
	comparisonFunc func(Repository, string, *testing.T),
) {
	repo := NewFakeRepository()
	svc := NewService(&repo)

	if initRepo {
		repo.Save(path, &initial)
	}

	err := operation(&svc, path)

	checkErrors(errorMessage, err, t)
	if errorMessage != "" {
		// Don't do validation checks if the error was not nil
		return
	}

	// Check if the rule was added
	comparisonFunc(&repo, path, t)
}

// MARK: Tests
func TestServiceInit(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		errorMessage string
	}{
		{
			name:         "Pass",
			path:         ".gitignore",
			errorMessage: "",
		},
	}

	for _, tc := range tests {
		repo := NewFakeRepository()
		svc := NewService(&repo)

		t.Run(tc.name, func(t *testing.T) {
			err := svc.Init(tc.path)

			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}

			if errMsg != tc.errorMessage {
				t.Errorf("expected error %s, got %s", tc.errorMessage, errMsg)
			}
		})
	}
}

func TestServiceAddFileRule(t *testing.T) {
	tests := []struct {
		name         string
		ignore       IgnoreFile
		path         string
		newRule      FileRule
		errorMessage string
		initRepo     bool
	}{
		{
			name: "Pass",
			ignore: IgnoreFile{
				rules: []Ruler{},
			},
			path: ".gitignore",
			newRule: FileRule{
				path: "todo.md",
				act:  INCLUDE,
			},
			errorMessage: "",
			initRepo:     true,
		},
		{
			name: "Fail-FileDoesNotExist",
			ignore: IgnoreFile{
				rules: []Ruler{},
			},
			path: ".gitignore",
			newRule: FileRule{
				path: "todo.md",
				act:  INCLUDE,
			},
			errorMessage: fileReadError.Error(),
			initRepo:     false,
		},
		{
			name: "Pass-AddAndAvoidConflicts",
			ignore: IgnoreFile{
				rules: []Ruler{
					DirectoryRule{
						name: "build",
						mode: RECURSIVE,
						act:  INCLUDE,
					},
				},
			},
			path: ".gitignore",
			newRule: FileRule{
				path: "build/important.txt",
				act:  EXCLUDE,
			},
			errorMessage: "",
			initRepo:     true,
		},
		{
			name: "Pass-AddAndFixConflicts",
			ignore: IgnoreFile{
				rules: []Ruler{
					DirectoryRule{
						name: "build",
						mode: RECURSIVE,
						act:  INCLUDE,
					},
				},
			},
			path: ".gitignore",
			newRule: FileRule{
				path: "build/important.txt",
				act:  EXCLUDE,
			},
			errorMessage: "",
			initRepo:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testServiceOperation(
				t,
				tc.ignore,
				tc.path,
				func(svc *Service, path string) error {
					_, err := svc.AddFileRule(path, tc.newRule.path, tc.newRule.act)
					return err
				},
				tc.errorMessage,
				tc.initRepo,
				func(repo Repository, path string, t *testing.T) {
					checkRuleExists(repo, path, tc.newRule, t)
				},
			)
		})
	}
}

func TestServiceAddExtensionRule(t *testing.T) {
	tests := []struct {
		name         string
		ignore       IgnoreFile
		path         string
		newRule      ExtensionRule
		errorMessage string
		initRepo     bool
	}{
		{
			name: "Pass",
			ignore: IgnoreFile{
				rules: []Ruler{},
			},
			path: ".gitignore",
			newRule: ExtensionRule{
				ext: "txt",
				act: INCLUDE,
			},
			errorMessage: "",
			initRepo:     true,
		},
		{
			name: "Fail-FileDoesNotExist",
			ignore: IgnoreFile{
				rules: []Ruler{},
			},
			path: ".gitignore",
			newRule: ExtensionRule{
				ext: "txt",
				act: INCLUDE,
			},
			errorMessage: fileReadError.Error(),
			initRepo:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testServiceOperation(
				t,
				tc.ignore,
				tc.path,
				func(svc *Service, path string) error {
					_, err := svc.AddExtensionRule(path, tc.newRule.ext, tc.newRule.act)
					return err
				},
				tc.errorMessage,
				tc.initRepo,
				func(repo Repository, path string, t *testing.T) {
					checkRuleExists(repo, path, tc.newRule, t)
				},
			)
		})
	}
}

func TestServiceAddDirectoryRule(t *testing.T) {
	tests := []struct {
		name         string
		ignore       IgnoreFile
		path         string
		newRule      DirectoryRule
		errorMessage string
		initRepo     bool
	}{
		{
			name: "Pass",
			ignore: IgnoreFile{
				rules: []Ruler{},
			},
			path: ".gitignore",
			newRule: DirectoryRule{
				name: "build",
				mode: DIRECTORY,
				act:  INCLUDE,
			},
			errorMessage: "",
			initRepo:     true,
		},
		{
			name: "Pass-AndFixConflict",
			ignore: IgnoreFile{
				rules: []Ruler{
					FileRule{
						path: "build/important.txt",
						act:  EXCLUDE,
					},
					FileRule{
						path: "logs/important",
						act:  INCLUDE,
					},
					DirectoryRule{
						name: "logs",
						mode: DIRECTORY,
						act:  INCLUDE,
					},
				},
			},
			path: ".gitignore",
			newRule: DirectoryRule{
				name: "build",
				mode: DIRECTORY,
				act:  INCLUDE,
			},
			errorMessage: "",
			initRepo:     true,
		},
		{
			name: "Fail-Conflict",
			ignore: IgnoreFile{
				rules: []Ruler{
					DirectoryRule{
						name: "build",
						mode: RECURSIVE,
						act:  INCLUDE,
					},
				},
			},
			path: ".gitignore",
			newRule: DirectoryRule{
				name: "build",
				mode: DIRECTORY,
				act:  INCLUDE,
			},
			errorMessage: unreachableRuleError.Error(),
			initRepo:     true,
		},
		{
			name: "Fail-FileDoesNotExist",
			ignore: IgnoreFile{
				rules: []Ruler{},
			},
			path: ".gitignore",
			newRule: DirectoryRule{
				name: "build",
				mode: DIRECTORY,
				act:  INCLUDE,
			},
			errorMessage: fileReadError.Error(),
			initRepo:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testServiceOperation(
				t,
				tc.ignore,
				tc.path,
				func(svc *Service, path string) error {
					_, err := svc.AddDirectoryRule(path, tc.newRule.name, tc.newRule.mode, tc.newRule.act)
					return err
				},
				tc.errorMessage,
				tc.initRepo,
				func(repo Repository, path string, t *testing.T) {
					checkRuleExists(repo, path, tc.newRule, t)
				},
			)
		})
	}
}

func TestServiceAddGlobRule(t *testing.T) {
	tests := []struct {
		name         string
		ignore       IgnoreFile
		path         string
		newRule      GlobRule
		errorMessage string
		initRepo     bool
	}{
		{
			name: "Pass",
			ignore: IgnoreFile{
				rules: []Ruler{},
			},
			path: ".gitignore",
			newRule: GlobRule{
				pattern: "buildlog*.txt",
				act:     INCLUDE,
			},
			errorMessage: "",
			initRepo:     true,
		},
		{
			name: "Fail-FileDoesNotExist",
			ignore: IgnoreFile{
				rules: []Ruler{},
			},
			path: ".gitignore",
			newRule: GlobRule{
				pattern: "buildlog*.txt",
				act:     INCLUDE,
			},
			errorMessage: fileReadError.Error(),
			initRepo:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testServiceOperation(
				t,
				tc.ignore,
				tc.path,
				func(svc *Service, path string) error {
					_, err := svc.AddGlobRule(path, tc.newRule.pattern, tc.newRule.act)
					return err
				},
				tc.errorMessage,
				tc.initRepo,
				func(repo Repository, path string, t *testing.T) {
					checkRuleExists(repo, path, tc.newRule, t)
				},
			)
		})
	}
}

func TestServiceDeleteFileRule(t *testing.T) {
	tests := []struct {
		name         string
		ignore       IgnoreFile
		path         string
		toRemove     FileRule
		errorMessage string
		initRepo     bool
	}{
		{
			name: "Pass",
			ignore: IgnoreFile{
				rules: []Ruler{
					FileRule{
						path: "buildlog.txt",
						act:  INCLUDE,
					},
				},
			},
			path: ".gitignore",
			toRemove: FileRule{
				path: "buildlog.txt",
				act:  INCLUDE,
			},
			errorMessage: "",
			initRepo:     true,
		},
		{
			name: "Fail-FileNotFound",
			ignore: IgnoreFile{
				rules: []Ruler{
					FileRule{
						path: "buildlog.txt",
						act:  INCLUDE,
					},
				},
			},
			path: ".gitignore",
			toRemove: FileRule{
				path: "buildlog.txt",
				act:  INCLUDE,
			},
			errorMessage: fileReadError.Error(),
			initRepo:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testServiceOperation(
				t,
				tc.ignore,
				tc.path,
				func(svc *Service, path string) error {
					_, err := svc.DeleteFileRule(path, tc.toRemove.path, tc.toRemove.act)
					return err
				},
				tc.errorMessage,
				tc.initRepo,
				func(repo Repository, path string, t *testing.T) {
					checkRuleDoesNotExist(repo, path, tc.toRemove, t)
				},
			)
		})
	}
}

func TestServiceDeleteExtensionRule(t *testing.T) {
	tests := []struct {
		name         string
		ignore       IgnoreFile
		path         string
		toRemove     ExtensionRule
		errorMessage string
		initRepo     bool
	}{
		{
			name: "Pass",
			ignore: IgnoreFile{
				rules: []Ruler{
					ExtensionRule{
						ext: "txt",
						act: INCLUDE,
					},
				},
			},
			path: ".gitignore",
			toRemove: ExtensionRule{
				ext: "txt",
				act: INCLUDE,
			},
			errorMessage: "",
			initRepo:     true,
		},
		{
			name: "Fail-FileNotFound",
			ignore: IgnoreFile{
				rules: []Ruler{
					ExtensionRule{
						ext: "txt",
						act: INCLUDE,
					},
				},
			},
			path: ".gitignore",
			toRemove: ExtensionRule{
				ext: "txt",
				act: INCLUDE,
			},
			errorMessage: fileReadError.Error(),
			initRepo:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testServiceOperation(
				t,
				tc.ignore,
				tc.path,
				func(svc *Service, path string) error {
					_, err := svc.DeleteExtensionRule(path, tc.toRemove.ext, tc.toRemove.act)
					return err
				},
				tc.errorMessage,
				tc.initRepo,
				func(repo Repository, path string, t *testing.T) {
					checkRuleDoesNotExist(repo, path, tc.toRemove, t)
				},
			)
		})
	}
}

func TestServiceDeleteDirectoryRule(t *testing.T) {
	tests := []struct {
		name         string
		ignore       IgnoreFile
		path         string
		toRemove     DirectoryRule
		errorMessage string
		initRepo     bool
	}{
		{
			name: "Pass",
			ignore: IgnoreFile{
				rules: []Ruler{
					DirectoryRule{
						name: "build",
						mode: DIRECTORY,
						act:  INCLUDE,
					},
				},
			},
			path: ".gitignore",
			toRemove: DirectoryRule{
				name: "build",
				mode: DIRECTORY,
				act:  INCLUDE,
			},
			errorMessage: "",
			initRepo:     true,
		},
		{
			name: "Fail-FileNotFound",
			ignore: IgnoreFile{
				rules: []Ruler{
					DirectoryRule{
						name: "build",
						mode: DIRECTORY,
						act:  INCLUDE,
					},
				},
			},
			path: ".gitignore",
			toRemove: DirectoryRule{
				name: "build",
				mode: DIRECTORY,
				act:  INCLUDE,
			},
			errorMessage: fileReadError.Error(),
			initRepo:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testServiceOperation(
				t,
				tc.ignore,
				tc.path,
				func(svc *Service, path string) error {
					_, err := svc.DeleteDirectoryRule(path, tc.toRemove.name, tc.toRemove.mode, tc.toRemove.act)
					return err
				},
				tc.errorMessage,
				tc.initRepo,
				func(repo Repository, path string, t *testing.T) {
					checkRuleDoesNotExist(repo, path, tc.toRemove, t)
				},
			)
		})
	}
}

func TestServiceDeleteGlobRule(t *testing.T) {
	tests := []struct {
		name         string
		ignore       IgnoreFile
		path         string
		toRemove     GlobRule
		errorMessage string
		initRepo     bool
	}{
		{
			name: "Pass",
			ignore: IgnoreFile{
				rules: []Ruler{
					GlobRule{
						pattern: "build*.txt",
						act:     INCLUDE,
					},
				},
			},
			path: ".gitignore",
			toRemove: GlobRule{
				pattern: "build*.txt",
				act:     INCLUDE,
			},
			errorMessage: "",
			initRepo:     true,
		},
		{
			name: "Fail-FileNotFound",
			ignore: IgnoreFile{
				rules: []Ruler{
					GlobRule{
						pattern: "build*.txt",
						act:     INCLUDE,
					},
				},
			},
			path: ".gitignore",
			toRemove: GlobRule{
				pattern: "build*.txt",
				act:     INCLUDE,
			},
			errorMessage: fileReadError.Error(),
			initRepo:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testServiceOperation(
				t,
				tc.ignore,
				tc.path,
				func(svc *Service, path string) error {
					_, err := svc.DeleteGlobRule(path, tc.toRemove.pattern, tc.toRemove.act)
					return err
				},
				tc.errorMessage,
				tc.initRepo,
				func(repo Repository, path string, t *testing.T) {
					checkRuleDoesNotExist(repo, path, tc.toRemove, t)
				},
			)
		})
	}
}

func TestServiceAnalyzeConflicts(t *testing.T) {
	tests := []struct {
		name          string
		ignore        IgnoreFile
		path          string
		initRepo      bool
		conflictCount int
		errorMessage  string
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
			path:          ".gitignore",
			initRepo:      true,
			conflictCount: 0,
			errorMessage:  "",
		},
		{
			name: "Pass-Empty",
			ignore: IgnoreFile{
				rules: []Ruler{},
			},
			path:          ".gitignore",
			initRepo:      true,
			conflictCount: 0,
			errorMessage:  "",
		},
		{
			name: "Pass-SomeConflicts",
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
			path:          ".gitignore",
			initRepo:      true,
			conflictCount: 1,
			errorMessage:  "",
		},
		{
			name:          "Fail-FileNotFound",
			ignore:        IgnoreFile{},
			path:          ".gitignore",
			initRepo:      false,
			conflictCount: 0,
			errorMessage:  fileReadError.Error(),
		},
	}

	for _, tc := range tests {
		repo := NewFakeRepository()
		svc := NewService(&repo)

		t.Run(tc.name, func(t *testing.T) {
			if tc.initRepo {
				repo.Save(tc.path, &tc.ignore)
			}

			conflicts, err := svc.AnalyzeConflicts(tc.path)

			checkErrors(tc.errorMessage, err, t)
			if tc.errorMessage != "" {
				// Don't do validation checks if the error was not nil
				return
			}

			if len(conflicts) != tc.conflictCount {
				t.Errorf("expected %d conflicts, found %d", tc.conflictCount, len(conflicts))
			}
		})
	}
}

func TestServiceMoveRule(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		initRepo     bool
		ignore       IgnoreFile
		ruleToMove   Ruler
		targetRule   Ruler
		direction    MoveDirection
		errorMessage string
	}{
		{
			name:     "Pass-Simple",
			path:     ".gitignore",
			initRepo: true,
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
			name:     "Pass-Complicated",
			path:     ".gitignore",
			initRepo: true,
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
			name:     "Fail-FileNotFound",
			path:     ".gitignore",
			initRepo: false,
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
			errorMessage: fileReadError.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testServiceOperation(
				t,
				tc.ignore,
				tc.path,
				func(svc *Service, path string) error {
					_, err := svc.MoveRule(path, tc.ruleToMove.Render(), tc.targetRule.Render(), tc.direction)
					return err
				},
				tc.errorMessage,
				tc.initRepo,
				func(repo Repository, path string, t *testing.T) {
					checkRuleOrdering(repo, path, tc.ruleToMove, tc.targetRule, tc.direction, t)
				},
			)
		})
	}
}

func TestServiceAutoFix(t *testing.T) {
	tests := []struct {
		name              string
		path              string
		initRepo          bool
		maxFixes          int
		ignore            IgnoreFile
		expectedConflicts int
		errorMessage      string
	}{
		{
			name:     "Pass-RedundantRules",
			path:     ".gitignore",
			initRepo: true,
			maxFixes: 10,
			ignore: IgnoreFile{
				rules: []Ruler{
					FileRule{path: "config.json", act: INCLUDE},
					FileRule{path: "config.json", act: INCLUDE}, // duplicate
					ExtensionRule{ext: "log", act: INCLUDE},
				},
			},
			expectedConflicts: 0,
			errorMessage:      "",
		},
		{
			name:     "Pass-UnreachableRule",
			path:     ".gitignore",
			initRepo: true,
			maxFixes: 10,
			ignore: IgnoreFile{
				rules: []Ruler{
					DirectoryRule{name: "build", mode: RECURSIVE, act: INCLUDE}, // build/**
					DirectoryRule{name: "build", mode: DIRECTORY, act: INCLUDE}, // build/ - unreachable
				},
			},
			expectedConflicts: 0,
			errorMessage:      "",
		},
		{
			name:     "Pass-MultipleConflicts",
			path:     ".gitignore",
			initRepo: true,
			maxFixes: 10,
			ignore: IgnoreFile{
				rules: []Ruler{
					ExtensionRule{ext: "log", act: INCLUDE},
					ExtensionRule{ext: "log", act: INCLUDE},                     // duplicate
					FileRule{path: "build/important.txt", act: EXCLUDE},         // should come after build/**
					DirectoryRule{name: "build", mode: RECURSIVE, act: INCLUDE}, // build/**
					DirectoryRule{name: "build", mode: DIRECTORY, act: INCLUDE}, // unreachable
				},
			},
			expectedConflicts: 0,
			errorMessage:      "",
		},
		{
			name:     "Pass-SemanticConflict-NoChange",
			path:     ".gitignore",
			initRepo: true,
			maxFixes: 10,
			ignore: IgnoreFile{
				rules: []Ruler{
					FileRule{path: "config.json", act: INCLUDE},
					FileRule{path: "config.json", act: EXCLUDE}, // semantic conflict - don't auto-fix
				},
			},
			expectedConflicts: 1,
			errorMessage:      "",
		},
		{
			name:              "Fail-FileNotFound",
			path:              ".gitignore",
			initRepo:          false,
			maxFixes:          10,
			ignore:            IgnoreFile{},
			expectedConflicts: 0,
			errorMessage:      fileReadError.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testServiceOperation(
				t,
				tc.ignore,
				tc.path,
				func(svc *Service, path string) error {
					_, err := svc.AutoFix(tc.path, tc.maxFixes)
					return err
				},
				tc.errorMessage,
				tc.initRepo,
				func(repo Repository, path string, t *testing.T) {
					var loaded IgnoreFile
					err := repo.Load(path, &loaded)
					if err != nil {
						t.Errorf("unexpected error loading file for comparison %s", err.Error())
						return
					}

					if foundConflicts := loaded.FindConflicts(); len(foundConflicts) > tc.expectedConflicts {
						t.Errorf("expected 0 conflicts, found %d", len(foundConflicts))
					}
				},
			)
		})
	}
}
