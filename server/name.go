package server

import (
	"strings"
	"unicode"
)

func ToFriendlyName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "api", "API")
	names := strings.Split(name, "-")
	for i := 1; i < len(names); i++ {
		n := []rune(names[i])
		n[0] = unicode.ToUpper(n[0])
		names[i] = string(n)
	}
	name = strings.Join(names, "")
	return name
}
