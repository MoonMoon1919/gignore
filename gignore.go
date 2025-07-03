package gignore

import (
	"errors"
	"fmt"
	"strings"
)

// MARK: Errors
var (
	semanticConflictError = errors.New("semantic conflict: same pattern with opposite actions")
	redundantRuleError    = errors.New("redundant rule: duplicate pattern and action")
	unreachableRuleError  = errors.New("unreachable rule: broader pattern makes this rule meaningless")

	ruleNotFoundError       = errors.New("rule not found")
	targetRuleNotFoundError = errors.New("target rule not found")
	ruleToMoveNotFoundError = errors.New("rule to move not found")

	emptyGlobPatternError     = errors.New("glob pattern cannot be empty")
	emptyPathError            = errors.New("path cannot be empty")
	emptyExtensionError       = errors.New("extension cannot be empty")
	emptyDirectoryNameError   = errors.New("directory cannot be empty")
	invalidActionError        = errors.New("invalid action")
	invalidDirectoryModeError = errors.New("invalid directory mode")
	invalidateDirectionError  = errors.New("invalid direction")

	sourceIdxOutOfRangeError = errors.New("from index out of range")
	targetIdxOutofRangeError = errors.New("target index out of range")
)

// MARK: Actions
const EXCLUDE_PREFIX = "!"
const INCLUDE_PREFIX = ""

type Action int

const (
	INCLUDE Action = iota + 1
	EXCLUDE
)

func ActionFromString(action string) (Action, error) {
	switch action {
	case "include":
		return INCLUDE, nil
	case "exclude":
		return EXCLUDE, nil
	default:
		return Action(0), invalidActionError
	}
}

func (a Action) Validate() error {
	switch a {
	case INCLUDE, EXCLUDE:
		return nil
	default:
		return invalidActionError
	}
}

func (a Action) Prefix() string {
	switch a {
	case INCLUDE:
		return INCLUDE_PREFIX
	case EXCLUDE:
		return EXCLUDE_PREFIX
	}

	return INCLUDE_PREFIX // logic default - we want to ignore
}

// MARK: Rules
type Ruler interface {
	Render() string
	Action() Action
	Pattern() string
}

// MARK: Files
type FileRule struct {
	path string
	act  Action
}

func validatePath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", emptyPathError
	}

	return path, nil
}

// NewFileRule creates a new FileRule with the specified path and action.
// The path is validated and cleaned before creating the rule.
//
// Parameters:
//   - path: The file system path for the rule. The path will be validated and cleaned.
//   - act: The action to be performed when the rule is triggered. Must be either INCLUDE or EXCLUDE.
//
// Returns a FileRule and an error. The error will be non-nil if:
//   - The provided path is invalid or cannot be cleaned
//   - The provided action fails validation
//
// Example:
//
//	rule, err := NewFileRule("/path/to/file.txt", INCLUDE)
//	if err != nil {
//	    log.Fatal(err)
//	}
func NewFileRule(path string, act Action) (FileRule, error) {
	cleanPath, err := validatePath(path)
	if err != nil {
		return FileRule{}, err
	}

	if err := act.Validate(); err != nil {
		return FileRule{}, err
	}

	return FileRule{
		path: cleanPath,
		act:  act,
	}, nil
}

func (r FileRule) Render() string {
	return fmt.Sprintf("%s%s", r.act.Prefix(), r.path)
}

func (r FileRule) Action() Action {
	return r.act
}

func (r FileRule) Pattern() string {
	return r.path
}

// MARK: Extensions
type ExtensionRule struct {
	ext string
	act Action
}

func validateExtension(ext string) (string, error) {
	ext = strings.TrimSpace(ext)
	ext = strings.TrimPrefix(ext, "*.")
	if ext == "" {
		return "", emptyExtensionError
	}

	return ext, nil
}

// NewExtensionRule creates a new ExtensionRule with the specified file extension and action.
// The extension is validated and cleaned before creating the rule. Leading whitespace,
// "*." prefixes are automatically removed from the extension.
//
// Parameters:
//   - ext: The file extension for the rule (e.g., "txt", "*.go", ".json").
//     Whitespace and "*." prefixes will be trimmed automatically.
//   - act: The action to be performed when the rule is triggered. Must be either INCLUDE or EXCLUDE.
//
// Returns an ExtensionRule and an error. The error will be non-nil if:
//   - The provided extension is empty after cleaning (whitespace and prefixes removed)
//   - The provided action fails validation
//
// Example:
//
//	rule, err := NewExtensionRule("*.txt", EXCLUDE)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// These are all equivalent:
//	rule1, _ := NewExtensionRule("go", INCLUDE)
//	rule2, _ := NewExtensionRule(".go", INCLUDE)
//	rule3, _ := NewExtensionRule("*.go", INCLUDE)
func NewExtensionRule(ext string, act Action) (ExtensionRule, error) {
	cleanedExt, err := validateExtension(ext)
	if err != nil {
		return ExtensionRule{}, err
	}

	if err := act.Validate(); err != nil {
		return ExtensionRule{}, err
	}

	return ExtensionRule{
		ext: cleanedExt,
		act: act,
	}, nil
}

func (r ExtensionRule) Render() string {
	return fmt.Sprintf("%s*.%s", r.act.Prefix(), r.ext)
}

func (r ExtensionRule) Action() Action {
	return r.act
}

func (r ExtensionRule) Pattern() string {
	rendered := r.Render()

	if strings.HasPrefix(rendered, "!") {
		return rendered[1:]
	}

	return rendered
}

// MARK: Directories
type DirectoryMode int

const (
	DIRECTORY DirectoryMode = iota + 1 // nameofdir/
	CHILDREN                           // nameofdir/*
	RECURSIVE                          // nameofdir/**
	ANYWHERE                           // **/nameofdir
	ROOT_ONLY                          // /build or /build/ (root only)
)

func DirectoryModeFromString(mode string) (DirectoryMode, error) {
	switch mode {
	case "directory":
		return DIRECTORY, nil
	case "children":
		return CHILDREN, nil
	case "recursive":
		return RECURSIVE, nil
	case "anywhere":
		return ANYWHERE, nil
	case "root_only":
		return ROOT_ONLY, nil
	default:
		return DirectoryMode(0), invalidDirectoryModeError
	}
}

func (d DirectoryMode) Validate() error {
	switch d {
	case DIRECTORY, RECURSIVE, CHILDREN, ANYWHERE, ROOT_ONLY:
		return nil
	default:
		return invalidDirectoryModeError
	}
}

func (d DirectoryMode) Prefix() string {
	switch d {
	case ANYWHERE:
		return "**/"
	case ROOT_ONLY:
		return "/"
	default:
		return ""
	}
}

func (d DirectoryMode) Suffix() string {
	switch d {
	case DIRECTORY:
		return "/"
	case CHILDREN:
		return "/*"
	case RECURSIVE:
		return "/**"
	default:
		return ""
	}
}

func validateDirectoryName(name string) (string, error) {
	name = strings.TrimSpace(name)
	name = strings.TrimPrefix(name, "/") // strip leading slash
	name = strings.TrimSuffix(name, "/") // strip trailing slash
	if name == "" {
		return "", emptyDirectoryNameError
	}
	return name, nil
}

type DirectoryRule struct {
	name string
	mode DirectoryMode
	act  Action
}

// NewDirectoryRule creates a new DirectoryRule with the specified directory path, mode, and action.
// The directory path is validated and cleaned before creating the rule. Leading and trailing
// whitespace is removed, and trailing slashes are stripped from the path.
//
// Parameters:
//   - path: The directory path for the rule (e.g., "node_modules", "build/", "/src").
//     Whitespace, leading, and trailing slashes will be trimmed automatically.
//   - mode: The directory matching mode that determines how the rule is applied:
//     DIRECTORY matches the directory itself (nameofdir/),
//     CHILDREN matches direct children (nameofdir/*),
//     RECURSIVE matches all descendants (nameofdir/**),
//     ANYWHERE matches the directory at any depth (**/nameofdir),
//     ROOT_ONLY matches only at the root level (/build or /build/).
//   - act: The action to be performed when the rule is triggered. Must be either INCLUDE or EXCLUDE.
//
// Returns a DirectoryRule and an error. The error will be non-nil if:
//   - The provided directory path is empty after cleaning (whitespace and slashes removed)
//   - The provided mode fails validation
//   - The provided action fails validation
//
// Example:
//
//	rule, err := NewDirectoryRule("node_modules", RECURSIVE, EXCLUDE)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// These are equivalent (trailing slash is stripped):
//	rule1, _ := NewDirectoryRule("build", ROOT_ONLY, EXCLUDE)
//	rule2, _ := NewDirectoryRule("build/", ROOT_ONLY, EXCLUDE)
func NewDirectoryRule(path string, mode DirectoryMode, act Action) (DirectoryRule, error) {
	cleanedName, err := validateDirectoryName(path)
	if err != nil {
		return DirectoryRule{}, err
	}

	if err := mode.Validate(); err != nil {
		return DirectoryRule{}, err
	}

	if err := act.Validate(); err != nil {
		return DirectoryRule{}, err
	}

	return DirectoryRule{
		name: cleanedName,
		mode: mode,
		act:  act,
	}, nil
}

func (r DirectoryRule) Render() string {
	return fmt.Sprintf("%s%s%s%s",
		r.act.Prefix(),
		r.mode.Prefix(),
		r.name,
		r.mode.Suffix(),
	)
}

func (r DirectoryRule) Action() Action {
	return r.act
}

func (r DirectoryRule) Pattern() string {
	rendered := r.Render()

	if strings.HasPrefix(rendered, "!") {
		return rendered[1:]
	}

	return rendered
}

// MARK: Glob
type GlobRule struct {
	pattern string
	act     Action
}

func validateGlobPattern(pattern string) (string, error) {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return "", emptyGlobPatternError
	}
	return pattern, nil
}

// NewGlobRule creates a new GlobRule with the specified glob pattern and action.
// The glob pattern is validated and cleaned before creating the rule. Leading and trailing
// whitespace is removed from the pattern.
//
// Parameters:
//   - pattern: The glob pattern for the rule (e.g., "test/**/*.go", "**/node_modules/**").
//     Whitespace will be trimmed automatically.
//   - act: The action to be performed when the rule is triggered. Must be either INCLUDE or EXCLUDE.
//
// Returns a GlobRule and an error. The error will be non-nil if:
//   - The provided pattern is empty after cleaning (whitespace removed)
//   - The provided action fails validation
//
// Example:
//
//	rule, err := NewGlobRule("*.log", EXCLUDE)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Match all Go test files recursively
//	testRule, err := NewGlobRule("**/*_test.go", INCLUDE)
//	if err != nil {
//	    log.Fatal(err)
//	}
func NewGlobRule(pattern string, act Action) (GlobRule, error) {
	cleanPattern, err := validateGlobPattern(pattern)
	if err != nil {
		return GlobRule{}, err
	}

	if err := act.Validate(); err != nil {
		return GlobRule{}, err
	}

	return GlobRule{
		pattern: cleanPattern,
		act:     act,
	}, nil
}

func (r GlobRule) Render() string {
	return fmt.Sprintf("%s%s", r.act.Prefix(), r.pattern)
}

func (r GlobRule) Action() Action {
	return r.act
}

func (r GlobRule) Pattern() string {
	return r.pattern
}

// MARK: IGNORE FILE
func rulesEqual(left, right Ruler) bool {
	return left.Pattern() == right.Pattern() && left.Action() == right.Action()
}

type IgnoreFile struct {
	rules []Ruler
}

func NewIgnoreFile() IgnoreFile {
	return IgnoreFile{rules: make([]Ruler, 0)}
}

// Adds a rule - used in parser
// Skips all validation! Only use when you can relax that constraint
func (f *IgnoreFile) addRule(rule Ruler) {
	f.rules = append(f.rules, rule)
}

func (f *IgnoreFile) findRuleIndex(target Ruler) int {
	for i, rule := range f.rules {
		if rulesEqual(rule, target) {
			return i
		}
	}

	return -1
}

func (f *IgnoreFile) ruleShouldComeBefore(newRule, existing Ruler) bool {
	if newRule.Action() == existing.Action() && subsumes(newRule, existing) {
		return true
	}

	return false
}

func (f *IgnoreFile) fixConflict(conflict Conflict) (Result, error) {
	switch conflict.ConflictType {
	case REDUNDANT_RULE:
		return f.deleteMatchingRule(conflict.Left, AUTOMATED_FIX)
	case UNREACHABLE_RULE:
		leftIdx := f.findRuleIndex(conflict.Left)
		rightIdx := f.findRuleIndex(conflict.Right)

		if rightIdx == leftIdx+1 {
			// If they're adjacent and specific rule comes after broader rule,
			// the specific rule is truly unreachable - remove it
			return f.deleteMatchingRule(conflict.Right, AUTOMATED_FIX)
		}

		return f.MoveRule(conflict.Right, conflict.Left, AFTER, AUTOMATED_FIX)
	case SEMANTIC_CONFLICT:
		return Result{
			Rule:   conflict.Left,
			Result: REVIEW_RECOMMENDED,
			Reason: FIX_UNKNOWN,
		}, nil
	case INEFFECTIVE_RULE:
		return f.MoveRule(conflict.Left, conflict.Right, AFTER, AUTOMATED_FIX)
	default:
		return Result{}, nil
	}
}

// FixConflicts attempts to automatically resolve conflicts within the IgnoreFile by running
// multiple passes of conflict detection and resolution. The method will stop early if no
// conflicts are found in a given pass.
//
// Parameters:
//   - maxPasses: The maximum number of conflict resolution passes to attempt. Each pass
//     finds all current conflicts and attempts to fix them before checking again.
//
// Returns a slice of Result containing descriptions of all fixes applied and an error.
// The error will be non-nil if any conflict resolution attempt fails.
//
// The method works by:
//  1. Finding all current conflicts in the IgnoreFile
//  2. Attempting to fix each conflict found
//  3. Repeating until no conflicts remain or maxPasses is reached
//  4. Recording each fix operation in the returned Results
//
// Example:
//
//	fixes, err := ignoreFile.FixConflicts(5)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	for _, fix := range fixes {
//	    fmt.Printf("Applied fix: %s\n", fix)
//	}
func (f *IgnoreFile) FixConflicts(maxPasses int) ([]Result, error) {
	fixLogs := make([]Result, 0)

	for range maxPasses {
		conflicts := f.FindConflicts()

		if len(conflicts) == 0 {
			break // All out of conflicts, good job
		}

		for _, conflict := range conflicts {
			description, err := f.fixConflict(conflict)
			if err != nil {
				return fixLogs, err
			}

			fixLogs = append(fixLogs, description)
		}
	}

	return fixLogs, nil
}

func (f *IgnoreFile) addRuleWithConflictResolution(rule Ruler) ([]Result, error) {
	idealInsertionPoint := len(f.rules) // default insertion point to the end

	for i, existing := range f.rules {
		// When adding a rule, check conflicts with each existing rule
		// The intervening rules are everything between existing rule and the end
		intervening := f.rules[i+1:]

		if conflict, found := checkConflict(existing, rule, intervening); found {
			switch conflict.ConflictType {
			case SEMANTIC_CONFLICT:
				return make([]Result, 0), semanticConflictError
			case REDUNDANT_RULE:
				return make([]Result, 0), redundantRuleError
			case UNREACHABLE_RULE:
				return make([]Result, 0), unreachableRuleError
			case INEFFECTIVE_RULE:
				idealInsertionPoint = i
			}
		}

		if f.ruleShouldComeBefore(rule, existing) && i < idealInsertionPoint {
			idealInsertionPoint = i
		}
	}

	f.rules = append(f.rules[:idealInsertionPoint], append([]Ruler{rule}, f.rules[idealInsertionPoint:]...)...)

	addition := Result{
		Rule:   rule,
		Result: ADDED,
		Reason: REQUESTED,
	}

	fixedConflicts, err := f.FixConflicts(20)
	if err != nil {
		return make([]Result, 0), err
	}

	fixedConflicts = append(fixedConflicts, Result{})
	copy(fixedConflicts[1:], fixedConflicts)
	fixedConflicts[0] = addition

	return fixedConflicts, nil
}

func (f *IgnoreFile) deleteMatchingRule(target Ruler, reason ActionReason) (Result, error) {
	for i, rule := range f.rules {
		if rulesEqual(rule, target) {
			f.rules = append(f.rules[:i], f.rules[i+1:]...)
			return Result{
				Rule:   target,
				Result: REMOVED,
				Reason: reason,
			}, nil
		}
	}

	return Result{}, ruleNotFoundError
}

// MARK: Facade methods

// AddFile adds a new file rule to the IgnoreFile with automatic conflict detection and resolution.
// The file path is validated and cleaned before creating the rule, and the method determines
// the optimal insertion point based on rule precedence and conflict analysis.
//
// Parameters:
//   - path: The file system path for the rule. The path will be validated and cleaned.
//   - action: The action to be performed when the rule matches. Must be either INCLUDE or EXCLUDE.
//
// Returns a slice of Result containing the addition operation and any subsequent conflict
// fixes, plus an error. The error will be non-nil if:
//   - The provided path is invalid or cannot be cleaned
//   - The provided action fails validation
//   - A semantic conflict, redundant rule, or unreachable rule is detected
//   - Automatic conflict resolution fails
//
// The method will automatically:
//   - Find the optimal insertion point for the new rule
//   - Detect and handle conflicts with existing rules
//   - Reposition the rule if it would be ineffective at the default location
//   - Run up to 20 passes of conflict resolution after insertion
//
// Example:
//
//	results, err := ignoreFile.AddFile("/path/to/config.json", EXCLUDE)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	for _, result := range results {
//	    fmt.Printf("Operation: %s\n", result)
//	}
func (f *IgnoreFile) AddFile(path string, action Action) ([]Result, error) {
	rule, err := NewFileRule(path, action)
	if err != nil {
		return make([]Result, 0), err
	}

	return f.addRuleWithConflictResolution(rule)
}

// AddExtension adds a new extension rule to the IgnoreFile with automatic conflict detection and resolution.
// The extension is validated and cleaned before creating the rule (whitespace and "*." prefixes are removed),
// and the method determines the optimal insertion point based on rule precedence and conflict analysis.
//
// Parameters:
//   - ext: The file extension for the rule (e.g., "txt", "*.go", ".json").
//     Whitespace and "*." prefixes will be trimmed automatically.
//   - action: The action to be performed when the rule matches. Must be either INCLUDE or EXCLUDE.
//
// Returns a slice of Result containing the addition operation and any subsequent conflict
// fixes, plus an error. The error will be non-nil if:
//   - The provided extension is empty after cleaning (whitespace and prefixes removed)
//   - The provided action fails validation
//   - A semantic conflict, redundant rule, or unreachable rule is detected
//   - Automatic conflict resolution fails
//
// The method will automatically:
//   - Find the optimal insertion point for the new rule
//   - Detect and handle conflicts with existing rules
//   - Reposition the rule if it would be ineffective at the default location
//   - Run up to 20 passes of conflict resolution after insertion
//
// Example:
//
//	results, err := ignoreFile.AddExtension("*.log", EXCLUDE)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// These are all equivalent:
//	ignoreFile.AddExtension("go", INCLUDE)
//	ignoreFile.AddExtension(".go", INCLUDE)
//	ignoreFile.AddExtension("*.go", INCLUDE)
func (f *IgnoreFile) AddExtension(ext string, action Action) ([]Result, error) {
	rule, err := NewExtensionRule(ext, action)
	if err != nil {
		return make([]Result, 0), err
	}

	return f.addRuleWithConflictResolution(rule)
}

// AddDirectory adds a new directory rule to the IgnoreFile with automatic conflict detection and resolution.
// The directory path is validated and cleaned before creating the rule (whitespace, leading, and trailing slashes are removed),
// and the method determines the optimal insertion point based on rule precedence and conflict analysis.
//
// Parameters:
//   - name: The directory path for the rule (e.g., "node_modules", "build/", "/src").
//     Whitespace, leading, and trailing slashes will be trimmed automatically.
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
//   - The provided directory path is empty after cleaning (whitespace and slashes removed)
//   - The provided mode fails validation
//   - The provided action fails validation
//   - A semantic conflict, redundant rule, or unreachable rule is detected
//   - Automatic conflict resolution fails
//
// The method will automatically:
//   - Find the optimal insertion point for the new rule
//   - Detect and handle conflicts with existing rules
//   - Reposition the rule if it would be ineffective at the default location
//   - Run up to 20 passes of conflict resolution after insertion
//
// Example:
//
//	results, err := ignoreFile.AddDirectory("node_modules", RECURSIVE, EXCLUDE)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Match only the build directory at root level
//	results, err = ignoreFile.AddDirectory("build", ROOT_ONLY, EXCLUDE)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (f *IgnoreFile) AddDirectory(name string, mode DirectoryMode, action Action) ([]Result, error) {
	rule, err := NewDirectoryRule(name, mode, action)
	if err != nil {
		return make([]Result, 0), err
	}

	return f.addRuleWithConflictResolution(rule)
}

// AddGlob adds a new glob rule to the IgnoreFile with automatic conflict detection and resolution.
// The glob pattern is validated and cleaned before creating the rule (whitespace is removed),
// and the method determines the optimal insertion point based on rule precedence and conflict analysis.
//
// Parameters:
//   - pattern: The glob pattern for the rule (e.g., "*.tmp", "test/**/*.go", "**/node_modules/**").
//     Whitespace will be trimmed automatically.
//   - action: The action to be performed when the rule matches. Must be either INCLUDE or EXCLUDE.
//
// Returns a slice of Result containing the addition operation and any subsequent conflict
// fixes, plus an error. The error will be non-nil if:
//   - The provided pattern is empty after cleaning (whitespace removed)
//   - The provided action fails validation
//   - A semantic conflict, redundant rule, or unreachable rule is detected
//   - Automatic conflict resolution fails
//
// The method will automatically:
//   - Find the optimal insertion point for the new rule
//   - Detect and handle conflicts with existing rules
//   - Reposition the rule if it would be ineffective at the default location
//   - Run up to 20 passes of conflict resolution after insertion
//
// Example:
//
//	results, err := ignoreFile.AddGlob("*.log", EXCLUDE)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Match all Go test files recursively
//	results, err = ignoreFile.AddGlob("**/*_test.go", INCLUDE)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (f *IgnoreFile) AddGlob(pattern string, action Action) ([]Result, error) {
	rule, err := NewGlobRule(pattern, action)
	if err != nil {
		return make([]Result, 0), err
	}

	return f.addRuleWithConflictResolution(rule)
}

// DeleteFile removes a file rule from the IgnoreFile that exactly matches the specified
// path and action. The path is validated and cleaned using the same process as AddFile
// to ensure accurate matching.
//
// Parameters:
//   - path: The file system path for the rule to delete. The path will be validated and cleaned.
//   - action: The action of the rule to delete. Must be either INCLUDE or EXCLUDE.
//
// Returns a Result containing details of the deletion operation and an error.
// The error will be non-nil if:
//   - The provided path is invalid or cannot be cleaned
//   - The provided action fails validation
//   - No matching rule is found in the IgnoreFile
//
// The method searches for the first rule that exactly matches both the cleaned path
// and the specified action, then removes it from the IgnoreFile.
//
// Example:
//
//	result, err := ignoreFile.DeleteFile("/path/to/config.json", EXCLUDE)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Deleted rule: %s\n", result)
func (f *IgnoreFile) DeleteFile(path string, action Action) (Result, error) {
	targetRule, err := NewFileRule(path, action)
	if err != nil {
		return Result{}, err
	}
	return f.deleteMatchingRule(targetRule, REQUESTED)
}

// DeleteExtension removes an extension rule from the IgnoreFile that exactly matches the specified
// extension and action. The extension is validated and cleaned using the same process as AddExtension
// (whitespace and "*." prefixes are removed) to ensure accurate matching.
//
// Parameters:
//   - ext: The file extension for the rule to delete (e.g., "txt", "*.go", ".json").
//     Whitespace and "*." prefixes will be trimmed automatically for matching.
//   - action: The action of the rule to delete. Must be either INCLUDE or EXCLUDE.
//
// Returns a Result containing details of the deletion operation and an error.
// The error will be non-nil if:
//   - The provided extension is empty after cleaning (whitespace and prefixes removed)
//   - The provided action fails validation
//   - No matching rule is found in the IgnoreFile
//
// The method searches for the first rule that exactly matches both the cleaned extension
// and the specified action, then removes it from the IgnoreFile.
//
// Example:
//
//	result, err := ignoreFile.DeleteExtension("*.log", EXCLUDE)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// These are all equivalent for matching:
//	ignoreFile.DeleteExtension("go", INCLUDE)
//	ignoreFile.DeleteExtension(".go", INCLUDE)
//	ignoreFile.DeleteExtension("*.go", INCLUDE)
func (f *IgnoreFile) DeleteExtension(ext string, action Action) (Result, error) {
	targetRule, err := NewExtensionRule(ext, action)
	if err != nil {
		return Result{}, err
	}
	return f.deleteMatchingRule(targetRule, REQUESTED)
}

// DeleteDirectory removes a directory rule from the IgnoreFile that exactly matches the specified
// directory name, mode, and action. The directory path is validated and cleaned using the same process
// as AddDirectory (whitespace and slashes are removed) to ensure accurate matching.
//
// Parameters:
//   - name: The directory path for the rule to delete (e.g., "node_modules", "build/", "/src").
//     Whitespace and slashes will be trimmed automatically for matching.
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
//   - The provided directory path is empty after cleaning (whitespace and slashes removed)
//   - The provided mode fails validation
//   - The provided action fails validation
//   - No matching rule is found in the IgnoreFile
//
// The method searches for the first rule that exactly matches the cleaned directory name,
// mode, and action, then removes it from the IgnoreFile.
//
// Example:
//
//	result, err := ignoreFile.DeleteDirectory("node_modules", RECURSIVE, EXCLUDE)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// These are equivalent for matching (trailing slash is stripped):
//	ignoreFile.DeleteDirectory("build", ROOT_ONLY, EXCLUDE)
//	ignoreFile.DeleteDirectory("build/", ROOT_ONLY, EXCLUDE)
func (f *IgnoreFile) DeleteDirectory(name string, mode DirectoryMode, action Action) (Result, error) {
	targetRule, err := NewDirectoryRule(name, mode, action)
	if err != nil {
		return Result{}, err
	}
	return f.deleteMatchingRule(targetRule, REQUESTED)
}

// DeleteGlob removes a glob rule from the IgnoreFile that exactly matches the specified
// pattern and action. The glob pattern is validated and cleaned using the same process as AddGlob
// (whitespace is removed) to ensure accurate matching.
//
// Parameters:
//   - pattern: The glob pattern for the rule to delete (e.g., "*.tmp", "test/**/*.go", "**/node_modules/**").
//     Whitespace will be trimmed automatically for matching.
//   - action: The action of the rule to delete. Must be either INCLUDE or EXCLUDE.
//
// Returns a Result containing details of the deletion operation and an error.
// The error will be non-nil if:
//   - The provided pattern is empty after cleaning (whitespace removed)
//   - The provided action fails validation
//   - No matching rule is found in the IgnoreFile
//
// The method searches for the first rule that exactly matches both the cleaned pattern
// and the specified action, then removes it from the IgnoreFile.
//
// Example:
//
//	result, err := ignoreFile.DeleteGlob("*.log", EXCLUDE)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Delete a specific test file pattern
//	result, err = ignoreFile.DeleteGlob("**/*_test.go", INCLUDE)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (f *IgnoreFile) DeleteGlob(pattern string, action Action) (Result, error) {
	targetRule, err := NewGlobRule(pattern, action)
	if err != nil {
		return Result{}, err
	}
	return f.deleteMatchingRule(targetRule, REQUESTED)
}

// MARK: Read methods
func (f IgnoreFile) Rules() []Ruler {
	return f.rules
}

// FindConflicts analyzes all rules in the IgnoreFile and returns a slice of detected conflicts.
// The method performs a comprehensive pairwise comparison of all rules, checking for various
// types of conflicts including semantic conflicts, redundant rules, unreachable rules, and
// ineffective rules.
//
// The conflict detection considers the order of rules and any intervening rules between each
// pair, as rule precedence affects conflict resolution. Each unique conflict is reported only
// once by avoiding duplicate comparisons and self-comparisons.
//
// Returns a slice of Conflict containing all detected conflicts. An empty slice indicates
// no conflicts were found. The conflicts are detected in the order rules appear in the file,
// which may be important for understanding rule precedence issues.
//
// Conflict types that may be detected:
//   - SEMANTIC_CONFLICT: Rules that contradict each other in meaning
//   - REDUNDANT_RULE: Rules that duplicate existing functionality
//   - UNREACHABLE_RULE: Rules that can never be triggered due to earlier rules
//   - INEFFECTIVE_RULE: Rules that would be more effective in a different position
//
// Example:
//
//	conflicts := ignoreFile.FindConflicts()
//	if len(conflicts) > 0 {
//	    fmt.Printf("Found %d conflicts:\n", len(conflicts))
//	    for _, conflict := range conflicts {
//	        fmt.Printf("- %s\n", conflict)
//	    }
//	} else {
//	    fmt.Println("No conflicts detected")
//	}
func (f IgnoreFile) FindConflicts() []Conflict {
	var conflicts []Conflict

	for i, rule1 := range f.rules {
		for j, rule2 := range f.rules {
			if i >= j { // avoid duplicates and self-comparison
				continue
			}

			if conflict, found := checkConflict(rule1, rule2, f.rules[i+1:j]); found {
				conflicts = append(conflicts, conflict)
			}
		}
	}

	return conflicts
}

// MARK: Movement
type MoveDirection int

const (
	BEFORE MoveDirection = iota + 1
	AFTER
)

func MoveDirectionFromString(direction string) (MoveDirection, error) {
	switch direction {
	case "before":
		return BEFORE, nil
	case "after":
		return AFTER, nil
	default:
		return MoveDirection(0), invalidActionError
	}
}

func (f *IgnoreFile) moveRule(from, to int) error {
	if from < 0 || from >= len(f.rules) {
		return sourceIdxOutOfRangeError
	}

	if from == to {
		return nil // No move needed
	}

	// Remove rule from current position
	rule := f.rules[from]
	f.rules = append(f.rules[:from], f.rules[from+1:]...)

	// Adjust target index if needed (now that we removed one element)
	if to > from {
		to--
	}

	// Validate adjusted target
	if to < 0 || to > len(f.rules) {
		return targetIdxOutofRangeError
	}

	// Insert at new position
	f.rules = append(f.rules[:to], append([]Ruler{rule}, f.rules[to:]...)...)
	return nil
}

// MoveRule relocates an existing rule to a new position relative to another rule in the IgnoreFile.
// The method finds both rules by exact match and moves the first rule to either before or after
// the target rule, depending on the specified direction.
//
// Parameters:
//   - ruleToMove: The rule to be relocated. Must exactly match an existing rule in the IgnoreFile.
//   - targetRule: The reference rule that defines the destination position. Must exactly match an existing rule.
//   - direction: The position relative to the target rule. Must be either BEFORE or AFTER.
//   - reason: The reason for the move operation. Must be REQUESTED, AUTOMATED_FIX, or FIX_UNKNOWN.
//
// Returns a Result containing details of the move operation and an error. The error will be non-nil if:
//   - The rule to move is not found in the IgnoreFile
//   - The target rule is not found in the IgnoreFile
//   - The direction parameter is invalid
//   - The calculated new position is out of range
//
// The method includes optimizations to avoid unnecessary moves:
//   - If moving BEFORE and the rule is already immediately before the target, no move occurs
//   - If moving AFTER and the rule is already immediately after the target, no move occurs
//   - If the calculated new position is the same as the current position, no move occurs
//
// Example:
//
//	// Move a file rule to appear before a directory rule
//	result, err := ignoreFile.MoveRule(fileRule, dirRule, BEFORE, REQUESTED)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if result.Result == MOVED {
//	    fmt.Printf("Successfully moved rule: %s\n", result)
//	}
func (f *IgnoreFile) MoveRule(ruleToMove, targetRule Ruler, direction MoveDirection, reason ActionReason) (Result, error) {
	moveIdx := f.findRuleIndex(ruleToMove)
	targetIdx := f.findRuleIndex(targetRule)

	if moveIdx == -1 {
		return Result{}, ruleToMoveNotFoundError
	}

	if targetIdx == -1 {
		return Result{}, targetRuleNotFoundError
	}

	var newIdx int
	switch direction {
	case BEFORE:
		newIdx = targetIdx

		// Only skip if rule is already immediately before target
		if moveIdx == targetIdx-1 {
			return Result{}, nil
		}
	case AFTER:
		newIdx = targetIdx + 1
		// Only skip if rule is already immediately after target
		if moveIdx == newIdx {
			return Result{}, nil
		}
	default:
		return Result{}, invalidateDirectionError
	}

	// No move needed if source and destination are the same
	if moveIdx == newIdx {
		return Result{}, nil
	}

	err := f.moveRule(moveIdx, newIdx)
	if err != nil {
		return Result{}, err
	}

	return Result{
		Rule:   ruleToMove,
		Result: MOVED,
		Reason: reason,
	}, nil
}
