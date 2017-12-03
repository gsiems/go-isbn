package isbn

import (
	"os"
	"testing"
)

func TestLoadRangeData(t *testing.T) {

	xml_file := os.Getenv("ISBN_RANGE_FILE")
	// "$HOME/Downloads/RangeMessage.xml"
	if xml_file == "" {
		t.Errorf("ISBN_RANGE_FILE Env variable not set")
	}

	// Before anything is loaded, the HasRangeData should return false
	want := false
	got := HasRangeData()
	if got != want {
		t.Errorf("HasRangeData() == %t, want %t", got, want)
	}

	// Try loading the range data
	want = true
	got, err := LoadRangeData(xml_file)
	if err != nil {
		t.Errorf("LoadRangeData(%q) == %t, want %t (%q)", xml_file, got, want, err)
	} else if got != want {
		t.Errorf("LoadRangeData(%q) == %t, want %t", xml_file, got, want)
	}

	// TODO: do we want/need to check the size of the range dataset?

	// Now re-load the range data
	want = true
	got, err = LoadRangeData(xml_file)
	if err != nil {
		t.Errorf("LoadRangeData(%q) == %t, want %t (%q)", xml_file, got, want, err)
	} else if got != want {
		t.Errorf("LoadRangeData(%q) == %t, want %t", xml_file, got, want)
	}

	// After loading, the HasRangeData should return true
	want = true
	got = HasRangeData()
	if got != want {
		t.Errorf("HasRangeData() == %t, want %t", got, want)
	}

}
