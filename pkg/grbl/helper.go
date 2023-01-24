package grbl

import "strings"

func IsRealtimeCommand(cmd []byte) bool {
	stringCmd := strings.TrimSpace(string(cmd))
	if len(stringCmd) == 1 {
		// ctrl + x
		if byte(cmd[0]) == 0x18 || IsFeedHold(cmd) {
			return true
		}
		switch rune(cmd[0]) {
		case '~', '?':
			return true
		}
	}
	return false
}

// IsFeedHold for some reason this doesn't return anything which means there is no way to scan
// a result.
func IsFeedHold(cmd []byte) bool {
	if len(cmd) == 0 {
		return false
	}
	if rune(string(cmd)[0]) == '!' {
		return true
	}
	return false
}
