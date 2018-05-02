package models

import (
	"testing"
)

func TestConstants(t *testing.T) {
	td := []struct {
		c constant
		e float64
	}{
		{constant{1.0}, 1.0},
		{constant{14.3}, 14.3},
	}
	m := Model{}
	for _, d := range td {
		seen := d.c.Value(m)
		expected := d.e
		if seen != expected {
			t.Errorf("Saw %f, expected %f", seen, expected)
		}
	}
}
	
func TestRefVariables(t *testing.T) {
	model := Model{}
	model.Variables = make(map[string]variable)
	model.Variables["test1"] = variable{name: "test1", expr: constant{2.0}}
	model.Variables["test2"] = variable{name: "test2", expr: constant{4.2}}

	td := []struct{
		r reference
		e float64
	}{
		{reference{"test1"}, 2.0},
		{reference{"test2"}, 4.2},
	}

	for _, d := range td {
		seen := d.r.Value(model)
		expected := d.e
		if seen != expected {
			t.Errorf("Unexpected referenced value, saw %f, expected %f", seen, expected)
		}
	}
}

func TestOperation(t *testing.T) {
	op := operation{"", constant{3.0}, constant{2.0}}
	td := []struct{
		op string
		e  float64
	}{
		{"+", 5.0},
		{"-", 1.0},
		{"*", 6.0},
		{"/", 1.5},
	}
	m := Model{}
	
	for _, d := range td {
		op.operator = d.op
		expected := d.e
		seen := op.Value(m)

		if seen != expected {
			t.Errorf("operator %s, saw %f, expected %f", d.op, seen, expected)
		}
	}
}
