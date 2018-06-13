// Capacity models for planner

package models

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"

	yaml "gopkg.in/yaml.v2"
)

// "Internal" representation of a model
type Model struct {
	Name      string
	Inputs    map[string]Input
	Outputs   []Output
	Variables map[string]variable
	Resources map[string]Expression
}

type ExternalOutput struct {
	Backend    string
	Input      string
	Expression string
}
// Serialization representation of a model
type ExternalModel struct {
	Name    string
	Inputs  []string
	Outputs []ExternalOutput
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
	m.Resources = make(map[string]Expression)

	return &m
}

// Creates a new input on a model
func (m *Model) NewInput(name string) {
	i := Input{name: name}
	i.values = []inputValue{}
	m.Inputs[name] = i
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
		m.Inputs[iName] = input
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
func topoSort(deps map[string][]string) ([]string, error) {
	used := make(map[string]bool)
	rv := []string{}

	for len(rv) < len(deps) {
		possibles := []string{}
		for candidate, reqs := range deps {
			if noDeps(reqs, used) {
				possibles = append(possibles, candidate)
			}
		}
		if len(possibles) == 0 {
			return nil, errors.New("No available candidates in topoSort")
		}
		for _, possible := range possibles {
			if !used[possible] {
				rv = append(rv, possible)
				used[possible] = true
			}
		}
	}

	return rv, nil
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
func ModelOrder(models map[string]*Model) ([]*Model, error) {
	deps := modelsToDepMap(models)
	rv := []*Model{}
	sorted, err := topoSort(deps)
	if err != nil {
		return nil, err
	}
	for _, name := range sorted {
		rv = append(rv, models[name])
	}
	return rv, nil
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
			expr, err := Parse(expression)
			if err == nil {
				m.NewOutput(backend, input, expr)
			}
		}
	}

	for v, e := range e.Variables {
		expr, err := Parse(e)
		if err == nil {
			m.Variables[v] = newVariable(v, expr)
		}
	}

	for resource, expr := range e.Resources {
		parsed, err := Parse(expr)
		if err == nil {
			m.Resources[resource] = parsed
		}
	}

	return m
}

func BuildExternal(data []byte) (ExternalModel, error) {
	rv := ExternalModel{}
	err := yaml.Unmarshal(data, &rv)
	return rv, err
}

func LoadExternalModels(r io.Reader) ([]ExternalModel, error) {
	rv := []ExternalModel{}
	data, readErr := ioutil.ReadAll(r)
	if readErr != nil {
		return nil, readErr
	}
	err := yaml.Unmarshal(data, &rv)
	return rv, err
}

func Propagate(models map[string]*Model, topLevel string, inputs map[string]Expression) error {
	top, ok := models[topLevel]
	if !ok {
		return errors.New(fmt.Sprintf("Top-level model %s not found."))
	}

	for name, value := range inputs {
		top.SetInput(name, "external", value.Value(*top))
	}

	sorted, sortErr := ModelOrder(models)
	if sortErr != nil {
		return sortErr
	}
	for _, model := range sorted {
		for _, val := range model.Variables {
			_ = val.Value(*model)
		}
		model.PropagateOutputs()
	}
	
	return nil
}

func PrintModel( w io.Writer, m *Model) {
	fmt.Fprintf(w, "- name: %s\n", m.Name)
	fmt.Fprintf(w, "  resources:\n")
	if ram, rOK := m.Resources["ram"]; rOK {
		fmt.Fprintf(w, "    ram: %f # per replica\n", ram.Value(*m))
	}
	if cores, cOK := m.Resources["cpu"]; cOK {
		fmt.Fprintf(w, "    cpu: %f # per replica\n", cores.Value(*m))
	}
	replicas, repOK := m.Resources["replicas"]
	if !repOK {
		replicas = constant{1.0}
	}
	r := math.Ceil(replicas.Value(*m))
	fmt.Fprintf(w, "    replicas: %.0f\n", r)
}

func allRAM(m *Model) float64 {
	ram, ok := m.Resources["ram"]
	if ok {
		r, ok2 := m.Resources["replicas"]
		if !ok2 {
			r = constant{1.0}
		}
		replicas := math.Ceil(r.Value(*m))
		return ram.Value(*m) * replicas
	}
	return 0
}

func allCPU(m *Model) float64 {
	cpu, ok := m.Resources["cpu"]
	if ok {
		r, ok2 := m.Resources["replicas"]
		if !ok2 {
			r = constant{1.0}
		}
		replicas := math.Ceil(r.Value(*m))
		return cpu.Value(*m) * replicas
	}
	return 0
}

func PrintModels(w io.Writer, models map[string]*Model) {
	ram := 0.0
	cpu := 0.0
	for _, model := range models {
		ram += allRAM(model)
		cpu += allCPU(model)
		PrintModel(w, model)
	}
	fmt.Fprintf(w, "\ntotals:\n ram: %f\n cpu: %f\n", ram, cpu)
}
