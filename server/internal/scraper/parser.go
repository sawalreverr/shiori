package scraper

import (
	"regexp"
	"strings"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

// ExtractByRegex finds all matches for a pattern
func (p *Parser) ExtractByRegex(html, pattern string) []string {
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(html, -1)

	results := make([]string, 0)
	for _, match := range matches {
		if len(match) > 1 {
			results = append(results, match[1])
		}
	}

	return results
}

// DecodeHTMLEntities converts &amp; to & etc
func (p *Parser) DecodeHTMLEntities(text string) string {
	replacements := map[string]string{
		"&nbsp;": " ",
		"&amp;":  "&",
		"&lt;":   "<",
		"&gt;":   ">",
		"&quot;": "\"",
		"&#39;":  "'",
	}

	for entity, char := range replacements {
		text = strings.ReplaceAll(text, entity, char)
	}

	return text
}
