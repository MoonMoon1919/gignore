package gignore

import (
	"strings"
)

type ConflictType string

const (
	SEMANTIC_CONFLICT ConflictType = "SEMANTIC_CONFLICT" // Same pattern, different actions
	REDUNDANT_RULE    ConflictType = "REDUNANT_RULE"     // Same pattern, same action
	UNREACHABLE_RULE  ConflictType = "UNREACHABLE_RULE"  // Broader rule makes specific on meaningless
	INEFFECTIVE_RULE  ConflictType = "INEFFECTIVE_RULE"  // Different action, but rule subsumes another rule
)

type Conflict struct {
	Left         Ruler
	Right        Ruler
	ConflictType ConflictType
}

func checkConflict(left, right Ruler, intervening []Ruler) (Conflict, bool) {
	if left.Pattern() == right.Pattern() {
		if left.Action() != right.Action() {
			return Conflict{
				Left:         left,
				Right:        right,
				ConflictType: SEMANTIC_CONFLICT,
			}, true
		}

		return Conflict{
			Left:         left,
			Right:        right,
			ConflictType: REDUNDANT_RULE,
		}, true
	}

	if left.Action() == right.Action() {
		if subsumes(left, right) {
			if hasInterveningExceptions(left, right, intervening) {
				return Conflict{}, false // Not a conflict due to intervening exceptions
			}

			return Conflict{Left: left, Right: right, ConflictType: UNREACHABLE_RULE}, true
		}

		if subsumes(right, left) {
			if hasInterveningExceptions(right, left, intervening) {
				return Conflict{}, false // Not a conflict due to intervening exceptions
			}

			// Intentionally flip right & left so conflict fixer can handle this state correctly
			return Conflict{Left: right, Right: left, ConflictType: UNREACHABLE_RULE}, true
		}
	}

	if left.Action() == EXCLUDE && right.Action() == INCLUDE {
		if subsumes(right, left) { // exception rule is broader than exclusion
			return Conflict{Left: left, Right: right, ConflictType: INEFFECTIVE_RULE}, true
		}
	}

	return Conflict{}, false
}

func directorySubsumes(rule DirectoryRule, specific Ruler) bool {
	switch other := specific.(type) {
	case DirectoryRule:
		if rule.name == other.name {
			return dirModeSubsumes(rule.mode, other.mode)
		}

		return pathSubsumes(rule.Pattern(), other.Pattern())

	case FileRule:
		return pathStartsWith(other.Pattern(), rule.name+"/")
	case GlobRule:
		if rule.mode == RECURSIVE {
			return pathStartsWith(other.Pattern(), strings.TrimSuffix(rule.Pattern(), "/**"))
		}

		return false
	}

	return false
}

func dirModeSubsumes(broader, narrower DirectoryMode) bool {
	// Check for recursive rules that subsume
	if broader == RECURSIVE && (narrower == DIRECTORY || narrower == CHILDREN) {
		return true
	}

	// Check if there is "child" rule that subsumes a base dir
	if broader == CHILDREN && narrower == DIRECTORY {
		return true
	}

	return false
}

// Checks if broader path contains a more specific path
func pathSubsumes(broader, specific string) bool {
	return strings.HasPrefix(specific, broader)
}

// Checks if a path starts with a specific prefix (e.g., directory containment)
func pathStartsWith(path, prefix string) bool {
	return strings.HasPrefix(path, prefix)
}

// Checks if a glob rule and another rule have matching extensions
func globSubsumes(glob GlobRule, other Ruler) bool {
	if strings.HasPrefix(glob.pattern, "*.") {
		if file, ok := other.(FileRule); ok {
			ext := strings.TrimPrefix(glob.pattern, "*.")
			return strings.HasSuffix(file.Pattern(), "."+ext)
		}
	}

	return false
}

func extensionSubsumes(ext ExtensionRule, other Ruler) bool {
	switch o := other.(type) {
	case FileRule:
		return strings.HasSuffix(o.Pattern(), "."+ext.ext)
	case GlobRule:
		if strings.HasSuffix(o.pattern, "."+ext.ext) {
			return true
		}
		return false
	default:
		// Extensions dont subsume other extensions OR directories
		return false
	}
}

func subsumes(left, right Ruler) bool {
	switch b := left.(type) {
	case DirectoryRule:
		return directorySubsumes(b, right)
	case GlobRule:
		return globSubsumes(b, right)
	case ExtensionRule:
		return extensionSubsumes(b, right)
	case FileRule:
		return false // Files never subsume other files
	}

	return false
}

func hasInterveningExceptions(broader, specific Ruler, intervening []Ruler) bool {
	for _, rule := range intervening {
		// If there's an exception rule (opposite action) that affects the same pattern space
		if rule.Action() != broader.Action() {
			// Check if this exception rule relates to the pattern space between broader and specific
			// For example: !build/important.txt between build/** and build/
			if ruleAffectsPatternSpace(rule, broader, specific) {
				return true
			}
		}
	}
	return false
}

func ruleAffectsPatternSpace(exception, broader, specific Ruler) bool {
	// Get the base patterns
	broaderBase := strings.TrimSuffix(broader.Pattern(), "/**")

	// The exception rule affects the pattern space if:
	// 1. It relates to the broader pattern's domain AND
	// 2. It could potentially affect files that the specific rule would also affect

	// Check if exception is within the broader pattern's domain
	if !strings.HasPrefix(exception.Pattern(), broaderBase) {
		return false
	}

	// Check if the exception could affect the same files as the specific rule
	switch specificRule := specific.(type) {
	case DirectoryRule:
		// If specific is build/ and exception is !build/important.txt
		// The exception affects files within the build/ directory
		if specificRule.mode == DIRECTORY {
			specificBase := specificRule.name
			return strings.HasPrefix(exception.Pattern(), specificBase)
		}
	case FileRule:
		// If specific is a file rule, check if exception affects that exact file or its directory
		return strings.HasPrefix(specificRule.Pattern(), strings.TrimPrefix(exception.Pattern(), "!"))
	}

	// Fallback to simpler check
	return strings.HasPrefix(exception.Pattern(), broaderBase)
}
