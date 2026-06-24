package telegram

import (
	"strconv"
	"strings"
)

func extractIDnumber(text string) int {
	parts := strings.Split(text, "/")
	if len(parts) == 0 {
		return 0
	}
	numStr := parts[len(parts)-1]
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0
	}
	return num
}
