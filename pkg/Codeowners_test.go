package codeowners

import (
	"reflect"
	"testing"
)

func TestBuildEntriesFromFile(t *testing.T) {
	outputs := []*Entry{
		&Entry{
			path:    "*",
			comment: "",
			suffix:  PathSufix(Flat),
			owners:  []string{"@default-codeowner"},
		},
		&Entry{
			path:    "*.rb",
			comment: "",
			suffix:  PathSufix(Type),
			owners:  []string{"@ruby-owner"},
		},
		&Entry{
			path:    "\\#file_with_pound.rb",
			comment: "",
			suffix:  PathSufix(Absolute),
			owners:  []string{"@owner-file-with-pound"},
		},
		&Entry{
			path:    "CODEOWNERS",
			comment: "",
			suffix:  PathSufix(Absolute),
			owners:  []string{"@multiple", "@code", "@owners"},
		},
		&Entry{
			path:    "LICENSE",
			comment: "",
			suffix:  PathSufix(Absolute),
			owners:  []string{"@legal", "janedoe@gitlab.com"},
		},
		&Entry{
			path:    "README",
			comment: "",
			suffix:  PathSufix(Absolute),
			owners:  []string{"@group", "@group/with-nested/subgroup"},
		},
		&Entry{
			path:    "docs/",
			comment: "",
			suffix:  PathSufix(Recursive),
			owners:  []string{"@all-docs"},
		},
		&Entry{
			path:    "docs/*",
			comment: "",
			suffix:  PathSufix(Flat),
			owners:  []string{"@root-docs"},
		},
		&Entry{
			path:    "lib/",
			comment: "",
			suffix:  PathSufix(Recursive),
			owners:  []string{"@lib-owner"},
		},
		&Entry{
			path:    "config/",
			comment: "",
			suffix:  PathSufix(Recursive),
			owners:  []string{"@config-owner"},
		},
	}

	entries, err := BuildEntriesFromFile("fixtures/testCODEOWNERS_Rules", false)

	if err != nil {
		t.Fatalf("expecting a non error")
		t.FailNow()
	}
	if len(entries) != len(outputs) {
		t.Fatalf("Expected output size of %d but got %d", len(outputs), len(entries))
		t.FailNow()
	}

	for i := range outputs {
		if !reflect.DeepEqual(entries[i], outputs[i]) {
			t.Errorf("Expected, \n %#v \n got \n %#v", outputs[i], entries[i])
		}

	}
}

func TestBuildFromFile(t *testing.T) {
	co, err := BuildFromFile("fixtures/testCODEOWNERS_Example")
	if err != nil {
		t.Fatalf("expecting a non error")
		t.FailNow()
	}
	testcases := []struct {
		input    string
		expected []string
	}{
		{
			input: "app/lib/network",
			expected: []string{
				"@a", "@b", "@c",
			},
		},
		{
			input: "app/vendor/hooli/",
			expected: []string{
				"@a", "@c",
			},
		},
		{
			input: "app/vendor/hooli/middle_out.go",
			expected: []string{
				"@a", "@c", "@richard",
			},
		},
		{
			input: "app/vendor/hooli/index.js",
			expected: []string{
				"@frontend", "@a", "@c", "@mike",
			},
		},
		{
			input: "app/vendor/hooli/index.react.js",
			expected: []string{
				"@frontend", "@a", "@c", "@mike",
			},
		},
	}

	for _, tc := range testcases {

		if out := co.FindOwners(tc.input); !sameStringSlice(out, tc.expected) {
			t.Errorf("%s : expected %v got %v", tc.input, tc.expected, out)
		}

	}
}

func TestBuildFromFileWildCard(t *testing.T) {
	co, err := BuildFromFile("fixtures/testCODEOWNERS_Example_Wildcard")
	if err != nil {
		t.Fatalf("expecting a non error")
		t.FailNow()
	}
	testcases := []struct {
		input    string
		expected []string
	}{
		{
			input: "app/lib/network",
			expected: []string{
				"@devs", "@a", "@b", "@c",
			},
		},
		{
			input: "app/vendor/hooli/",
			expected: []string{
				"@devs", "@a", "@c",
			},
		},
		{
			input: "app/vendor/package.json",
			expected: []string{
				"@devs", "@b", "@a",
			},
		},
		{
			input: "app/vendor/hooli/middle_out.go",
			expected: []string{
				"@devs", "@a", "@c", "@richard",
			},
		},
		{
			input: "app/vendor/hooli/index.js",
			expected: []string{
				"@devs", "@frontend", "@a", "@c", "@mike",
			},
		},
		{
			input: "app/vendor/hooli/index.react.js",
			expected: []string{
				"@devs", "@frontend", "@a", "@c", "@mike",
			},
		},
	}

	for _, tc := range testcases {

		if out := co.FindOwners(tc.input); !sameStringSlice(out, tc.expected) {
			t.Errorf("%s : expected %v got %v", tc.input, tc.expected, out)
		}

	}
}

func TestAddOwner(t *testing.T) {
	co, err := BuildFromFile("fixtures/testCODEOWNERS_Example_Wildcard")
	if err != nil {
		t.Fatalf("expecting a non error")
		t.FailNow()
	}
	testcases := []struct {
		path      string
		newOwners []string
	}{
		{
			path: "app/lib/concurrency/file.php",
			newOwners: []string{
				"@tod", "@marria", "@scott",
			},
		},
	}

	for _, tc := range testcases {
		co.AddOwner(tc.path, tc.newOwners...)
		out := co.FindOwners(tc.path)
		for _, newOwner := range tc.newOwners {
			if !contains(newOwner, out...) {
				t.Errorf("%s : expected new owners to have %v but did not", tc.path, newOwner)
			}
		}
	}
}

func TestAddOwnerSuffixs(t *testing.T) {
	co, err := BuildFromFile("fixtures/testCODEOWNERS_Example_Wildcard")
	if err != nil {
		t.Fatalf("expecting a non error")
		t.FailNow()
	}
	testcases := []struct {
		entryPath string
		testPath  string
		newOwners []string
	}{
		{
			entryPath: "app/lib/concurrency/*",
			testPath:  "app/lib/concurrency/fork.php",
			newOwners: []string{
				"@tod", "@marria", "@scott",
			},
		},
		{
			entryPath: "app/lib/concurrency/",
			testPath:  "app/lib/concurrency/fork.php",
			newOwners: []string{
				"@tod", "@marria", "@scott",
			},
		},
		{
			entryPath: "app/lib/concurrency/",
			testPath:  "app/lib/concurrency/tests/forktest.php",
			newOwners: []string{
				"@tod", "@marria", "@scott",
			},
		},
		{
			entryPath: "*.php",
			testPath:  "app/lib/concurrency/fork.php",
			newOwners: []string{
				"@tod", "@marria", "@scott",
			},
		},
	}

	for _, tc := range testcases {
		co.AddOwner(tc.entryPath, tc.newOwners...)
		out := co.FindOwners(tc.testPath)
		for _, newOwner := range tc.newOwners {
			if !contains(newOwner, out...) {
				t.Errorf(" entrypath: %s \n testpath: %s \n: expected new owners to have %v but did not", tc.entryPath, tc.testPath, newOwner)
			}
		}
	}
}

func TestRemoveOwner(t *testing.T) {
	co, err := BuildFromFile("fixtures/testCODEOWNERS_Example_Wildcard")
	if err != nil {
		t.Fatalf("expecting a non error")
		t.FailNow()
	}
	foo := co.FindOwners("app/lib/network/foo.php")
	if sameStringSlice(foo, []string{"@a", "@b", "@c"}) {
		t.Errorf("not expected owners for foo.php")
	}
	co.RemoveOwner("@c")
	foo = co.FindOwners("app/lib/network/foo.php")
	if sameStringSlice(foo, []string{"@a", "@b"}) {
		t.Errorf("not expected owners for foo.php")
	}

	foo = co.FindOwners("/app/vendor/hooli/middle_out.go")
	if sameStringSlice(foo, []string{"@devs"}) {
		t.Errorf("not expected owners for foo.php")
	}
}

func TestRemovePath(t *testing.T) {
	co, err := BuildFromFile("fixtures/testCODEOWNERS_Example_Wildcard")
	if err != nil {
		t.Fatalf("expecting a non error")
		t.FailNow()
	}

	foo := co.FindOwners("app/vendor/hooli/middle_out.go")

	if !sameStringSlice(foo, []string{"@a", "@c", "@richard", "@devs"}) {
		t.Errorf("not expected owners for foo.php %v", foo)
	}

	co.RemovePath("app/vendor/hooli/middle_out.go")
	foo = co.FindOwners("app/vendor/hooli/middle_out.go")
	if !sameStringSlice(foo, []string{"@a", "@c", "@devs"}) {
		t.Errorf("not expected owners for foo.php %v", foo)
	}

	co.RemovePath("*")
	foo = co.FindOwners("app/vendor/hooli/middle_out.go")
	if !sameStringSlice(foo, []string{"@a", "@c"}) {
		t.Errorf("not expected owners for foo.php %v", foo)
	}
}

func TestReplaceOwner(t *testing.T) {
	co, err := BuildFromFile("fixtures/testCODEOWNERS_Example_Wildcard")
	if err != nil {
		t.Fatalf("expecting a non error")
		t.FailNow()
	}

	//base
	foo := co.FindOwners("app/vendor/hooli/middle_out.go")
	if !sameStringSlice(foo, []string{"@a", "@c", "@richard", "@devs"}) {
		t.Errorf("not expected owners for foo.php %v", foo)
	}

	co.ReplaceOwner("@devs", "@eng")
	foo = co.FindOwners("app/vendor/hooli/middle_out.go")
	if !sameStringSlice(foo, []string{"@a", "@c", "@richard", "@eng"}) {
		t.Errorf("not expected owners for foo.php %v", foo)
	}

	co.ReplaceOwner("@richard", "@scott")
	foo = co.FindOwners("app/vendor/hooli/middle_out.go")
	if !sameStringSlice(foo, []string{"@a", "@c", "@scott", "@eng"}) {
		t.Errorf("not expected owners for foo.php %v", foo)
	}
}

func TestPrint(t *testing.T) {
	co, err := BuildFromFile("fixtures/testCODEOWNERS_Example_Wildcard")
	if err != nil {
		t.Fatalf("expecting a non error")
		t.FailNow()
	}
	co.Print()
}

func contains(str string, arr ...string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

// borrowed from https://stackoverflow.com/questions/36000487/check-for-equality-on-slices-without-order
func sameStringSlice(x, y []string) bool {
	if len(x) != len(y) {
		return false
	}
	// create a map of string -> int
	diff := make(map[string]int, len(x))
	for _, _x := range x {
		// 0 value for int is 0, so just increment a counter for the string
		diff[_x]++
	}
	for _, _y := range y {
		// If the string _y is not in diff bail out early
		if _, ok := diff[_y]; !ok {
			return false
		}
		diff[_y]--
		if diff[_y] == 0 {
			delete(diff, _y)
		}
	}
	if len(diff) == 0 {
		return true
	}
	return false
}
