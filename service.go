package gignore

type Repository interface {
	Load(path string, ignoreFile *IgnoreFile) error
	Save(path string, ignoreFile *IgnoreFile) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return Service{repo: repo}
}

// Creates a new ignore file
func (s *Service) Init(path string) error {
	ignore := NewIgnoreFile()

	return s.repo.Save(path, &ignore)
}

// MARK: Add methods

// AddFileRule adds a new file rule to an ignore file using an atomic load-modify-save operation.
// The method loads the ignore file from the specified path, adds the file rule with automatic
// conflict detection and resolution, then saves the updated file.
//
// Parameters:
//   - path: The file system path to the ignore file to modify.
//   - filePath: The file system path for the new rule. The path will be validated and cleaned.
//   - action: The action to be performed when the rule matches. Must be either INCLUDE or EXCLUDE.
//
// Returns a slice of Result containing the addition operation and any subsequent conflict
// fixes, plus an error. The error will be non-nil if:
//   - The ignore file cannot be loaded from the specified path
//   - The provided filePath is invalid or cannot be cleaned
//   - The provided action fails validation
//   - A semantic conflict, redundant rule, or unreachable rule is detected
//   - Automatic conflict resolution fails
//   - The updated ignore file cannot be saved
//
// Example:
//
//	results, err := service.AddFileRule(".gitignore", "config.json", EXCLUDE)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	for _, result := range results {
//	    fmt.Printf("Operation: %s\n", result.Log())
//	}
func (s *Service) AddFileRule(path, filePath string, action Action) ([]Result, error) {
	var results []Result

	err := s.loadModifySave(path, func(ignoreFile *IgnoreFile) error {
		var err error
		results, err = ignoreFile.AddFile(filePath, action)
		return err
	})

	return results, err
}

// AddExtensionRule adds a new extension rule to an ignore file using an atomic load-modify-save operation.
// The method loads the ignore file from the specified path, adds the extension rule with automatic
// conflict detection and resolution, then saves the updated file.
//
// Parameters:
//   - path: The file system path to the ignore file to modify.
//   - ext: The file extension for the new rule (e.g., "txt", "*.go", ".json").
//     Whitespace and "*." prefixes will be trimmed automatically.
//   - action: The action to be performed when the rule matches. Must be either INCLUDE or EXCLUDE.
//
// Returns a slice of Result containing the addition operation and any subsequent conflict
// fixes, plus an error. The error will be non-nil if:
//   - The ignore file cannot be loaded from the specified path
//   - The provided extension is empty after cleaning (whitespace and prefixes removed)
//   - The provided action fails validation
//   - A semantic conflict, redundant rule, or unreachable rule is detected
//   - Automatic conflict resolution fails
//   - The updated ignore file cannot be saved
//
// Example:
//
//	results, err := service.AddExtensionRule(".gitignore", "*.log", EXCLUDE)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// These are all equivalent:
//	service.AddExtensionRule(".gitignore", "go", INCLUDE)
//	service.AddExtensionRule(".gitignore", ".go", INCLUDE)
//	service.AddExtensionRule(".gitignore", "*.go", INCLUDE)
func (s *Service) AddExtensionRule(path, ext string, action Action) ([]Result, error) {
	var results []Result

	err := s.loadModifySave(path, func(ignoreFile *IgnoreFile) error {
		var err error
		results, err = ignoreFile.AddExtension(ext, action)
		return err
	})

	return results, err
}

// AddDirectoryRule adds a new directory rule to an ignore file using an atomic load-modify-save operation.
// The method loads the ignore file from the specified path, adds the directory rule with automatic
// conflict detection and resolution, then saves the updated file.
//
// Parameters:
//   - path: The file system path to the ignore file to modify.
//   - name: The directory path for the new rule (e.g., "node_modules", "build/", "/src").
//     Whitespace and leading/trailing slashes will be trimmed automatically.
//   - mode: The directory matching mode that determines how the rule is applied:
//     DIRECTORY matches the directory itself (nameofdir/),
//     CHILDREN matches direct children (nameofdir/*),
//     RECURSIVE matches all descendants (nameofdir/**),
//     ANYWHERE matches the directory at any depth (**/nameofdir),
//     ROOT_ONLY matches only at the root level (/build or /build/).
//   - action: The action to be performed when the rule matches. Must be either INCLUDE or EXCLUDE.
//
// Returns a slice of Result containing the addition operation and any subsequent conflict
// fixes, plus an error. The error will be non-nil if:
//   - The ignore file cannot be loaded from the specified path
//   - The provided directory path is empty after cleaning (whitespace and slashes removed)
//   - The provided mode fails validation
//   - The provided action fails validation
//   - A semantic conflict, redundant rule, or unreachable rule is detected
//   - Automatic conflict resolution fails
//   - The updated ignore file cannot be saved
//
// Example:
//
//	results, err := service.AddDirectoryRule(".gitignore", "node_modules", RECURSIVE, EXCLUDE)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Match only the build directory at root level
//	results, err = service.AddDirectoryRule(".gitignore", "build", ROOT_ONLY, EXCLUDE)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (s *Service) AddDirectoryRule(path, name string, mode DirectoryMode, action Action) ([]Result, error) {
	var results []Result

	err := s.loadModifySave(path, func(ignoreFile *IgnoreFile) error {
		var err error
		results, err = ignoreFile.AddDirectory(name, mode, action)
		return err
	})

	return results, err
}

// AddGlobRule adds a new glob rule to an ignore file using an atomic load-modify-save operation.
// The method loads the ignore file from the specified path, adds the glob rule with automatic
// conflict detection and resolution, then saves the updated file.
//
// Parameters:
//   - path: The file system path to the ignore file to modify.
//   - pattern: The glob pattern for the new rule (e.g., "*.tmp", "test/**/*.go", "**/node_modules/**").
//     Whitespace will be trimmed automatically.
//   - action: The action to be performed when the rule matches. Must be either INCLUDE or EXCLUDE.
//
// Returns a slice of Result containing the addition operation and any subsequent conflict
// fixes, plus an error. The error will be non-nil if:
//   - The ignore file cannot be loaded from the specified path
//   - The provided pattern is empty after cleaning (whitespace removed)
//   - The provided action fails validation
//   - A semantic conflict, redundant rule, or unreachable rule is detected
//   - Automatic conflict resolution fails
//   - The updated ignore file cannot be saved
//
// Example:
//
//	results, err := service.AddGlobRule(".gitignore", "*.log", EXCLUDE)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Match all Go test files recursively
//	results, err = service.AddGlobRule(".gitignore", "**/*_test.go", INCLUDE)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (s *Service) AddGlobRule(path, pattern string, action Action) ([]Result, error) {
	var results []Result

	err := s.loadModifySave(path, func(ignoreFile *IgnoreFile) error {
		var err error
		results, err = ignoreFile.AddGlob(pattern, action)
		return err
	})

	return results, err
}

// MARK: Remove methods

// DeleteFileRule removes a file rule from an ignore file using an atomic load-modify-save operation.
// The method loads the ignore file from the specified path, removes the first rule that exactly
// matches the specified file path and action, then saves the updated file.
//
// Parameters:
//   - path: The file system path to the ignore file to modify.
//   - filePath: The file system path for the rule to delete. The path will be validated and cleaned
//     using the same process as AddFileRule to ensure accurate matching.
//   - action: The action of the rule to delete. Must be either INCLUDE or EXCLUDE.
//
// Returns a Result containing details of the deletion operation and an error.
// The error will be non-nil if:
//   - The ignore file cannot be loaded from the specified path
//   - The provided filePath is invalid or cannot be cleaned
//   - The provided action fails validation
//   - No matching rule is found in the ignore file
//   - The updated ignore file cannot be saved
//
// Example:
//
//	result, err := service.DeleteFileRule(".gitignore", "config.json", EXCLUDE)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Deleted rule: %s\n", result.Log())
func (s *Service) DeleteFileRule(path, filePath string, action Action) (Result, error) {
	var results Result

	err := s.loadModifySave(path, func(ignoreFile *IgnoreFile) error {
		var err error
		results, err = ignoreFile.DeleteFile(filePath, action)
		return err
	})

	return results, err
}

// DeleteExtensionRule removes an extension rule from an ignore file using an atomic load-modify-save operation.
// The method loads the ignore file from the specified path, removes the first rule that exactly
// matches the specified extension and action, then saves the updated file.
//
// Parameters:
//   - path: The file system path to the ignore file to modify.
//   - ext: The file extension for the rule to delete (e.g., "txt", "*.go", ".json").
//     Whitespace and "*." prefixes will be trimmed automatically for matching using the same
//     process as AddExtensionRule to ensure accurate matching.
//   - action: The action of the rule to delete. Must be either INCLUDE or EXCLUDE.
//
// Returns a Result containing details of the deletion operation and an error.
// The error will be non-nil if:
//   - The ignore file cannot be loaded from the specified path
//   - The provided extension is empty after cleaning (whitespace and prefixes removed)
//   - The provided action fails validation
//   - No matching rule is found in the ignore file
//   - The updated ignore file cannot be saved
//
// Example:
//
//	result, err := service.DeleteExtensionRule(".gitignore", "*.log", EXCLUDE)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// These are all equivalent for matching:
//	service.DeleteExtensionRule(".gitignore", "go", INCLUDE)
//	service.DeleteExtensionRule(".gitignore", ".go", INCLUDE)
//	service.DeleteExtensionRule(".gitignore", "*.go", INCLUDE)
func (s *Service) DeleteExtensionRule(path, ext string, action Action) (Result, error) {
	var results Result

	err := s.loadModifySave(path, func(ignoreFile *IgnoreFile) error {
		var err error
		results, err = ignoreFile.DeleteExtension(ext, action)
		return err
	})

	return results, err
}

// DeleteDirectoryRule removes a directory rule from an ignore file using an atomic load-modify-save operation.
// The method loads the ignore file from the specified path, removes the first rule that exactly
// matches the specified directory name, mode, and action, then saves the updated file.
//
// Parameters:
//   - path: The file system path to the ignore file to modify.
//   - name: The directory path for the rule to delete (e.g., "node_modules", "build/", "/src").
//     Whitespace and trailing slashes will be trimmed automatically for matching using the same
//     process as AddDirectoryRule to ensure accurate matching.
//   - mode: The directory matching mode of the rule to delete:
//     DIRECTORY matches the directory itself (nameofdir/),
//     CHILDREN matches direct children (nameofdir/*),
//     RECURSIVE matches all descendants (nameofdir/**),
//     ANYWHERE matches the directory at any depth (**/nameofdir),
//     ROOT_ONLY matches only at the root level (/build or /build/).
//   - action: The action of the rule to delete. Must be either INCLUDE or EXCLUDE.
//
// Returns a Result containing details of the deletion operation and an error.
// The error will be non-nil if:
//   - The ignore file cannot be loaded from the specified path
//   - The provided directory path is empty after cleaning (whitespace and slashes removed)
//   - The provided mode fails validation
//   - The provided action fails validation
//   - No matching rule is found in the ignore file
//   - The updated ignore file cannot be saved
//
// Example:
//
//	result, err := service.DeleteDirectoryRule(".gitignore", "node_modules", RECURSIVE, EXCLUDE)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// These are equivalent for matching (trailing slash is stripped):
//	service.DeleteDirectoryRule(".gitignore", "build", ROOT_ONLY, EXCLUDE)
//	service.DeleteDirectoryRule(".gitignore", "build/", ROOT_ONLY, EXCLUDE)
func (s *Service) DeleteDirectoryRule(path, name string, mode DirectoryMode, action Action) (Result, error) {
	var results Result

	err := s.loadModifySave(path, func(ignoreFile *IgnoreFile) error {
		var err error
		results, err = ignoreFile.DeleteDirectory(name, mode, action)
		return err
	})

	return results, err
}

// DeleteGlobRule removes a glob rule from an ignore file using an atomic load-modify-save operation.
// The method loads the ignore file from the specified path, removes the first rule that exactly
// matches the specified glob pattern and action, then saves the updated file.
//
// Parameters:
//   - path: The file system path to the ignore file to modify.
//   - pattern: The glob pattern for the rule to delete (e.g., "test/**/*.go", "**/node_modules/**").
//     Whitespace will be trimmed automatically for matching using the same process as AddGlobRule
//     to ensure accurate matching.
//   - action: The action of the rule to delete. Must be either INCLUDE or EXCLUDE.
//
// Returns a Result containing details of the deletion operation and an error.
// The error will be non-nil if:
//   - The ignore file cannot be loaded from the specified path
//   - The provided pattern is empty after cleaning (whitespace removed)
//   - The provided action fails validation
//   - No matching rule is found in the ignore file
//   - The updated ignore file cannot be saved
//
// Example:
//
//	// Delete a specific test file pattern
//	result, err = service.DeleteGlobRule(".gitignore", "**/*_test.go", INCLUDE)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (s *Service) DeleteGlobRule(path, pattern string, action Action) (Result, error) {
	var results Result

	err := s.loadModifySave(path, func(ignoreFile *IgnoreFile) error {
		var err error
		results, err = ignoreFile.DeleteGlob(pattern, action)
		return err
	})

	return results, err
}

// MARK: Move methods

// MoveRule relocates an existing rule relative to another rule in an ignore file using an atomic load-modify-save operation.
// The method loads the ignore file from the specified path, parses both rule patterns to find the corresponding rules,
// moves the first rule to the specified position relative to the target rule, then saves the updated file.
//
// Parameters:
//   - path: The file system path to the ignore file to modify.
//   - rulePattern: The pattern string representing the rule to be moved (e.g., "*.log", "node_modules/").
//     The pattern will be parsed to determine rule type and must exactly match an existing rule.
//   - targetRulePattern: The pattern string representing the target rule that defines the destination position.
//     The pattern will be parsed and must exactly match an existing rule.
//   - direction: The position relative to the target rule. Must be either BEFORE or AFTER.
//
// Returns a Result containing details of the move operation and an error. The error will be non-nil if:
//   - The ignore file cannot be loaded from the specified path
//   - Either rule pattern cannot be parsed or is invalid
//   - The rule to move is not found in the ignore file
//   - The target rule is not found in the ignore file
//   - The direction parameter is invalid
//   - The calculated new position is out of range
//   - The updated ignore file cannot be saved
//
// The method includes optimizations to avoid unnecessary moves when rules are already in the correct position.
//
// Example:
//
//	// Move a log exclusion rule to appear before a directory rule
//	result, err := service.MoveRule(".gitignore", "*.log", "node_modules/", BEFORE)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if result.Result == MOVED {
//	    fmt.Printf("Successfully moved rule: %s\n", result.Log())
//	}
func (s *Service) MoveRule(path, rulePattern, targetRulePattern string, direction MoveDirection) (Result, error) {
	var results Result

	err := s.loadModifySave(path, func(f *IgnoreFile) error {
		ruleToMove, err := parseRule(rulePattern)
		if err != nil {
			return err
		}

		targetRule, err := parseRule(targetRulePattern)
		if err != nil {
			return err
		}

		results, err = f.MoveRule(ruleToMove, targetRule, direction, REQUESTED)
		return err
	})

	return results, err
}

// MARK: Fixers

// AutoFix automatically resolves conflicts in an ignore file using an atomic load-modify-save operation.
// The method loads the ignore file from the specified path, runs multiple passes of conflict detection
// and resolution, then saves the updated file.
//
// Parameters:
//   - path: The file system path to the ignore file to analyze and fix.
//   - maxPasses: The maximum number of conflict resolution passes to attempt. Each pass finds all
//     current conflicts and attempts to fix them before checking again. The method will stop early
//     if no conflicts are found in a given pass.
//
// Returns a slice of Result containing descriptions of all fixes applied and an error.
// The error will be non-nil if:
//   - The ignore file cannot be loaded from the specified path
//   - Any conflict resolution attempt fails
//   - The updated ignore file cannot be saved
//
// Example:
//
//	fixes, err := service.AutoFix(".gitignore", 5)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	if len(fixes) > 0 {
//	    fmt.Printf("Applied %d fixes:\n", len(fixes))
//	    for _, fix := range fixes {
//	        fmt.Printf("- %s\n", fix.Log())
//	    }
//	} else {
//	    fmt.Println("No conflicts found")
//	}
func (s *Service) AutoFix(path string, maxPasses int) ([]Result, error) {
	var fixes []Result

	err := s.loadModifySave(path, func(f *IgnoreFile) error {
		var err error
		fixes, err = f.FixConflicts(maxPasses)
		return err
	})

	return fixes, err
}

// MARK: Analyzers

// AnalyzeConflicts loads an ignore file and returns all detected conflicts without making any modifications.
// This method provides a read-only analysis of rule conflicts, useful for inspection and reporting
// before deciding whether to apply automatic fixes.
//
// Parameters:
//   - path: The file system path to the ignore file to analyze.
//
// Returns a slice of Conflict containing all detected conflicts and an error.
// The error will be non-nil if:
//   - The ignore file cannot be loaded from the specified path
//
// Unlike AutoFix, this method performs no modifications to the file - it only reads and analyzes.
// The conflicts are detected using comprehensive pairwise comparison of all rules, considering
// rule order and any intervening rules between each pair.
//
// Conflict types that may be detected:
//   - SEMANTIC_CONFLICT: Rules that contradict each other in meaning
//   - REDUNDANT_RULE: Rules that duplicate existing functionality
//   - UNREACHABLE_RULE: Rules that can never be triggered due to earlier rules
//   - INEFFECTIVE_RULE: Rules that would be more effective in a different position
//
// Example:
//
//	conflicts, err := service.AnalyzeConflicts(".gitignore")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	if len(conflicts) > 0 {
//	    fmt.Printf("Found %d conflicts:\n", len(conflicts))
//	    for _, conflict := range conflicts {
//	        fmt.Printf("- %s\n", conflict)
//	    }
//	} else {
//	    fmt.Println("No conflicts detected")
//	}
func (s *Service) AnalyzeConflicts(path string) ([]Conflict, error) {
	var ignoreFile IgnoreFile
	if err := s.repo.Load(path, &ignoreFile); err != nil {
		return nil, err
	}

	return ignoreFile.FindConflicts(), nil
}

// Helper to reduce duplication
func (s *Service) loadModifySave(path string, modify func(*IgnoreFile) error) error {
	var ignoreFile IgnoreFile

	err := s.repo.Load(path, &ignoreFile)
	if err != nil {
		return err
	}

	if err := modify(&ignoreFile); err != nil {
		return err
	}

	return s.repo.Save(path, &ignoreFile)
}
