package crawler

import (
	"context"
	"regexp"
	"strings"

	"github.com/myzie/web/fetch"
)

// MatchType defines the type of pattern matching for parser rules
type MatchType string

const (
	MatchExact  MatchType = "exact"  // Exact domain match
	MatchRegex  MatchType = "regex"  // Regular expression match
	MatchSuffix MatchType = "suffix" // Domain suffix match (e.g., ".com")
	MatchPrefix MatchType = "prefix" // Domain prefix match (e.g., "blog.")
	MatchGlob   MatchType = "glob"   // Glob pattern match (e.g., "*.example.com")
)

// ParserRule defines a flexible rule for matching domains to parsers
type ParserRule struct {
	Pattern  string         // The pattern to match against
	Type     MatchType      // The type of matching to perform
	Parser   Parser         // The parser to use for matching domains
	Priority int            // Priority for rule evaluation (higher = first)
	compiled *regexp.Regexp // Compiled regex for performance (internal use)
}

// Parser is an interface describing a webpage parser. It accepts the fetched
// page and returns a parsed object.
type Parser interface {
	Parse(ctx context.Context, page *fetch.Response) (any, error)
}

// Compile compiles regex patterns for the parser rule if needed
func (r *ParserRule) Compile() error {
	switch r.Type {
	case MatchRegex:
		compiled, err := regexp.Compile(r.Pattern)
		if err != nil {
			return err
		}
		r.compiled = compiled
	case MatchGlob:
		// Convert glob pattern to regex
		regexPattern := globToRegex(r.Pattern)
		compiled, err := regexp.Compile(regexPattern)
		if err != nil {
			return err
		}
		r.compiled = compiled
	}
	return nil
}

// globToRegex converts a glob pattern to a regular expression
func globToRegex(pattern string) string {
	// Escape special regex characters except * and ?
	pattern = regexp.QuoteMeta(pattern)
	// Replace escaped glob characters with regex equivalents
	pattern = strings.ReplaceAll(pattern, "\\*", ".*")
	pattern = strings.ReplaceAll(pattern, "\\?", ".")
	// Anchor the pattern
	return "^" + pattern + "$"
}

// NewExactRule creates a parser rule that matches domains exactly
func NewExactRule(domain string, parser Parser, priority int) ParserRule {
	return ParserRule{
		Pattern:  domain,
		Type:     MatchExact,
		Parser:   parser,
		Priority: priority,
	}
}

// NewRegexRule creates a parser rule that matches domains using a regular expression
func NewRegexRule(pattern string, parser Parser, priority int) ParserRule {
	return ParserRule{
		Pattern:  pattern,
		Type:     MatchRegex,
		Parser:   parser,
		Priority: priority,
	}
}

// NewSuffixRule creates a parser rule that matches domains by suffix (e.g., ".com", ".org")
func NewSuffixRule(suffix string, parser Parser, priority int) ParserRule {
	return ParserRule{
		Pattern:  suffix,
		Type:     MatchSuffix,
		Parser:   parser,
		Priority: priority,
	}
}

// NewPrefixRule creates a parser rule that matches domains by prefix (e.g., "blog.", "api.")
func NewPrefixRule(prefix string, parser Parser, priority int) ParserRule {
	return ParserRule{
		Pattern:  prefix,
		Type:     MatchPrefix,
		Parser:   parser,
		Priority: priority,
	}
}

// NewGlobRule creates a parser rule that matches domains using glob patterns (e.g., "*.example.com")
func NewGlobRule(pattern string, parser Parser, priority int) ParserRule {
	return ParserRule{
		Pattern:  pattern,
		Type:     MatchGlob,
		Parser:   parser,
		Priority: priority,
	}
}
