package markov

import (
	"bufio"
	"io"
	"regexp"
)

// DefaultTokenizer is a default implementation of the Tokenizer interface.
// It uses regular expressions to split text into words and punctuation,
// and identifies sentence-ending punctuation as End-Of-Chain (EOC) tokens.
// Its behavior can be customized with functional options.
type DefaultTokenizer struct {
	separator         string
	eoc               string
	separatorRegex    *regexp.Regexp
	eocRegex          *regexp.Regexp
	separatorExcRegex *regexp.Regexp
	eocExcRegex       *regexp.Regexp
}

// Option Is a function that configures a DefaultTokenizer.
type Option func(*DefaultTokenizer)

// WithSeparator Sets the character used for joining tokens during generation.
// Default: " "
func WithSeparator(sep string) Option {
	return func(t *DefaultTokenizer) {
		t.separator = sep
	}
}

// WithEOC Sets the string to use in final output for an EOC token.
// Default: "."
func WithEOC(eoc string) Option {
	return func(t *DefaultTokenizer) {
		t.eoc = eoc
	}
}

// WithSeparatorRegex sets the regex string to use when splitting input text.
// Default: `[\w']+|[.,!?;]`
func WithSeparatorRegex(splitRegex string) Option {
	return func(t *DefaultTokenizer) {
		t.separatorRegex = regexp.MustCompile(splitRegex)
	}
}

// WithEOCRegex sets the regex string to use when deciding whether a token is an EOC token or not.
// Default: `^[.!?]$`
func WithEOCRegex(eocRegex string) Option {
	return func(t *DefaultTokenizer) {
		t.eocRegex = regexp.MustCompile(eocRegex)
	}
}

// WithSeparatorExcRegex sets the regex string to use when deciding whether to add a separator before a token.
func WithSeparatorExcRegex(splitExcRegex string) Option {
	return func(t *DefaultTokenizer) {
		t.separatorExcRegex = regexp.MustCompile(splitExcRegex)
	}
}

// WithEOCExcRegex sets the regex string to use when deciding whether to add an EOC token after the last token.
func WithEOCExcRegex(eocRegex string) Option {
	return func(t *DefaultTokenizer) {
		t.eocExcRegex = regexp.MustCompile(eocRegex)
	}
}

// NewDefaultTokenizer creates a new tokenizer with default settings, which can be
// overridden by providing one or more Option functions.
func NewDefaultTokenizer(opts ...Option) *DefaultTokenizer {
	t := &DefaultTokenizer{
		separator: " ",
		eoc:       ".",
		// This regex finds sequences of word characters (letters, numbers, underscore)
		// OR single instances of common punctuation.
		separatorRegex: regexp.MustCompile(`[\w']+|[.,!?;]`),
		// This regex checks if a token is one of the sentence-ending punctuation marks.
		eocRegex: regexp.MustCompile(`^[.!?]$`),
		// This regex checks for characters that don't get a separator put before them.
		separatorExcRegex: regexp.MustCompile(`^[.,!?;]`),
		// This regex checks for characters that don't get an EOC put after them.
		eocExcRegex: regexp.MustCompile(`^[.,!?;]`),
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// Separator Returns the configured separator string.
func (t *DefaultTokenizer) Separator(_, next string) string {
	if t.separatorExcRegex.MatchString(next) {
		return ""
	}
	return t.separator
}

// EOC Returns the configured end-of-chain replacement string.
func (t *DefaultTokenizer) EOC(last string) string {
	if t.eocExcRegex.MatchString(last) {
		return ""
	}
	return t.eoc
}

// NewStream Returns the stream processor.
func (t *DefaultTokenizer) NewStream(r io.Reader) StreamTokenizer {
	return &DefaultStreamTokenizer{
		scanner:    bufio.NewScanner(r),
		buffer:     []string{},
		splitRegex: t.separatorRegex,
		eosRegex:   t.eocRegex,
	}
}

// DefaultStreamTokenizer is the default implementation of the StreamTokenizer interface.
// It uses a bufio.Scanner and regular expressions to read and tokenize a stream.
type DefaultStreamTokenizer struct {
	scanner    *bufio.Scanner
	buffer     []string
	splitRegex *regexp.Regexp
	eosRegex   *regexp.Regexp
}

// Next returns the next token from the stream. It returns a Token and a nil error on
// success. When the stream is exhausted, it returns a nil Token and io.EOF.
// Any other error indicates a problem reading from the underlying stream.
func (s *DefaultStreamTokenizer) Next() (*Token, error) {
	for len(s.buffer) == 0 { // Loop until we have tokens
		if !s.scanner.Scan() {
			if err := s.scanner.Err(); err != nil {
				return nil, err
			}
			return nil, io.EOF
		}
		s.buffer = s.splitRegex.FindAllString(s.scanner.Text(), -1)
	}

	// We have tokens in the buffer. Process the next one.
	word := s.buffer[0]
	s.buffer = s.buffer[1:] // Consume the token

	// Return the word and whether it is an EOC token or not
	return &Token{Text: word, EOC: s.eosRegex.MatchString(word)}, nil
}
