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
			t.Errorf("Saw %v, expected %v", seen, expected)
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

func cmpDepMap(a, b map[string][]string) bool {
	for name, aDeps := range a {
		if bDeps, ok := b[name]; ok {
			trace := map[string]int{}
			for _, r := range aDeps {
				trace[r] = 1
			}
			for _, r := range bDeps {
				trace[r] += 1
			}
			for _, val := range trace {
				if val != 2 {
					return false
				}
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

	seen1, _ := topoSort(req1)
	if !satisfiesDepMap(seen1, req1) {
		t.Errorf("toposort returned %s in error", seen1)
	}

	// TODO: add more test cases...
	td := []struct {
		m map[string][]string
		e bool
	}{
		{
			map[string][]string{
				"z": {"x", "y"},
				"x": {"a"},
				"y": {"w"},
				"w": {"b", "c"},
				"c": {"a", "b"},
				"b": {"a"},
				"a": {},
			},
			false,
		},
		{
			map[string][]string{"a": {"b"}, "b": {"a"}},
			true,
		},
	}

	for ix, d := range td {
		seen, err := topoSort(d.m)
		if (err != nil) !=  d.e{
			t.Errorf("Incorrect error, expected it to be %v, it was %v", d.e, !d.e)
		}
		if !satisfiesDepMap(seen, d.m) {
			t.Errorf("Test %d, data %s is not correct", ix, seen)
		}
	}
}

func TestModelsToDepMap(t *testing.T) {
	m1 := Model{
		Name: "TestTop",
		Outputs: []Output{{"test2", "blah", constant{1}}, {"test3", "blah", constant{1}}},
	}
	m2 := Model{
		Name: "test2",
		Outputs: []Output{{"test3", "blah", constant{1}}},
	}
	m3 := Model{
		Name: "test3",
	}

	models := map[string]*Model{
		"TestTop": &m1,
		"test2": &m2,
		"test3": &m3,
	}

	expected := map[string][]string{
		"test3": {"TestTop", "test3"},
		"test2": {"TestTop"},
		"TestTop": {},
	}

	seen := modelsToDepMap(models)

	if cmpDepMap(seen, expected) {
		t.Errorf("derp, %s is not %s", seen, expected)
	}
}

func cmpModels(a, b *Model) string {
	if a.Name != b.Name {
		return "names differ"
	}
	inSeen := map[string]int{}
	for aName, _ := range a.Inputs {
		inSeen[aName] = 1
	}
	for bName, _ := range b.Inputs {
		inSeen[bName] += 1
	}
	for name, val := range inSeen {
		if val != 2 {
			return "incommensurate input names"
		}
		aIn := a.Inputs[name]
		bIn := b.Inputs[name]
		if aIn.name != bIn.name || aIn.name != name {
			return "input has different names"
		}
	}
	for ix, aOut := range a.Outputs {
		bOut := b.Outputs[ix]
		if aOut.backend != bOut.backend {
			return "non-matching output names"
		}
		if aOut.input != bOut.input {
			return "non-matching inputs for output"
		}
		if !compareExpr(aOut.value, bOut.value) {
			return "outputs have different expressions"
		}
	}
	varSeen := map[string]int{}
	for name, _ := range a.Variables {
		varSeen[name] = 1
	}
	for name, _ := range b.Variables {
		varSeen[name] += 1
	}
	for name, count := range varSeen {
		if count != 2 {
			return "incommensurate variables"
		}
		aVar := a.Variables[name]
		bVar := b.Variables[name]
		if aVar.name != name || aVar.name != bVar.name {
			return "variable with incmopatible name"
		}
		if !compareExpr(aVar.expr, bVar.expr) {
			return "variable have different expressions"
		}
	}

	for _, name := range []string{"ram", "cpu", "replicas"} {
		aVal, aOK := a.Resources[name]
		bVal, bOK := b.Resources[name]

		if aOK != bOK {
			return "incommensurate resources"
		}
		if !compareExpr(aVal, bVal) {
			return "resource expression differ"
		}
	}
	
	return ""
}

func TestModelFromExternal(t *testing.T) {
	ext := ExternalModel{
		Name: "test",
		Inputs: []string{"qps"},
		Outputs: []ExternalOutput{
			{"bloop", "qps", "3"},
		},
		Variables: map[string]string{"foo": "3*qps+4"},
		Resources: map[string]string{
			"ram": "10",
			"cpu": "0.2",
			"replicas": "qps/200",
		},
	}
	expected := Model{
		Name: "test",
		Inputs: map[string]Input{"qps": Input{name: "qps"},},
		Outputs: []Output{{"bloop", "qps", constant{3}}},
		Variables: map[string]variable{
			"foo": variable{
				name: "foo",
				expr: operation{"+", operation{"*", constant{3}, reference{"qps"}}, constant{4}}}},
		Resources: map[string]Expression{
			"ram": constant{10},
			"cpu": constant{0.2},
			"replicas": operation{"/", reference{"qps"}, constant{200}},
		},
	}

	seen := ModelFromExternal(ext)

	if status := cmpModels(seen, &expected); status != "" {
		t.Errorf("Model conversion failed, %s.\n%v\n%v", status, seen, &expected)
	}
}

func TestUnmarshal(t *testing.T) {
	input := `name: test
inputs:
 - qps
 - foons
outputs:
 - backend: bob
   input: qps
   expression: 2*qps
 - backend: alice
   input: foons
   expression: foons/2
variables:
  combined: foons+qps
resources:
  ram: 64
  cpu: 1.3
  replicas: combined / 3
`
	_, err := BuildExternal([]byte(input))
	if err != nil {
		t.Errorf("Unmarshal saw an error, %s", err)
	}
}
