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
	s = strings.TrimSpace(s)
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

// ParseTimeFormat returns time formatting with spesific format
func (p *Parser) ParseTimeFormat(s string, layout string) (time.Time, error) {
	monthMap := map[string]string{
		"januari":   "January",
		"februari":  "February",
		"maret":     "March",
		"april":     "April",
		"mei":       "May",
		"juni":      "June",
		"juli":      "July",
		"agustus":   "August",
		"september": "September",
		"oktober":   "October",
		"november":  "November",
		"desember":  "December",
	}

	formatted := strings.TrimSpace(s)
	formatted = strings.ReplaceAll(formatted, "WIB", "+0700")
	formatted = strings.ReplaceAll(formatted, ",", "")

	formatted = strings.ToLower(formatted)

	for id, en := range monthMap {
		if strings.Contains(formatted, id) {
			formatted = strings.Replace(formatted, id, en, 1)
			break
		}
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")
	return time.ParseInLocation(layout, formatted, loc)
}
