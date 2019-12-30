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
			path:    "/docs/",
			comment: "",
			suffix:  PathSufix(Recursive),
			owners:  []string{"@all-docs"},
		},
		&Entry{
			path:    "/docs/*",
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
			path:    "/config/",
			comment: "",
			suffix:  PathSufix(Recursive),
			owners:  []string{"@config-owner"},
		},
	}

	entries := BuildEntriesFromFile("fixtures/testCODEOWNERS_Rules", false)

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
	co := BuildFromFile("fixtures/testCODEOWNERS_Example")
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
	}

	for _, tc := range testcases {

		if out := co.findOwners(tc.input); !reflect.DeepEqual(out, tc.expected) {
			t.Errorf("%s : expected %v got %v", tc.input, tc.expected, out)
		}

	}
}
