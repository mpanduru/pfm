package app

import (
	"fmt"
	"regexp"
	"strings"

	"example.com/pfm/internal/db"
)

type compiledRule struct {
	ID       int64
	Name     string
	Category string
	Re       *regexp.Regexp
	Priority int64
}

func compileRules(rows []db.RuleRow) ([]compiledRule, error) {
	out := make([]compiledRule, 0, len(rows))
	for _, r := range rows {
		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			return nil, fmt.Errorf("rule %d (%s) bad regex: %w", r.ID, r.Name, err)
		}
		out = append(out, compiledRule{
			ID: r.ID, Name: r.Name, Category: r.Category, Re: re, Priority: r.Priority,
		})
	}
	return out, nil
}

func matchCategory(rules []compiledRule, payee, memo string) (string, *compiledRule) {
	text := strings.TrimSpace(payee + " " + memo)
	for _, r := range rules {
		if r.Re.MatchString(text) {
			return r.Category, &r
		}
	}
	return "", nil
}
