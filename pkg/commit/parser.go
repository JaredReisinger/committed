package commit

import (
	"fmt"
	"regexp"
	"strings"
)

// ParseMessage parses a commit message into a structured Message.
// It attempts to extract as much conventional commit structure as possible,
// even from malformed messages.
func ParseMessage(raw string) (*Message, error) {
	msg := &Message{
		RawMessage: raw,
	}

	lines := strings.Split(strings.TrimSpace(raw), "\n")

	if len(lines) == 0 {
		return msg, nil
	}

	// Parse header (first line)
	header := strings.TrimSpace(lines[0])
	if err := parseHeader(header, msg); err != nil {
		// If header parsing fails, treat entire message as description
		msg.Description = raw
		return msg, nil
	}

	// Parse body and footers
	if len(lines) > 1 {
		bodyLines := []string{}
		footerLines := []string{}

		// Find the first footer-like line
		footerStart := -1
		for i := 1; i < len(lines); i++ {
			line := strings.TrimSpace(lines[i])

			// Check if this looks like a footer (token: value)
			if strings.Contains(line, ": ") {
				parts := strings.SplitN(line, ": ", 2)
				if len(parts) == 2 && isFooterToken(parts[0]) {
					footerStart = i
					break
				}
			}
		}

		// If we found footers, everything before is body
		if footerStart != -1 {
			for i := 1; i < footerStart; i++ {
				bodyLines = append(bodyLines, lines[i])
			}
			for i := footerStart; i < len(lines); i++ {
				footerLines = append(footerLines, lines[i])
			}
		} else {
			// No footers, everything after header is body
			bodyLines = lines[1:]
		}

		// Set body (trim empty lines from start/end)
		if len(bodyLines) > 0 {
			// Remove leading/trailing empty lines
			start := 0
			for start < len(bodyLines) && strings.TrimSpace(bodyLines[start]) == "" {
				start++
			}
			end := len(bodyLines)
			for end > start && strings.TrimSpace(bodyLines[end-1]) == "" {
				end--
			}
			if start < end {
				msg.Body = strings.Join(bodyLines[start:end], "\n")
			}
		}

		// Parse footers
		for _, line := range footerLines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if strings.Contains(line, ": ") {
				parts := strings.SplitN(line, ": ", 2)
				if len(parts) == 2 {
					msg.Footers = append(msg.Footers, Footer{
						Token: parts[0],
						Value: parts[1],
					})
				}
			}
		}
	}

	return msg, nil
}

// parseHeader parses the header line: type(scope): description
func parseHeader(header string, msg *Message) error {
	// Match: type(scope): description or type: description
	re := regexp.MustCompile(`^(\w+)(?:\(([^)]+)\))?:\s*(.+)$`)
	matches := re.FindStringSubmatch(header)

	if len(matches) != 4 {
		return fmt.Errorf("invalid header format")
	}

	msg.Type = matches[1]
	if matches[2] != "" {
		msg.Scope = matches[2]
	}
	msg.Description = matches[3]

	return nil
}

// isFooterToken checks if a string looks like a conventional footer token.
func isFooterToken(token string) bool {
	// Common footer tokens
	commonTokens := []string{
		"Closes", "Fixes", "Resolves", "BREAKING CHANGE", "BREAKING-CHANGE",
		"Signed-off-by", "Co-authored-by", "Reviewed-by", "Tested-by",
	}

	token = strings.TrimSpace(token)
	for _, ct := range commonTokens {
		if strings.EqualFold(token, ct) {
			return true
		}
	}

	// Also accept tokens that look like issue references
	if strings.Contains(token, "#") || strings.Contains(token, "-") {
		return true
	}

	return false
}
