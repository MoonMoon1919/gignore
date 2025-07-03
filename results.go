package gignore

import "fmt"

// ActionResult represents the outcome of an operation performed on a rule within an IgnoreFile
type ActionResult int

const (
	// REVIEW_RECOMMENDED indicates that manual review is suggested for the operation.
	REVIEW_RECOMMENDED ActionResult = iota + 1
	// FIXED indicates that a conflict or issue was automatically resolved.
	FIXED
	// ADDED indicates that a new rule was successfully added to the IgnoreFile.
	ADDED
	// MOVED indicates that an existing rule was relocated to a different position.
	MOVED
	// REMOVED indicates that an existing rule was deleted from the IgnoreFile.
	REMOVED
)

func (a ActionResult) String() string {
	switch a {
	case REVIEW_RECOMMENDED:
		return "REVIEW_RECOMMENDED"
	case FIXED:
		return "FIXED"
	case MOVED:
		return "MOVED"
	case ADDED:
		return "ADDED"
	case REMOVED:
		return "REMOVED"
	default:
		return ""
	}
}

// ActionReason represents the cause or motivation behind an operation on a rule.
type ActionReason int

const (
	// REQUESTED indicates the operation was explicitly requested by the user.
	REQUESTED ActionReason = iota + 1
	// AUTOMATED_FIX indicates the operation was performed automatically to resolve a conflict.
	AUTOMATED_FIX
	// FIX_UNKNOWN indicates the operation addresses a semantic conflict where the appropriate
	// fix is not apparent enough to be performed automatically and requires manual intervention.
	FIX_UNKNOWN
)

func (a ActionReason) String() string {
	switch a {
	case REQUESTED:
		return "REQUESTED"
	case AUTOMATED_FIX:
		return "AUTOMATED_FIX"
	case FIX_UNKNOWN:
		return "FIX_UNKNOWN"
	default:
		return ""
	}
}

// Result represents the outcome of an operation performed on an IgnoreFile, including
// details about what rule was affected, what happened to it, and why.
type Result struct {
	Rule   Ruler
	Result ActionResult
	Reason ActionReason
}

// Log returns a formatted string representation of the Result suitable for logging
// or display purposes. The format includes the operation type, rule content, and reason.
//
// Example output: "ADDED: Rule '*.log', Reason: REQUESTED"
func (r Result) Log() string {
	return fmt.Sprintf("%s: Rule '%s', Reason: %s", r.Result.String(), r.Rule.Render(), r.Reason.String())
}
