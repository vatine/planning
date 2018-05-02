// evaluator

package models

type Expression interface {
	Value(Model) float64
}

type variable struct {
	name   string
	expr   Expression
	cache  float64
	cached bool
}

type operation struct {
	operator string
	left Expression
	right Expression
}

type constant struct {
	value float64
}

type reference struct {
	name string
}


func newVariable(name string, expr Expression) variable {
	return variable{name, expr, 0.0, false}
}

func (c constant) Value(m Model) float64 {
	return c.value
}

func (r reference) Value(m Model) float64 {
	v, ok := m.Inputs[r.name]
	if ok {
		return v.Value(m)
	}
	v2, ok2 := m.Variables[r.name]
	if ok2 {
		return v2.Value(m)
	}
	return -100000.0
}

func (i Input) Value(m Model) float64 {
	acc := 0.0
	for _, iv := range i.values {
		acc += iv.value
	}
	return acc
}

func (v variable) Value(m Model) float64 {
	if !v.cached {
		v.cache = v.expr.Value(m)
		v.cached = true
	}
	return v.cache
}

func (v operation) Value(m Model) float64 {
	lv := v.left.Value(m)
	rv := v.right.Value(m)

	switch v.operator {
	case "+": return lv + rv
	case "-": return lv - rv
	case "*": return lv * rv
	case "/": return lv / rv
	}

	return -100000.0
}
