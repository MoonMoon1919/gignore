package gignore

import (
	"errors"
	"log"
	"strings"
)

var invalidDirectoryError = errors.New("invalid directory error")

func isExtensionPattern(line string) bool {
	return strings.HasPrefix(line, "*.") &&
		!strings.Contains(line[2:], "/") && // path separator is a glob pattern
		!strings.Contains(line[2:], "*") // multiple wildcards is a glob pattern
}

func isDirectoryPattern(line string) bool {
	return strings.HasSuffix(line, "/") || // build/
		strings.HasSuffix(line, "/*") || // build/*
		strings.HasSuffix(line, "/**") || // build/**
		strings.HasPrefix(line, "**/") || // **/build
		strings.HasPrefix(line, "/") // /build
}

func parseDirectoryRule(line string, action Action) (DirectoryRule, error) {
	var name string
	var mode DirectoryMode

	switch {
	case strings.HasSuffix(line, "/**"):
		mode = RECURSIVE
		name = strings.TrimSuffix(line, "/**")
	case strings.HasSuffix(line, "/*"):
		mode = CHILDREN
		name = strings.TrimSuffix(line, "/*")
	case strings.HasSuffix(line, "/"):
		mode = DIRECTORY
		name = strings.TrimSuffix(line, "/")
	case strings.HasPrefix(line, "**/"):
		mode = ANYWHERE
		name = strings.TrimPrefix(line, "**/")
	case strings.HasPrefix(line, "/"):
		mode = ROOT_ONLY
		name = strings.TrimPrefix(line, "/")
	default:
		return DirectoryRule{}, invalidDirectoryError
	}

	return NewDirectoryRule(name, mode, action)
}

func isGlobPattern(line string) bool {
	return strings.Contains(line, "*") ||
		strings.Contains(line, "?") ||
		strings.Contains(line, "[")
}

func parseRule(line string) (Ruler, error) {
	action := INCLUDE
	if strings.HasPrefix(line, "!") {
		action = EXCLUDE
		line = line[1:]
	}

	if isExtensionPattern(line) {
		return NewExtensionRule(line, action)
	}

	if isDirectoryPattern(line) {
		return parseDirectoryRule(line, action)
	}

	if isGlobPattern(line) {
		return NewGlobRule(line, action)
	}

	return NewFileRule(line, action)
}

// Parse converts ignore file content from a string into rules and populates the provided IgnoreFile.
// The function automatically detects rule types (file, extension, directory, or glob patterns) and
// creates the appropriate rule instances. Invalid lines are logged and skipped rather than causing
// the entire parsing operation to fail.
//
// Parameters:
//   - content: The string content of an ignore file to parse.
//   - ignoreFile: A pointer to the IgnoreFile instance to populate with the parsed rules.
//
// The parsing logic follows these rules:
//   - Empty lines and lines starting with "#" (comments) are ignored
//   - Lines starting with "!" are treated as EXCLUDE actions, otherwise INCLUDE
//   - Lines matching "*.ext" (no path separators or additional wildcards) become extension rules
//   - Lines ending with "/", "/*", "/**" or starting with "**/" or "/" become directory rules
//   - Lines containing "*", "?", or "[" become glob rules
//   - All other lines become file rules
//
// Directory pattern detection:
//   - "dirname/" → DIRECTORY mode
//   - "dirname/*" → CHILDREN mode
//   - "dirname/**" → RECURSIVE mode
//   - "**/dirname" → ANYWHERE mode
//   - "/dirname" → ROOT_ONLY mode
//
// Returns nil on success. Parse errors for individual lines are logged but do not stop
// processing - invalid lines are skipped and parsing continues.
//
// Example:
//
//	content := `# My ignore file
//	*.log
//	!important.log
//	node_modules/
//	build/**
//	/temp`
//
//	var ignoreFile IgnoreFile
//	err := Parse(content, &ignoreFile)
//	if err != nil {
//	    log.Fatal(err)
//	}
func Parse(content string, ignoreFile *IgnoreFile) error {
	lines := strings.Split(content, "\n")

	for linNum, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		rule, err := parseRule(line)
		if err != nil {
			// Log and ignore errors
			log.Printf("error loading line %d, preserving %s as is. error: %v", linNum+1, line, err)
			continue
		}

		ignoreFile.addRule(rule)
	}

	return nil
}
