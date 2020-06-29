package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		vstr     string
		expected string
	}{
		{"1.0.0", "1"},
		{"1.0.1", "1.0.1"},
		{"1.ga", "1"},
		{"1.0.final", "1"},
		{"1.final", "1"},
		{"1.0", "1"},
		{"1.", "1"},
		{"1-", "1"},
		{"1.0.0-foo.0.0", "1-foo"},
		{"1.0.0-0.0.0", "1"},
	}
	for _, test := range tests {
		v, err := New(test.vstr)
		if err != nil {
			t.Errorf("could not create new maven version: %v", err)
		}
		assert.Equal(t, test.expected, v.TrimmedString())
	}
}

func TestPadForComparison(t *testing.T) {
	tests := []struct {
		vstr string
		wstr string
	}{
		{"1.1.1", "1"},
		{"1.0", "1.1.2"},
		{"1.0", "1.1-2"},
		{"1.1.1", "1.1.2"},
		{"1-1.1", "1.1.2"},
	}
	for _, test := range tests {
		v, _ := New(test.vstr)
		w, _ := New(test.wstr)
		vp, wp := padForComparison(v, w)
		smaller := &vp
		bigger := &wp
		if len(w.fields) < len(v.fields) {
			smaller = &wp
			bigger = &vp
		}
		assert.Equal(t, len(vp.fields), len(wp.fields))
		for c := len(smaller.fields); c < len(bigger.fields); c++ {
			assert.Equal(t, bigger.fields[c].dot, smaller.fields[c].dot)
			if smaller.fields[c].dot {
				assert.Equal(t, "0", smaller.fields[c].value)
			} else {
				assert.Equal(t, "", smaller.fields[c].value)
			}
		}
	}
}

func TestVersions_Less(t *testing.T) {
	tests := []struct {
		lesser  string
		greater string
	}{
		{"1", "2"},
		{"1.5", "2"},
		{"1", "2.5"},
		{"1.0", "1.1"},
		{"1.1", "1.2"},
		{"1.0.0", "1.1"},
		{"1.1", "1.2.0"},

		{"1.1.2.alpha1", "1.1.2"},
		{"1.1.2.alpha1", "1.1.2.beta1"},
		{"1.1.2.beta1", "1.2"},

		{"1.0-alpha-1", "1.0"},
		{"1.0-alpha-1", "1.0-alpha-2"},
		{"1.0-alpha-2", "1.0-alpha-15"},
		{"1.0-alpha-1", "1.0-beta-1"},

		{"1.0-beta-1", "1.0-SNAPSHOT"},
		{"1.0-SNAPSHOT", "1.0"},
		{"1.0-alpha-1-SNAPSHOT", "1.0-alpha-1"},

		{"1.0", "1.0-1"},
		{"1.0-1", "1.0-2"},
		{"2.0", "2.0-1"},
		{"2.0.0", "2.0-1"},
		{"2.0-1", "2.0.1"},

		{"2.0.1-klm", "2.0.1-lmn"},
		{"2.0.1", "2.0.1-xyz"},
		{"2.0.1-xyz-1", "2.0.1-1-xyz"},

		{"2.0.1", "2.0.1-123"},
		{"2.0.1-xyz", "2.0.1-123"},

		{"1.2.3-10000000000", "1.2.3-10000000001"},
		{"1.2.3-1", "1.2.3-10000000001"},
		{"2.3.0-v200706262000", "2.3.0-v200706262130"}, // org.eclipse:emf:2.3.0-v200706262000
		// org.eclipse.wst.common_core.feature_2.0.0.v200706041905-7C78EK9E_EkMNfNOd2d8qq
		{"2.0.0.v200706041905-7C78EK9E_EkMNfNOd2d8qq", "2.0.0.v200706041906-7C78EK9E_EkMNfNOd2d8qq"},
		{"1", "2"},
		{"1.5", "2"},
		{"1", "2.5"},
		{"1.0", "1.1"},
		{"1.1", "1.2"},
		{"1.0.0", "1.1"},
		{"1.1", "1.2.0"},

		{"1.1.2.alpha1", "1.1.2"},
		{"1.1.2.alpha1", "1.1.2.beta1"},
		{"1.1.2.beta1", "1.2"},

		{"1.0-alpha-1", "1.0"},
		{"1.0-alpha-1", "1.0-alpha-2"},
		{"1.0-alpha-2", "1.0-alpha-15"},
		{"1.0-alpha-1", "1.0-beta-1"},

		{"1.0-beta-1", "1.0-SNAPSHOT"},
		{"1.0-SNAPSHOT", "1.0"},
		{"1.0-alpha-1-SNAPSHOT", "1.0-alpha-1"},

		{"1.0", "1.0-1"},
		{"1.0-1", "1.0-2"},
		{"2.0", "2.0-1"},
		{"2.0.0", "2.0-1"},
		{"2.0-1", "2.0.1"},

		{"2.0.1-klm", "2.0.1-lmn"},
		{"2.0.1", "2.0.1-xyz"},
		{"2.0.1-xyz-1", "2.0.1-1-xyz"},

		{"2.0.1", "2.0.1-123"},
		{"2.0.1-xyz", "2.0.1-123"},

		{"1.2.3-10000000000", "1.2.3-10000000001"},
		{"1.2.3-1", "1.2.3-10000000001"},
		{"2.3.0-v200706262000", "2.3.0-v200706262130"}, // org.eclipse:emf:2.3.0-v200706262000
		// org.eclipse.wst.common_core.feature_2.0.0.v200706041905-7C78EK9E_EkMNfNOd2d8qq
		{"2.0.0.v200706041905-7C78EK9E_EkMNfNOd2d8qq", "2.0.0.v200706041906-7C78EK9E_EkMNfNOd2d8qq"},

		{"1-SNAPSHOT", "2-SNAPSHOT"},
		{"1.5-SNAPSHOT", "2-SNAPSHOT"},
		{"1-SNAPSHOT", "2.5-SNAPSHOT"},
		{"1.0-SNAPSHOT", "1.1-SNAPSHOT"},
		{"1.1-SNAPSHOT", "1.2-SNAPSHOT"},
		{"1.0.0-SNAPSHOT", "1.1-SNAPSHOT"},
		{"1.1-SNAPSHOT", "1.2.0-SNAPSHOT"},

		{"1.0-alpha-1-SNAPSHOT", "1.0-SNAPSHOT"},
		{"1.0-alpha-1-SNAPSHOT", "1.0-alpha-2-SNAPSHOT"},
		{"1.0-alpha-1-SNAPSHOT", "1.0-beta-1-SNAPSHOT"},

		{"1.0-beta-1-SNAPSHOT", "1.0-SNAPSHOT-SNAPSHOT"},
		{"1.0-SNAPSHOT-SNAPSHOT", "1.0-SNAPSHOT"},
		{"1.0-alpha-1-SNAPSHOT-SNAPSHOT", "1.0-alpha-1-SNAPSHOT"},

		{"1.0-SNAPSHOT", "1.0-1-SNAPSHOT"},
		{"1.0-1-SNAPSHOT", "1.0-2-SNAPSHOT"},
		{"2.0-SNAPSHOT", "2.0-1-SNAPSHOT"},
		{"2.0.0-SNAPSHOT", "2.0-1-SNAPSHOT"},
		{"2.0-1-SNAPSHOT", "2.0.1-SNAPSHOT"},

		{"2.0.1-klm-SNAPSHOT", "2.0.1-lmn-SNAPSHOT"},
		{"2.0.1-SNAPSHOT", "2.0.1-123-SNAPSHOT"},
		{"2.0.1-xyz-SNAPSHOT", "2.0.1-123-SNAPSHOT"},
		{"1.0-RC1", "1.0-SNAPSHOT"},
		{"1.0-rc1", "1.0-SNAPSHOT"},
		{"1.0-rc-1", "1.0-SNAPSHOT"},
	}
	for _, test := range tests {
		i, err := New(test.lesser)
		if err != nil {
			t.Errorf("error creating version %s: %v", test.lesser, err)
		}
		j, err := New(test.greater)
		if err != nil {
			t.Errorf("error creating version %s: %v", test.greater, err)
		}
		v := Versions([]Version{i, j})

		assert.True(t, v.Less(0, 1), "incorrect comparison %s - %s\n%+v\n%+v",
			test.lesser, test.greater, i, j)
	}
}

func TestParseRequirements(t *testing.T) {
	tests := []struct {
		req string
	}{
		//1.0: "Soft" requirement on 1.0 (just a recommendation, if it matches all other ranges for the dependency)
		//[1.0]: "Hard" requirement on 1.0
		//(,1.0]: x <= 1.0
		//[1.2,1.3]: 1.2 <= x <= 1.3
		//[1.0,2.0): 1.0 <= x < 2.0
		//[1.5,): x >= 1.5
		//(,1.0],[1.2,): x <= 1.0 or x >= 1.2; multiple sets are comma-separated
		//(,1.1),(1.1,)
		{"1.0"},
		{"[1.0]"},
		{"(,1.0]"},
		{"[1.2,1.3]"},
		{"[1.0,2.0)"},
		{"[1.5,)"},
		{"(,1.0],[1.2,)"},
		{"(,1.1),(1.1,)"},
	}
	for _, test := range tests {
		c, err := parseRequirement(test.req)
		if err != nil {
			t.Errorf("could not parse requirement %s: %v", test.req, err)
		}
		assert.True(t, len(c) > 0)
		//t.Logf("%s\n%+v\n\n", test.req, c)
	}
}

func TestVersion_Equal(t *testing.T) {
	tests := []struct {
		v string
		w string
	}{
		{"1", "1"},
		{"1", "1.0"},
		{"1", "1.0.0"},
		{"1.0", "1.0.0"},
		{"1", "1-0"},
		{"1", "1.0-0"},
		{"1.0", "1.0-0"},
		// no separator between number and character
		{"1a", "1-a"},
		{"1a", "1.0-a"},
		{"1a", "1.0.0-a"},
		{"1.0a", "1-a"},
		{"1.0.0a", "1-a"},
		{"1x", "1-x"},
		{"1x", "1.0-x"},
		{"1x", "1.0.0-x"},
		{"1.0x", "1-x"},
		{"1.0.0x", "1-x"},

		// aliases
		{"1ga", "1"},
		{"1final", "1"},
		{"1cr", "1rc"},

		// special "aliases" a, b and m for alpha, beta and milestone
		{"1a1", "1-alpha-1"},
		{"1b2", "1-beta-2"},
		{"1m3", "1-milestone-3"},

		// case insensitive
		{"1X", "1x"},
		{"1A", "1a"},
		{"1B", "1b"},
		{"1M", "1m"},
		{"1Ga", "1"},
		{"1GA", "1"},
		{"1Final", "1"},
		{"1FinaL", "1"},
		{"1FINAL", "1"},
		{"1Cr", "1Rc"},
		{"1cR", "1rC"},
		{"1m3", "1Milestone3"},
		{"1m3", "1MileStone3"},
		{"1m3", "1MILESTONE3"},
	}
	for _, test := range tests {
		v, err := New(test.v)
		if err != nil {
			t.Errorf("could not create version from %s: %v", test.v, err)
		}
		w, err := New(test.w)
		if err != nil {
			t.Errorf("could not create version from %s: %v", test.w, err)
		}
		assert.True(t, v.Equal(w), "%s not evaluated as equal to %s", test.v, test.w)
	}
}

func TestVersion_Satisfies(t *testing.T) {
	tests := []struct {
		req       string
		version   string
		satisfies bool
	}{
		{"[1.0]", "1.0", true},
		{"[1.0]", "0.9", false},
		{"[1.0]", "1.1", false},
		{"(,1.0]", "0.5", true},
		{"(,1.0]", "1.0", true},
		{"(,1.0]", "1.1", false},
		{"(,1.0)", "0.9", true},
		{"(,1.0)", "1.0", false},
		{"(,1.0)", "1.1", false},
		{"(1.0,)", "1.0", false},
		{"(1.0,)", "1.1", true},
		{"(1.0,)", "2.0", true},
		{"[1.0,)", "1.0", true},
		{"[1.0,)", "1.1", true},
		{"[1.0,)", "2.0", true},
		{"[1.2,1.3]", "1.2", true},
		{"[1.2,1.3]", "1.2.5", true},
		{"[1.2,1.3]", "1.3", true},
		{"[1.2,1.3]", "1.1", false},
		{"[1.2,1.3]", "1.4", false},
		{"(1.2,1.3)", "1.2", false},
		{"(1.2,1.3)", "1.2.5", true},
		{"(1.2,1.3)", "1.3", false},
		{"(1.2,1.3)", "1.1", false},
		{"(1.2,1.3)", "1.4", false},
		{"(1.2,1.3]", "1.2", false},
		{"(1.2,1.3]", "1.2.5", true},
		{"(1.2,1.3]", "1.3", true},
		{"(1.2,1.3]", "1.1", false},
		{"(1.2,1.3]", "1.4", false},
		{"(,1.0],[1.2,)", "0.5", true},
		{"(,1.0],[1.2,)", "1.0", true},
		{"(,1.0],[1.2,)", "1.1", false},
		{"(,1.0],[1.2,)", "1.2", true},
		{"(,1.0],[1.2,)", "1.3", true},
		{"(,1.1),(1.1,)", "0.5", true},
		{"(,1.1),(1.1,)", "1.0", true},
		{"(,1.1),(1.1,)", "1.1", false},
		{"(,1.1),(1.1,)", "1.1.1", true},
		{"(,1.1),(1.1,)", "2.0", true},
	}
	for _, test := range tests {
		v, err := New(test.version)
		if err != nil {
			t.Errorf("error creating version %s: %v", test.version, err)
		}
		assert.Equal(t, test.satisfies, v.Satisfies(test.req), "should version %s satisfy %s? %t ; but test does not agree.", test.version, test.req, test.satisfies)
	}
}

func TestParseRequirement_Invalid(t *testing.T) {
	tests := []string{
		"(1.0)",
		"[1.0)",
		"(1.0]",
		"(1.0,1.0]",
		"[1.0,1.0)",
		"(1.0,1.0)",
		"[1.1,1.0]",
		"[1.0,1.2),1.3",
		// overlap
		//"[1.0,1.2),(1.1,1.3]" ,
		//// overlap
		//"[1.1,1.3),(1.0,1.2]" ,
		//// ordering
		//"(1.1,1.2],[1.0,1.1)" ,
	}
	for _, test := range tests {
		_, err := parseRequirement(test)
		assert.NotNil(t, err, "did not error on invalid requirement: %s", test)
	}
}

func TestNormaliseVersion(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{"1-1.foo-bar1baz-.1", "1-1.foo-bar-1-baz-0.1"},
	}
	for _, test := range tests {
		norm := normaliseVersion(test.in)
		assert.Equal(t, test.out, norm, "normalisation of %s incorrect", test.in)
	}
}

func TestHyphenateAlphaNumeric(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{"foo1bar", "foo-1-bar"},
		{"foo-1bar", "foo-1-bar"},
		{"foo-1-bar", "foo-1-bar"},
		{"foo1", "foo-1"},
		{"1bar", "1-bar"},
		{"foo-1-bar2foo", "foo-1-bar-2-foo"},
		{"foo123bar", "foo-123-bar"},
		{"foo-123bar", "foo-123-bar"},
		{"foo-123-bar", "foo-123-bar"},
		{"foo123", "foo-123"},
		{"123bar", "123-bar"},
		{"foo-bar-1baz-0", "foo-bar-1-baz-0"},
	}
	for _, test := range tests {
		norm := hyphenateAlphaNumeric(test.in)
		assert.Equal(t, test.out, norm, "hypenating of %s incorrect", test.in)
	}
}
