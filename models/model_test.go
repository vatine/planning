package models

import (
	//"strings"
	"testing"
)

func TestNoDeps(t *testing.T) {
	used := make(map[string]bool)
	used["used1"] = true
	used["used2"] = true
	used["used3"] = true

	td := []struct {
		reqs []string
		e    bool
	}{
		{[]string{},
			true,
		},
		{[]string{"used1"}, true},
		{[]string{"used1", "not used"}, false},
	}

	for _, d := range td {
		seen := noDeps(d.reqs, used)
		expected := d.e
		if seen != expected {
			t.Errorf("Saw %s, expected %s", seen, expected)
		}
	}
}

func satisfiesDepMap(result []string, constraints map[string][]string) bool {
	check := make(map[string]bool)

	for _, name := range result {
		if check[name] {
			return false
		}
		check[name] = true

		for _, req := range constraints[name] {
			if !check[req] {
				return false
			}
		}
	}
	return true
}

func TestTopoSort(t *testing.T) {
	req1 := map[string][]string{
		"test3": {"test1", "test2"},
		"test2": {"test1"},
		"test1": {},
	}
	if satisfiesDepMap([]string{"test1", "test3", "test2"}, req1) {
		t.Errorf("satisfiesDepMap is teh bork.")
	}
	if !satisfiesDepMap([]string{"test1", "test2", "test3"}, req1) {
		t.Errorf("satisfiesDepMap is teh bork.")
	}

	seen1 := topoSort(req1)
	if !satisfiesDepMap(seen1, req1) {
		t.Errorf("toposort returned %s in error", seen1)
	}

	td := []map[string][]string {
		{
			"z": {"x", "y"},
			"x": {"a"},
			"y": {"w"},
			"w": {"b", "c"},
			"c": {"a", "b"},
			"b": {"a"},
			"a": {},
		},
	}

	for ix, d := range td {
		seen := topoSort(d)
		if !satisfiesDepMap(seen, d) {
			t.Errorf("Test %d, data %s is not correct", ix, seen)
		}
	}
}
