package isbn

import (
	"os"
	"testing"
)

func TestISBN(t *testing.T) {

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

	cases := []struct {
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
		{"9780590132053", true},
        {"978-8891230195", true}.
		{"9780590132053F", false},
		{"9780590732053", false},
		{"081666303x", false},
		{"", false},
	}
	for _, c := range cases {
		_, err := ParseISBN(c.in)
		if err != nil && c.want {
			t.Errorf("ParseISBN(%q) == fail, want success (%q)", c.in, err)
		} else if err == nil && !c.want {
			t.Errorf("ParseISBN(%q) == success, want fail", c.in)
		}
	}

	_, _ = UnloadRangeData()
}
