package templating

import (
	"math/rand/v2"
	"strings"
)

// randomLink generates a plausible, random, relative URL path.
func (tm *TemplateManager) randomLink() string {
	var builder strings.Builder
	builder.WriteByte('/')

	// Generate the first path segment, ensuring it's not in the whitelist.
	for {
		segment := randomWord()
		path := "/" + segment
		if _, isWhitelisted := tm.whitelistMap[path]; !isWhitelisted {
			builder.WriteString(segment)
			break
		}
	}

	// Add additional random subpaths.
	numSubpaths := rand.IntN(tm.config.MaxSubpaths-tm.config.MinSubpaths+1) + tm.config.MinSubpaths
	for i := 1; i < numSubpaths; i++ {
		builder.WriteByte('/')
		builder.WriteString(randomWord())
	}

	// End with a trailing slash for a directory-like appearance.
	builder.WriteByte('/')
	return builder.String()
}

// randomQueryLink generates a random URL path and appends a specified number of random query parameters.
func (tm *TemplateManager) randomQueryLink(keyCount int) string {
	path := tm.randomLink()
	if keyCount <= 0 {
		return path
	}

	var builder strings.Builder
	builder.WriteString(path)
	builder.WriteByte('?')

	for i := 0; i < keyCount; i++ {
		key := randomWord()
		// Using the pure `randomString` function for the value.
		value := randomString("alphanum", 12)
		builder.WriteString(key)
		builder.WriteByte('=')
		builder.WriteString(value)
		if i < keyCount-1 {
			builder.WriteByte('&')
		}
	}
	return builder.String()
}
