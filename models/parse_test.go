package models

import (
	"testing"
)

func tokenEqual(a, b token) bool {
	return a.t == b.t && a.repr == b.repr
}

func compareExpr(a, b Expression) bool {
	switch a.(type) {
	case variable:
		switch b.(type) {
		case variable:
			v1 := a.(variable)
			v2 := b.(variable)
			if v1.name == v2.name {
				return compareExpr(v1.expr, v2.expr)
			}
		}
	case constant:
		switch b.(type) {
		case constant: return a.(constant).value == b.(constant).value
		}
	case operation:
		switch b.(type) {
		case operation:
			o1 := a.(operation)
			o2 := b.(operation)
			if o1.operator == o2.operator {
				return compareExpr(o1.left, o2.left) && compareExpr(o1.right, o2.right)
			}
		}
	case reference:
		switch b.(type) {
		case reference:
			return a.(reference).name == b.(reference).name
		}
	}
	return false
}

func TestTokenizer(t *testing.T) {
	td := []struct{
		s string
		tokens []token
	}{
		{"12", []token{token{number, "12"}}},
		{" 12", []token{token{number, "12"}}},
		{"heynonnynonny", []token{token{ref, "heynonnynonny"}}},
		{"1+2", []token{token{number, "1"}, token{operator, "+"}, token{number, "2"}}},
		{" 1     + 2       ", []token{token{number, "1"}, token{operator, "+"}, token{number, "2"}}},
		{" 1 * ()    + 2       ", []token{
			token{number, "1"}, token{operator, "*"},
			token{open, "("}, token{closed, ")"},
			token{operator, "+"}, token{number, "2"}}},
	}

	for ix, d := range td {
		c := tokenize(d.s) 
		for tIx, expected := range d.tokens {
			seen := <- c
			if !tokenEqual(seen, expected) {
				t.Errorf("test %d, pos %d, Expected %t, saw %t", ix, tIx, seen, expected)
			}
		}
	}
}

func TestParser(t *testing.T) {
	td := []struct{
		s    string
		err  bool
		expr Expression
	}{
		{"1+2", false, operation{operator: "+", left: constant{1.0}, right: constant{2.0}}},
		{"1+2*3", false,
			operation{operator: "+",
				left: constant{1.0},
				right: operation{
					operator: "*",
					left: constant{2.0},
					right: constant{3.0}}}},
		{"1*2+3", false,
			operation{operator: "+",
				left: operation{
					operator: "*",
					left: constant{1.0},
					right: constant{2.0}},
				right: constant{3.0}}},
		{"a*2 + 7*b", false,
			operation{operator: "+",
				left: operation{
					operator: "*",
					left: reference{"a"},
					right: constant{2}},
				right: operation{
					operator: "*",
					left: constant{7},
					right: reference{"b"}}}},
		{"(a+b)*3", false,
			operation{
				operator: "*",
				left: operation{
					operator: "+",
					left: reference{"a"},
					right: reference{"b"}},
				right: constant{3}}},
		{"1+2)-3", true, constant{1}},
	}

	for ix, d := range td {
		seen, e := Parse(d.s)
		if e != nil {
			if !d.err {
				t.Errorf("test %d, did not expect error, saw %s", ix, e)
			}
		} else {
			if d.err {
				t.Errorf("test %d, did not see expected error", ix)
			} else {
				if !compareExpr(seen, d.expr) {
					t.Errorf("test %d, expression fail\n%t\n%t", ix, seen, d.expr)
				}
			}
		}
	}
}
