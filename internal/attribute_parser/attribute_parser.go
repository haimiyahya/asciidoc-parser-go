// Package attribute_parser provides attribute parsing utilities for AsciiDoc.
package attribute_parser

import (
	"regexp"
	"strings"
)

// ParseAttributes parses attribute syntax from a string and returns an Attributes map.
// Supports: key="value", role="text", id="name", style="css", etc.
func ParseAttributes(attrs string) map[string]string {
	result := make(map[string]string)

	// Parse key=value pairs (including role, id, style attributes)
	// Matches patterns like: key=value, role="note", id=myid, etc.
	keyValueRegex := regexp.MustCompile(`(\w+)=("[^"]*"|[\w./-]+)`)
	for _, match := range keyValueRegex.FindAllStringSubmatch(attrs, -1) {
		if len(match) >= 3 {
			key := strings.TrimSpace(match[1])
			value := strings.TrimSpace(match[2])
			// Remove quotes if present
			value = strings.Trim(value, `"`)
			result[key] = value
		}
	}

	// Parse document attributes like :toc:
	docAttrRegex := regexp.MustCompile(`:([\w-]+):`)
	for _, match := range docAttrRegex.FindAllStringSubmatch(attrs, -1) {
		if len(match) >= 2 {
			key := strings.TrimSpace(match[1])
			result[key] = ""
		}
	}

	return result
}
