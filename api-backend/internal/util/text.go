package util

import (
	"strconv"
	"strings"
)

func ToInt(numberString string) int {
	num, err := strconv.Atoi(numberString)
	if err != nil {
		return 0
	}

	return num
}

func NextToLastField(value string) string {
	fields := strings.Fields(value)
	if len(fields) < 2 {
		return ""
	}

	return fields[len(fields)-2]
}

func NormalizeName(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func ShortenID(id string) string {
	parts := strings.Split(strings.Trim(id, "/"), "-")
	return parts[len(parts)-1]
}

func TrimLabel(value string) string {
	_, after, found := strings.Cut(value, ":")
	if found {
		return strings.TrimSpace(after)
	}

	return strings.TrimSpace(value)
}
