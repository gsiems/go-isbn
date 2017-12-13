// Copyright 2017 Gregory Siems. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package isbn

import (
	"fmt"
	"os"
	"testing"
)

func TestISBN(t *testing.T) {

	// Test the ISBN parsing/validation without range data
	cases0 := []struct {
		in   string
		want bool
	}{
		{"88 04 47328 2", false},
		{"978-8804473282", false},
		{"0547928246", false},
		{"978-0547928241", false},
		{"978 0670013951", false},
		{"089686281x", false},
		{"9780822527602", false},
		{"9780590d32053", false},
		{"978-8891230195", false},
		{"9780590132053F", false},
		{"9780590732053", false},
		{"081666303x", false},
		{"", false},
	}
	for _, c := range cases0 {
		_, err := ParseISBN(c.in)
		if err != nil && c.want {
			t.Errorf("ParseISBN(%q) == fail, want success (%q)", c.in, err)
		} else if err == nil && !c.want {
			t.Errorf("ParseISBN(%q) == success, want fail", c.in)
		}
	}

	// Ensure that the range data is loaded
	if !HasRangeData() {
		xml_file := os.Getenv("ISBN_RANGE_FILE")
		// "$HOME/Downloads/RangeMessage.xml"
		if xml_file == "" {
			t.Errorf("ISBN_RANGE_FILE Env variable not set")
		}

		want := true
		got, err := LoadRangeData(xml_file)
		if err != nil {
			t.Errorf("LoadRangeData(%q) == %t, want %t (%q)", xml_file, got, want, err)
		} else if got != want {
			t.Errorf("LoadRangeData(%q) == %t, want %t", xml_file, got, want)
		}
	}

	// Test the ISBN CalcCheckDigit
	cases1 := []struct {
		in   string
		want string
	}{
		{"88 04 47328 2", "2"},
		{"978-8804473282", "2"},
		{"0547928246", "6"},
		{"978-0547928241", "1"},
		{"978 0670013951", "1"},
		{"089686281x", "X"},
		{"9780822527602", "2"},
		{"97805S0132053", ""},
		{"978-8891230195", "5"},
		{"978-059013205F", ""},
		{"9780590132053F", ""},
		{"9780590732053", "5"},
		{"081666303x", "3"},
		{"", ""},
	}
	for _, c := range cases1 {
		got, err := CalcCheckDigit(c.in)
		if c.want != got {
			if err != nil {
				t.Errorf("CalcCheckDigit(%q) == '%s', want '%s' (%q)", c.in, got, c.want, err)
			} else {
				t.Errorf("CalcCheckDigit(%q) == '%s', want '%s'", c.in, got, c.want)
			}
		}
	}

	// Test the ISBN parsing/validation
	cases2 := []struct {
		in   string
		want bool
	}{
		{"88 04 47328 2", true},
		{"978-8804473282", true},
		{"0547928246", true},
		{"978-0547928241", true},
		{"978 0670013951", true},
		{"089686281x", true},
		{"9780822527602", true},
		{"9780590d32053", false},
		{"978-8891230195", true},
		{"9780590132053F", false},
		{"9780590732053", false},
		{"081666303x", false},
		{"", false},
	}
	for _, c := range cases2 {
		_, err := ParseISBN(c.in)
		if err != nil && c.want {
			t.Errorf("ParseISBN(%q) == fail, want success (%q)", c.in, err)
		} else if err == nil && !c.want {
			t.Errorf("ParseISBN(%q) == success, want fail", c.in)
		}
	}

	// Test the ISBN Stringer interface
	cases3 := []struct {
		in   string
		want string
	}{
		{"88 04 47328 2", "978-88-04-47328-2 (88-04-47328-2)"},
		{"978-8804473282", "978-88-04-47328-2 (88-04-47328-2)"},
		{"0547928246", "978-0-547-92824-1 (0-547-92824-6)"},
		{"978-0547928241", "978-0-547-92824-1 (0-547-92824-6)"},
		{"978 0670013951", "978-0-670-01395-1 (0-670-01395-1)"},
		{"089686281x", "978-0-89686-281-4 (0-89686-281-X)"},
		{"9780822527602", "978-0-8225-2760-2 (0-8225-2760-X)"},
		{"9780590132053", "978-0-590-13205-3 (0-590-13205-9)"},
		{"978-8891230195", "978-88-912-3019-5 (88-912-3019-7)"},
		{"9780590132053F", ""},
		{"9780590732053", ""},
		{"081666303x", ""},
		{"0816bad66303x", ""},
		{"", ""},
	}
	for _, c := range cases3 {

		isbn, _ := ParseISBN(c.in)
		got := fmt.Sprint(isbn)
		if got != c.want {
			t.Errorf("Print() == %q, want %q", got, c.want)
		}
	}

	// Test the ISBN10 and ISBN13 conversion functionality
	cases4 := []struct {
		in     string
		want13 string
		want10 string
	}{
		{"88 04 47328 2", "9788804473282", "8804473282"},
		{"978-8804473282", "9788804473282", "8804473282"},
		{"0547928246", "9780547928241", "0547928246"},
		{"978-0547928241", "9780547928241", "0547928246"},
		{"978 0670013951", "9780670013951", "0670013951"},
		{"089686281x", "9780896862814", "089686281X"},
		{"9780822527602", "9780822527602", "082252760X"},
		{"9780590132053", "9780590132053", "0590132059"},
		{"978-8891230195", "9788891230195", "8891230197"},
		{"9780590132053F", "", ""},
		{"9780590732053", "", ""},
		{"081666303x", "", ""},
		{"", "", ""},
	}
	for _, c := range cases4 {
		isbn, _ := ParseISBN(c.in)

		got13 := isbn.ISBN13()
		if got13 != c.want13 {
			t.Errorf("ISBN13() == %q, want %q", got13, c.want13)
		}

		got10 := isbn.ISBN10()
		if got10 != c.want10 {
			t.Errorf("ISBN10() == %q, want %q", got10, c.want10)
		}
	}

	_, _ = UnloadRangeData()
}
