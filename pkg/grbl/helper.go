package grbl

import (
	"context"
	"strings"
)

const WakeUp = "\r\n\r\n"

func IsRealtimeCommand(cmd []byte) bool {
	// ctrl + x
	if byte(cmd[0]) == 0x18 {
		return true
	}
	stringCmd := strings.TrimSpace(string(cmd))
	if len(stringCmd) == 1 {
		switch rune(stringCmd[0]) {
		case '~', '?', '!':
			return true
		}
	}
	return false
}

func IsFeedHold(cmd []byte) bool {
	if len(cmd) == 0 {
		return false
	}
	if rune(string(cmd)[0]) == '!' {
		return true
	}
	return false
}

func IsStartResume(cmd []byte) bool {
	if len(cmd) == 0 {
		return false
	}
	if rune(string(cmd)[0]) == '~' {
		return true
	}
	return false
}

func IsEmptyCommand(cmd []byte) bool {
	if strings.TrimSpace(string(cmd)) == "" {
		return true
	}
	return false
}

func IsPushMsg(msg []byte) bool {
	msgStr := strings.TrimSpace(string(msg))
	if strings.HasPrefix(msgStr, "<") && strings.HasSuffix(msgStr, ">") {
		return true
	}
	if strings.HasPrefix(msgStr, "[MSG:") && strings.HasSuffix(msgStr, "]") {
		return true
	}
	if strings.HasPrefix(msgStr, "[echo:") && strings.HasSuffix(msgStr, "]") {
		return true
	}
	if strings.HasPrefix(msgStr, "Grbl ") {
		return true
	}
	if strings.HasPrefix(msgStr, "ALARM:") {
		return true
	}
	return false
}

func ExpectConfirmation(cmd []byte) bool {
	if IsEmptyCommand(cmd) || IsFeedHold(cmd) {
		return false
	}
	return true
}

func ScanCtx(ctx context.Context, scanner interface{ Scan() bool }) (bool, error) {
	res := make(chan bool)
	go func(res chan bool) {
		res <- scanner.Scan()
	}(res)
	select {
	case scanRes := <-res:
		return scanRes, nil
	case <-ctx.Done():
		return false, ctx.Err()
	}
}
