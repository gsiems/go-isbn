// Copyright 2017 Gregory Siems. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package isbn

import (
	"os"
	"testing"
)

func TestLoadRangeData(t *testing.T) {

	xmlFile := os.Getenv("ISBN_RANGE_FILE")
	if xmlFile == "" {
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
	got, err := LoadRangeData(xmlFile)
	if err != nil {
		t.Errorf("LoadRangeData(%q) == %t, want %t (%q)", xmlFile, got, want, err)
	} else if got != want {
		t.Errorf("LoadRangeData(%q) == %t, want %t", xmlFile, got, want)
	}

	// TODO: do we want/need to check the size of the range dataset?

	// Now re-load the range data
	want = true
	got, err = LoadRangeData(xmlFile)
	if err != nil {
		t.Errorf("LoadRangeData(%q) == %t, want %t (%q)", xmlFile, got, want, err)
	} else if got != want {
		t.Errorf("LoadRangeData(%q) == %t, want %t", xmlFile, got, want)
	}

	// After loading, the HasRangeData should return true
	want = true
	got = HasRangeData()
	if got != want {
		t.Errorf("HasRangeData() == %t, want %t", got, want)
	}

	// Ensure that unloading also works
	want = true
	got, err = UnloadRangeData()
	if err != nil {
		t.Errorf("UnloadRangeData() == %t, want %t (%q)", got, want, err)
	} else if got != want {
		t.Errorf("UnloadRangeData() == %t, want %t", got, want)
	}
}
