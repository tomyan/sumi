package script

import "strings"

// findStateAssignments scans function body lines for assignments to known state variables.
// Looks for "stateVar = expr" patterns (plain = not :=).
func findStateAssignments(body string, stateNames map[string]bool) []StateAssignment {
	if len(stateNames) == 0 {
		return nil
	}

	var assignments []StateAssignment
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		assignments = appendLineAssignments(assignments, trimmed, stateNames)
	}
	return assignments
}

// appendLineAssignments checks a single trimmed line for state assignments
// and appends any matches to the given slice.
func appendLineAssignments(assignments []StateAssignment, trimmed string, stateNames map[string]bool) []StateAssignment {
	for name := range stateNames {
		if isStateAssignment(trimmed, name) {
			assignments = append(assignments, StateAssignment{VarName: name, Line: trimmed})
		}
	}
	return assignments
}

// isStateAssignment checks whether a trimmed line is an assignment to the given state variable.
func isStateAssignment(trimmed, name string) bool {
	if !strings.HasPrefix(trimmed, name) {
		return false
	}
	afterName := trimmed[len(name):]
	if len(afterName) > 0 && isIdentChar(afterName[0]) {
		return false
	}
	return isPlainAssignment(strings.TrimLeft(afterName, " \t"))
}

// isPlainAssignment checks that a string starts with "=" but not ":=" or "==".
func isPlainAssignment(rest string) bool {
	if len(rest) == 0 || rest[0] != '=' {
		return false
	}
	return len(rest) < 2 || (rest[1] != '=' && rest[1] != ':')
}
