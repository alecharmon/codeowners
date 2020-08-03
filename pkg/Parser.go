package codeowners

import (
	"bytes"
	"io"
	"path/filepath"
	"regexp"
	"strings"
)

type Entry struct {
	path    string
	suffix  PathSufix
	comment string
	owners  []string
}

// Parser represents a parser.
type Parser struct {
	s   *Scanner
	buf struct {
		tok Token  // last read token
		lit string // last read literal
		n   int    // buffer size (max=1)
	}
}

func newEntry() *Entry {
	return &Entry{
		owners: make([]string, 0),
		suffix: PathSufix(None),
	}
}

// NewParser returns a new instance of Parser.
func NewParser(r io.Reader) *Parser {
	return &Parser{s: NewScanner(r)}
}

// Parse parses a line from a codeowners file.
func (p *Parser) Parse() (*Entry, error) {
	entry := newEntry()

	tok, lit := p.scanIgnoreWhitespace()

	//Capture the comment (entire line is a comment)
	if tok == HASH {
		var b bytes.Buffer
		b.WriteString(lit)
		for {
			tok, lit = p.scan()
			if tok == EOF {
				break
			}
			b.WriteString(lit)
		}
		entry.comment = b.String()
		return entry, nil
	}

	var b bytes.Buffer
	for tok != WS {
		b.WriteString(lit)
		tok, lit = p.scan()
	}
	//TODO: Validate/normalize file path?
	path := b.String()
	if len(path) > 1 && []rune(path)[0] == '/' {
		path = path[1:]
	}

	entry.path = path
	entry.suffix = DetermineSuffix(entry.path)

	tok, lit = p.scanIgnoreWhitespace()
	for tok != EOF {
		b.Reset()
		for tok != WS {
			if tok == EOF {
				break
			}
			if tok == HASH {
				b.Reset()
				b.WriteString(lit)
				for {
					tok, lit = p.scan()
					if tok == EOF {
						break
					}
					b.WriteString(lit)
				}
				entry.comment = b.String()
				b.Reset()
				break
			}
			b.WriteString(lit)
			tok, lit = p.scan()
		}
		if owner := b.String(); isvalidOwner(owner) {
			entry.owners = append(entry.owners, owner)
		}
		tok, lit = p.scanIgnoreWhitespace()
	}

	// Return the successfully parsed statement.
	return entry, nil
}

func isvalidOwner(owner string) bool {
	if len(owner) < 1 || len(owner) > 254 {
		return false
	}

	//Does the owner start with a @ and only have one => @owner-name
	if strings.Index(owner, "@") == 0 && strings.Index(owner, "@") == strings.LastIndex(owner, "@") {
		return true
	}

	//Is it a valid email
	var rxEmail = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	if rxEmail.MatchString(owner) {
		return true
	}

	return false
}

// DetermineSuffix assings a sufix to a given path
func DetermineSuffix(path string) PathSufix {
	base := filepath.Base(path)
	ext := filepath.Ext(path)

	baseFirstCh := []rune(base)[0]
	if baseFirstCh == '*' && ext != "" {
		return PathSufix(Type)
	}

	if baseFirstCh == '*' && ext == "" {
		return PathSufix(Flat)
	}

	if pathR := []rune(path); pathR[len(pathR)-1] == '/' {
		return PathSufix(Recursive)
	}

	return PathSufix(Absolute)
}

func (p *Parser) scan() (tok Token, lit string) {
	// If we have a token on the buffer, then return it.
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}

	// Otherwise read the next token from the scanner.
	tok, lit = p.s.Scan()

	// Save it to the buffer in case we unscan later.
	p.buf.tok, p.buf.lit = tok, lit

	return
}

// scanIgnoreWhitespace scans the next non-whitespace token.
func (p *Parser) scanIgnoreWhitespace() (tok Token, lit string) {
	tok, lit = p.scan()
	if tok == WS {
		tok, lit = p.scan()
	}
	return
}

// unscan pushes the previously read token back onto the buffer.
func (p *Parser) unscan() { p.buf.n = 1 }
