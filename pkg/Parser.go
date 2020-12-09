package codeowners

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
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

func NewEntry() *Entry {
	return &Entry{
		owners: make([]string, 0),
		suffix: PathSufix(None),
	}
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

// NewParser returns a new instance of Parser.
func NewParser(r io.Reader) *Parser {
	return &Parser{s: NewScanner(r)}
}

// Parse parses a line from a codeowners file.
func (p *Parser) Parse() (*Entry, error) {
	entry := NewEntry()
	lineByte, err := ioutil.ReadAll(p.s.r)
	if err != nil {
		log.Fatal(err)
	}
	line := string(lineByte)
	if line[0] == '#' {
		entry.comment = line
		entry.suffix = PathSufix(None)
		return entry, nil
	}

	parts := strings.Fields(line)

	if len(parts) < 2 {
		if isValidOwner(parts[0]) {
			return nil, errors.New("Missing path for entry")
		}
		return nil, errors.New("Invalid entry")
	}

	path := parts[0]

	if path[0] == '/' {
		path = path[1:]
	}

	entry.path = path
	entry.suffix = DetermineSuffix(entry.path)

	for i, p := range parts[1:] {
		if p[0] == '#' {
			entry.comment = strings.Join(parts[i+1:], " ")
			return entry, nil
		}
		if isValidOwner(p) == false {
			return nil, fmt.Errorf("(%s) is an invalid owner", p)
		}
		entry.owners = append(entry.owners, p)
	}
	return entry, nil
}

func isValidOwner(owner string) bool {
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
