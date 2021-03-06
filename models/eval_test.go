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
	
func TestUploadBug(t *testing.T) {
	model := Model{}
	model.Inputs = map[string]Input{}
	model.NewInput("qps")
	model.SetInput("qps", "test", 300.0)
	model.Variables = make(map[string]variable)
	model.Resources = make(map[string]Expression)
	model.Variables["uploads_per_replica"] = newVariable("uploads_per_replica", constant{30.0})
	model.Variables["seconds_per_upload"] = newVariable("seconds_per_upload", constant{90.0})
	model.Variables["simultaneous_uploads"] = newVariable("simultaneous_uploads", operation{"*", reference{"qps"}, reference{"seconds_per_upload"}})
	model.Resources["replicas"] = operation{"/", reference{"simultaneous_uploads"}, reference{"uploads_per_replica"}}

	replicas := model.Resources["replicas"].Value(model)

	if replicas != 900.0 {
		t.Errorf("Saw %f replicas, expected 900", replicas)
	}
	
}

func TestRefInputs(t *testing.T) {
	model := Model{}
	model.Inputs = map[string]Input{}
	model.NewInput("in1")
	model.NewInput("in2")
	model.SetInput("in1", "test", 1.0)
	model.SetInput("in2", "test", 2.0)

	td := []struct {
		expr Expression
		expt float64
	}{
		{reference{"in1"}, 1.0},
		{reference{"in2"}, 2.0},
		{operation{"+", reference{"in1"}, reference{"in2"}}, 3.0},
	}

	for _, d := range td {
		seen := d.expr.Value(model)
		if seen != d.expt {
			t.Errorf("UNexpected input test value, saw %f, expected %f", seen, d.expt)
		}
	}
}

func TestRefVariables(t *testing.T) {
	model := Model{}
	model.Variables = make(map[string]variable)
	model.Variables["test1"] = variable{"test1", constant{2.0}, []float64{0.0}, []bool{false}}
	model.Variables["test2"] = variable{"test2", constant{4.2}, []float64{0.0}, []bool{false}}

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

	for name, val := range model.Variables {
		if !val.cached[0] {
			t.Errorf("Variable %s not cached", name)
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
