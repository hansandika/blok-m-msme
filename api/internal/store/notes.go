package store

import (
	"strconv"
	"strings"
	"unicode"
)

// NotePatch is the structured overlay parsed from suggestion notes.
// See the comment on ApplySuggestion for the correction / new_place field rules.
type NotePatch struct {
	Name         *string
	Category     *string
	Price        *string
	OpenLate     *bool
	Lat          *float64
	Lng          *float64
	Neighborhood *string
	Address      *string
	Hours        *string
	Description  *string
	SourceNote   *string
	Vibes        []string
}

var allowedVibes = map[string]bool{
	"quiet": true, "lively": true, "hidden": true, "hangout": true,
	"date-night": true, "after-office": true, "quick-bite": true, "family": true,
}

var allowedCategories = map[string]bool{
	"japanese": true, "street": true, "cafe": true, "retail": true, "other": true,
}

var noteKeyAliases = map[string]string{
	"name": "name", "category": "category", "price": "price",
	"openlate": "openlate", "larut": "openlate",
	"lat": "lat", "latitude": "lat",
	"lng": "lng", "lon": "lng", "longitude": "lng",
	"neighborhood": "neighborhood", "address": "address",
	"hours": "hours", "hour": "hours",
	"description": "description",
	"source":      "source", "sourcenote": "source",
	"vibe": "vibes", "vibes": "vibes",
}

// ParseNotes extracts key:value lines and a few bare tokens from free-text notes.
func ParseNotes(notes string) NotePatch {
	var p NotePatch
	var leftover []string
	for _, raw := range strings.Split(notes, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		key, val, ok := splitNoteKV(line)
		if !ok {
			leftover = append(leftover, line)
			continue
		}
		switch key {
		case "name":
			p.Name = strPtr(val)
		case "category":
			c := strings.ToLower(val)
			if allowedCategories[c] {
				p.Category = strPtr(c)
			}
		case "price":
			if pr, ok := parsePrice(val); ok {
				p.Price = strPtr(pr)
			}
		case "openlate":
			if b, err := ParseBool(val); err == nil && b != nil {
				p.OpenLate = b
			}
		case "lat":
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				p.Lat = &f
			}
		case "lng":
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				p.Lng = &f
			}
		case "neighborhood":
			p.Neighborhood = strPtr(val)
		case "address":
			p.Address = strPtr(val)
		case "hours":
			p.Hours = strPtr(val)
		case "description":
			p.Description = strPtr(val)
		case "source":
			p.SourceNote = strPtr(val)
		case "vibes":
			p.Vibes = parseVibeList(val)
		}
	}

	applyBareNoteTokens(&p, strings.Join(leftover, "\n"))
	return p
}

func splitNoteKV(line string) (key, val string, ok bool) {
	i := strings.Index(line, ":")
	if i <= 0 {
		return "", "", false
	}
	compact := compactKey(line[:i])
	canon, known := noteKeyAliases[compact]
	if !known {
		return "", "", false
	}
	val = strings.TrimSpace(line[i+1:])
	if val == "" {
		return "", "", false
	}
	return canon, val, true
}

func compactKey(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func applyBareNoteTokens(p *NotePatch, rest string) {
	lower := strings.ToLower(rest)
	if p.OpenLate == nil {
		switch {
		case strings.Contains(lower, "not open late"), strings.Contains(lower, "closes early"), strings.Contains(lower, "tutup awal"):
			f := false
			p.OpenLate = &f
		case strings.Contains(lower, "open late"), strings.Contains(lower, "open-late"),
			containsWord(lower, "larut"):
			t := true
			p.OpenLate = &t
		}
	}
	if p.Price == nil {
		if pr, ok := findBarePrice(rest); ok {
			p.Price = strPtr(pr)
		}
	}
}

func containsWord(s, word string) bool {
	idx := 0
	for {
		i := strings.Index(s[idx:], word)
		if i < 0 {
			return false
		}
		i += idx
		beforeOK := i == 0 || !isWordChar(rune(s[i-1]))
		end := i + len(word)
		afterOK := end >= len(s) || !isWordChar(rune(s[end]))
		if beforeOK && afterOK {
			return true
		}
		idx = i + 1
		if idx >= len(s) {
			return false
		}
	}
}

func isWordChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

func findBarePrice(s string) (string, bool) {
	for _, tok := range strings.Fields(s) {
		if pr, ok := parsePrice(strings.Trim(tok, ".,;")); ok {
			return pr, true
		}
	}
	return "", false
}

func parsePrice(s string) (string, bool) {
	s = strings.TrimSpace(s)
	switch s {
	case "$", "$$", "$$$":
		return s, true
	default:
		return "", false
	}
}

func parseVibeList(s string) []string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == ';' || unicode.IsSpace(r)
	})
	out := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, p := range parts {
		v := strings.ToLower(strings.TrimSpace(p))
		if !allowedVibes[v] || seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	return out
}

func strPtr(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(*s)
}

func mergeVibes(existing, extra []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(existing)+len(extra))
	for _, v := range append(existing, extra...) {
		v = strings.ToLower(strings.TrimSpace(v))
		if v == "" || seen[v] || !allowedVibes[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	return out
}

// Slugify turns a place name into a URL-safe slug.
func Slugify(name string) string {
	var b strings.Builder
	lastHyphen := false
	for _, r := range strings.ToLower(strings.TrimSpace(name)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastHyphen = false
			continue
		}
		if b.Len() > 0 && !lastHyphen {
			b.WriteByte('-')
			lastHyphen = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "place"
	}
	return out
}
