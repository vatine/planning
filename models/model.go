// Capacity models for planner

package models

import (
	"fmt"

	// "gopkg.in/yaml.v2"
)

// "Internal" representation of a model
type Model struct {
	Name      string
	Inputs    map[string]Input
	Outputs   []Output
	Variables map[string]variable
}

// Serialization representation of a model
type ExternalModel struct {
	Name    string
	Inputs  []string
	Outputs []struct {
		Backend    string
		Input      string
		Expression string
	}
	Variables map[string]string
	Resources map[string]string
}

// Model inputs
type Input struct {
	name string
	values []inputValue
}

// Individual input, to allow for (future) tracking of contributions
type inputValue struct {
	source string
	value float64
}

// Output representation
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

// Creates a new input on a model
func (m *Model) NewInput(name string) {
	i := Input{name: name}
	i.values = []inputValue{}
}

// Creates a new output on a model
func (m *Model) NewOutput(destination, sink string, value Expression) {
	o := Output{backend: destination, input: sink, value: value}
	m.Outputs = append(m.Outputs, o)
}

// Sets a specific input on a model to a specific value
func (m *Model) SetInput(iName, from string, v float64) {
	input, ok := m.Inputs[iName]
	if ok {
		iv := inputValue{source: from, value: v}
		input.values = append(input.values, iv)
	}
}

// Ensures that all declared outputs of a model are properly fed to
// the models that should have the data.
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

// Utility function to check if all "feeds that to this" have been fulfilled
func noDeps(reqs []string, used map[string]bool) bool {
	for _, name := range reqs {
		if !used[name] {
			return false
		}
	}
	return true
}

// Topologically sorts a dependency tree. Expects a map keyed by model
// name, with a list of model names that a specific module depends on.
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

// Extracts a dependency map from a map of models.
func modelsToDepMap(models map[string]*Model) map[string][]string {
	rv := make(map[string][]string)
	for name, m := range models {
		for _, o := range m.Outputs {
			rv[o.backend] = append(rv[o.backend], name)
		}
	}

	return rv
}

// Returns a slice of models, in the order they need to be evaluated
// to propagate values properly.
func ModelOrder(models map[string]*Model) []*Model {
	deps := modelsToDepMap(models)
	rv := []*Model{}
	for _, name := range topoSort(deps) {
		rv = append(rv, models[name])
	}
	return rv
}

func ModelFromExternal(e ExternalModel) *Model {
	m := New(e.Name)
	for _, input := range e.Inputs {
		m.NewInput(input)
	}
	for _, output := range e.Outputs {
		backend := output.Backend
		input := output.Input
		expression := output.Expression

		if backend != "" && input != "" && expression != "" {
			expr, err := parse(expression)
			if err == nil {
				m.NewOutput(backend, input, expr)
			}
		}
	}

	for v, e := range e.Variables {
		expr, err := parse(e)
		if err == nil {
			m.Variables[v] = newVariable(v, expr)
		}
	}

	return m
}
