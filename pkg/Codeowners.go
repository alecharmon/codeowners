package codeowners

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"

	"io"
	"log"
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

// BuildEntries ...
func BuildEntries(input []byte, includeComments bool) ([]*Entry, error) {
	entries := []*Entry{}
	reader := bufio.NewReader(bytes.NewReader(input))

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
	return entries, nil
}

// BuildEntriesFromFile from an file path, absolute or relative, builds the entries for the CODEOWNERS file
func BuildEntriesFromFile(filePath string, includeComments bool) ([]*Entry, error) {
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return BuildEntries(bytes, includeComments)
}

func newNode() *node {
	return &node{
		entries: []*Entry{},
	}
}

// BuildFromFile from an file path, absolute or relative, builds the index for the CODEOWNERS file
func BuildFromFile(filePath string) (*CodeOwners, error) {
	entries, err := BuildEntriesFromFile(filePath, false)
	if err != nil {
		return nil, err
	}
	return createIndexFromEntries(entries)
}

// BuildFromFile from an file path, absolute or relative, builds the index for the CODEOWNERS file
func BuildIndex(input []byte) (*CodeOwners, error) {
	entries, err := BuildEntries(input, false)
	if err != nil {
		return nil, err
	}
	return createIndexFromEntries(entries)
}

// createIndexFromEntries ...
func createIndexFromEntries(entries []*Entry) (*CodeOwners, error) {
	t := &CodeOwners{
		trie.NewPathTrie(),
	}

	for _, entry := range entries {
		t.addOwnerByEntry(entry)
	}

	return t, nil
}

func (n *node) addEntry(e *Entry) {
	n.entries = append(n.entries, e)
}

func (t *CodeOwners) addOwnerByEntry(entry *Entry) {
	var n *node
	var ok bool
	value := t.Get(entry.path)
	if value == nil {
		n = newNode()
	} else {
		n, ok = value.(*node)
		if !ok {
			out, _ := fmt.Printf("%v, %v", ok, n)
			panic(out)
		}
	}

	n.addEntry(entry)
	path := entry.path
	if []rune(path)[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	t.Put(path, n)
}

func (t *CodeOwners) AddOwner(path string, owners ...string) {
	t.addOwnerByEntry(&Entry{
		path:   path,
		owners: owners,
		suffix: DetermineSuffix(path),
	})
}

func (t *CodeOwners) RemovePath(path string) {
	t.Put(path, nil)
}

func (t *CodeOwners) RemoveOwner(owner string) {
	walker := func(key string, value interface{}) error {
		if value == nil {
			return nil
		}
		n, ok := value.(*node)
		if !ok {
			log.Fatal("Structure of the code owner index is malformed")
		}
		for _, en := range n.entries {
			newOwners := en.owners
			for i, o := range en.owners {
				if o == owner {
					RemoveIndex(newOwners, i)
				}
			}
			en.owners = newOwners
		}

		return nil
	}
	t.Walk(walker)
}

func (t *CodeOwners) ReplaceOwner(oldOwner, newOwner string) {
	walker := func(key string, value interface{}) error {
		if value == nil {
			return nil
		}
		n, ok := value.(*node)
		if !ok {
			log.Fatal("Structure of the code owner index is malformed")
		}
		for _, en := range n.entries {
			newOwners := en.owners
			for i, o := range en.owners {
				if o == oldOwner {
					newOwners[i] = newOwner
				}
			}
			en.owners = newOwners
		}

		return nil
	}
	t.Walk(walker)
}

// FindOwners ...
func (t *CodeOwners) FindOwners(path string) []string {
	owners := []string{}
	walker := func(key string, value interface{}) error {
		if value == nil {
			return nil
		}
		n, ok := value.(*node)
		if !ok {
			panic("Structure of the index is malformed")
		}
		for _, en := range n.entries {
			if en.suffix == PathSufix(Recursive) || en.suffix == PathSufix(Absolute) || en.suffix == PathSufix(Flat) {
				owners = append(owners, en.owners...)
			}
		}

		return nil
	}
	t.WalkKey(path, walker)

	//get the base wild card type
	ext := "*" + filepath.Ext(path)
	extEntry := t.Get(ext)
	if extEntry != nil {
		n, ok := extEntry.(*node)
		if !ok {
			log.Fatal("Structure of the code owner index is malformed")
		}
		for _, en := range n.entries {
			owners = append(owners, en.owners...)
		}
	}

	//get the base wild card
	ext = "*"
	extEntry = t.Get(ext)
	if extEntry != nil {
		n, ok := extEntry.(*node)
		if !ok {
			log.Fatal("Structure of the code owner index is malformed")
		}
		for _, en := range n.entries {
			owners = append(owners, en.owners...)
		}
	}
	return removeDuplicatesUnordered(owners)
}

func (t *CodeOwners) Print() {
	var b bytes.Buffer
	t.serialize(&b)
	fmt.Print(b.String())
}

func (t *CodeOwners) serialize(b *bytes.Buffer) {
	toSort := []string{}
	walker := func(key string, value interface{}) error {
		if value == nil {
			return nil
		}
		n, ok := value.(*node)
		if !ok {
			panic("Structure of the index is malformed")
		}

		for _, en := range n.entries {
			if en.comment != "" {
				toSort = append(toSort, fmt.Sprintf("%s %s #%s", en.path, strings.Join(en.owners, " "), en.comment))
			}
			toSort = append(toSort, fmt.Sprintf("%s %s", en.path, strings.Join(en.owners, " ")))
		}

		return nil
	}
	t.Walk(walker)
	sort.Strings(toSort)

	b.WriteString(strings.Join(toSort, "\n"))
}

func removeDuplicatesUnordered(elements []string) []string {
	encountered := map[string]bool{}

	// Create a map of all unique elements.
	for v := range elements {
		encountered[elements[v]] = true
	}

	// Place all keys from the map into a slice.
	result := []string{}
	for key := range encountered {
		result = append(result, key)
	}
	return result
}

func RemoveIndex(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}
