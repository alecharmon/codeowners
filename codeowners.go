package codeowners

import (
	"bufio"
	"fmt"

	"io"
	"log"
	"os"
	"strings"

	"github.com/alecharmon/trie"
)

type node struct {
	entries []*Entry
}

// CodeOwners search index for a CODEOWNER file
type CodeOwners struct {
	*trie.PathTrie
}

func BuildEntriesFromFile(filePath string, includeComments bool) []*Entry {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	entries := []*Entry{}
	reader := bufio.NewReader(file)

	for {
		line, _, err := reader.ReadLine()

		if err == io.EOF {
			break
		}
		if len(line) < 1 {
			continue
		}

		parser := NewParser(strings.NewReader(string(line)))
		entry, err := parser.Parse()
		if err != nil {
			panic(err)
		}

		if (entry.suffix == PathSufix(None)) && !includeComments {
			continue
		}

		entries = append(entries, entry)
	}
	return entries
}

func newNode() *node {
	return &node{
		entries: []*Entry{},
	}
}

func BuildFromFile(filePath string) *CodeOwners {

	t := &CodeOwners{
		trie.NewPathTrie(),
	}
	var n *node
	var ok bool

	for _, entry := range BuildEntriesFromFile(filePath, false) {
		value := t.Get(entry.path)
		if value == nil {
			n = newNode()
		} else {
			//n, ok = value.(*node)
			if !ok {
				out, _ := fmt.Printf("%v, %v", ok, n)
				panic(out)
			}
		}

		n.addEntry(entry)
		t.Put(entry.path, n)
	}

	return t
}

func (n *node) addEntry(e *Entry) {
	n.entries = append(n.entries, e)
}

func (t *CodeOwners) findOwners(filepath string) []string {
	owners := []string{}
	walker := func(key string, value interface{}) error {
		n, ok := value.(*node)
		if !ok {
			panic("Structure of the code owner index is malformed")
		}

		for _, en := range n.entries {
			owners = append(owners, en.owners...)
		}

		return nil
	}
	t.WalkKey(filepath, walker)
	return owners
}
