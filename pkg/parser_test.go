package codeowners

import (
	"reflect"
	"strings"
	"testing"
)

func TestSimpleParser(t *testing.T) {
	testcases := []struct {
		input  string
		output *Entry
	}{
		{
			input: "#This is a code owners file",
			output: &Entry{
				path:    "",
				comment: "#This is a code owners file",
				suffix:  PathSufix(None),
				owners:  make([]string, 0),
			},
		},
		{
			input: "app/lib/index.html @product",
			output: &Entry{
				path:   "app/lib/index.html",
				suffix: PathSufix(Absolute),
				owners: []string{"@product"},
			},
		},
		{
			input: "app/lib/index.html @product @leadership",
			output: &Entry{
				path:   "app/lib/index.html",
				suffix: PathSufix(Absolute),
				owners: []string{"@product", "@leadership"},
			},
		},
		{
			input: "app/lib/index.html @product alecharmon@outlook.com",
			output: &Entry{
				path:   "app/lib/index.html",
				suffix: PathSufix(Absolute),
				owners: []string{"@product", "alecharmon@outlook.com"},
			},
		},
		{
			input: "app/lib/* @product alecharmon@outlook.com",
			output: &Entry{
				path:   "app/lib/*",
				suffix: PathSufix(Flat),
				owners: []string{"@product", "alecharmon@outlook.com"},
			},
		},
		{
			input: "app/lib/*.js @product alecharmon@outlook.com",
			output: &Entry{
				path:   "app/lib/*.js",
				suffix: PathSufix(Type),
				owners: []string{"@product", "alecharmon@outlook.com"},
			},
		},
		{
			input: "app/lib/ @product alecharmon@outlook.com",
			output: &Entry{
				path:   "app/lib/",
				suffix: PathSufix(Recursive),
				owners: []string{"@product", "alecharmon@outlook.com"},
			},
		},
		{
			input: "app/lib/ @product alecharmon@outlook.com #some comment",
			output: &Entry{
				path:    "app/lib/",
				suffix:  PathSufix(Recursive),
				comment: "#some comment",
				owners:  []string{"@product", "alecharmon@outlook.com"},
			},
		},
	}
	for _, tc := range testcases {
		p := NewParser(strings.NewReader(tc.input))
		got, err := p.Parse()

		if err != nil {
			t.Errorf(`%s err: %v, want no error`, tc.input, err)
		}
		if !reflect.DeepEqual(got, tc.output) {
			t.Errorf("Input %s: got, \n %#v \n want \n %#v", tc.input, got, tc.output)
		}
	}
}
