// Capacity models for planner

package models

import (
	"fmt"
)

type Model struct {
	Name string
	Inputs map[string]Input
	Outputs []Output
	Variables map[string]variable
}

type Input struct {
	name string
	values []inputValue
}

type inputValue struct {
	source string
	value float64
}

type Output struct {
	backend string
	input   string
	value   Expression
}

var models map[string]*Model

// Creates a new model with the given name
func New(name string) *Model {
	if spec, ok := models[name]; ok {
		return spec
	}
		
	m := Model{Name: name}
	m.Inputs = make(map[string]Input)
	m.Outputs = []Output{}
	m.Variables = make(map[string]variable)

	return &m
}

func (m *Model) NewInput(name string) {
	i := Input{name: name}
	i.values = []inputValue{}
}

func (m *Model) NewOutput(destination, sink string, value Expression) {
	o := Output{backend: destination, input: sink, value: value}
	m.Outputs = append(m.Outputs, o)
}

func (m *Model) SetInput(iName,from string, v float64) {
	input, ok := m.Inputs[iName]
	if ok {
		iv := inputValue{source: from, value: v}
		input.values = append(input.values, iv)
	}
}

func (m *Model) PropagateOutputs() {
	for _, o := range m.Outputs {
		dst, ok := models[o.backend]
		if ok {
			dst.SetInput(o.input, m.Name, o.value.Value(*m))
		} else {
			fmt.Printf("<model %s> No model named %s\n", m.Name, o.backend)
		}
	}
}


func noDeps(reqs []string, used map[string]bool) bool {
	for _, name := range reqs {
		if !used[name] {
			return false
		}
	}
	return true
}

func topoSort(deps map[string][]string) []string {
	used := make(map[string]bool)
	rv := []string{}

	for len(rv) < len(deps) {
		possibles := []string{}
		for candidate, reqs := range deps {
			if noDeps(reqs, used) {
				possibles = append(possibles, candidate)
			}
		}
		for _, possible := range possibles {
			if !used[possible] {
				rv = append(rv, possible)
				used[possible] = true
			}
		}
	}

	return rv
}
