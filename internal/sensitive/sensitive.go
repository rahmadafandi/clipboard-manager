package sensitive

import "regexp"

// Patterns that indicate sensitive data.
var patterns = []*regexp.Regexp{
	// API keys / tokens
	regexp.MustCompile(`(?i)(sk-[a-zA-Z0-9]{20,})`),                       // OpenAI
	regexp.MustCompile(`(?i)(ghp_[a-zA-Z0-9]{36,})`),                      // GitHub PAT
	regexp.MustCompile(`(?i)(gho_[a-zA-Z0-9]{36,})`),                      // GitHub OAuth
	regexp.MustCompile(`(?i)(glpat-[a-zA-Z0-9\-]{20,})`),                  // GitLab PAT
	regexp.MustCompile(`AKIA[0-9A-Z]{16}`),                                // AWS access key
	regexp.MustCompile(`(?i)(xox[bpsa]-[a-zA-Z0-9\-]{10,})`),             // Slack token
	regexp.MustCompile(`(?i)(ya29\.[a-zA-Z0-9_\-]{50,})`),                // Google OAuth
	regexp.MustCompile(`(?i)bearer\s+[a-zA-Z0-9\-_.]{20,}`),              // Bearer token
	regexp.MustCompile(`(?i)(api[_-]?key|apikey|secret[_-]?key)\s*[:=]\s*\S{8,}`), // Generic key=value

	// Passwords
	regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[:=]\s*\S{4,}`),

	// Private keys
	regexp.MustCompile(`-----BEGIN (RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----`),

	// Connection strings
	regexp.MustCompile(`(?i)(mongodb|postgres|mysql|redis)://\S+:\S+@`),
}

// IsSensitive returns true if text matches any known sensitive data pattern.
func IsSensitive(text string) bool {
	for _, p := range patterns {
		if p.MatchString(text) {
			return true
		}
	}
	return false
}
