package scraper

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

// StripTags removes HTML tags from text
func (p *Parser) StripTags(html string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	text := re.ReplaceAllString(html, "")

	text = strings.TrimSpace(text)
	re = regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")

	return text
}

// DecodeHTMLEntities converts &amp; to & etc
func (p *Parser) DecodeHTMLEntities(text string) string {
	replacements := map[string]string{
		"&nbsp;": "",
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

// ExtractBetween returns string parsing
func (p *Parser) ExtractBetween(s, start, end string) string {
	i := strings.Index(s, start)
	if i == -1 {
		return ""
	}
	i += len(start)

	j := strings.Index(s[i:], end)
	if j == -1 {
		return ""
	}
	return s[i : i+j]
}

// ParseTime returns time formatting in bahasa
func (p *Parser) ParseTime(s string) time.Time {
	now := time.Now()
	s = p.DecodeHTMLEntities(s)
	s = strings.ToLower(s)

	parts := strings.Split(s, " ")
	if len(parts) < 2 {
		return time.Time{}
	}

	value, err := strconv.Atoi(parts[0])
	if err != nil {
		return time.Time{}
	}

	unit := parts[1]
	switch {
	case strings.Contains(unit, "menit"):
		return now.Add(-time.Duration(value) * time.Minute)
	case strings.Contains(unit, "jam"):
		return now.Add(-time.Duration(value) * time.Hour)
	case strings.Contains(unit, "hari"):
		return now.AddDate(0, 0, -value)
	case strings.Contains(unit, "minggu"):
		return now.AddDate(0, 0, -7*value)
	default:
		return time.Time{}
	}
}
