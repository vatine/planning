package models

import (
	"errors"
	"fmt"
)

const (
	number int = iota
	ref
	operator
	open
	closed
)

type token struct {
	t int
	repr string
}

func tokenize(s string) <-chan token {
	c := make(chan token)

	go tokenizeInner(s, c)
	return c
}

func tokenizeInner(s string, c chan<- token) {
	pos := 0

	for pos >= 0 {
		next, t := oneToken(s, pos)
		if next >= 0 {
			c <- t
		}
		pos = next
	}
	close(c)
}

func oneToken(s string, p int) (int, token) {
	for ;p < len(s) && s[p] == ' '; p++ {}
	if p >= len(s) {
		return -1, token{}
	}
	start := p
	switch {
	case s[start] >= '0' && s[start] <= '9' || s[start] == '-':
		return tokenNumber(s, start)
	case s[start] == '+':
		end := start + 1
		t := token{operator, "+"}
		return end, t
	case s[start] == '-':
		end := start + 1
		t := token{operator, "-"}
		return end, t
	case s[start] == '*':
		end := start + 1
		t := token{operator, "*"}
		return end, t
	case s[start] == '/':
		end := start + 1
		t := token{operator, "/"}
		return end, t
	case s[start] == '(':
		end := start + 1
		t := token{open, "("}
		return end, t
	case s[start] == ')':
		end := start + 1
		t := token{closed, ")"}
		return end, t
	}
	return tokenReference(s, start)
}

func tokenNumber(s string, start int) (int, token) {
	for end := start+1; end < len(s); end++ {
		switch {
		case s[end] >= '0' && s[end] <= '9':
			_ = true
		case s[end] == '.':
			_ = true
		default:
			return end, token{number, s[start:end]}
		}	
	}
	return len(s), token{number, s[start:len(s)]}
}

func tokenReference(s string, start int) (int, token) {
	for end := start; end < len(s); end++ {
		switch {
		case s[end] == ' ':
			return end, token{ref, s[start:end]}
		case s[end] == '*':
			return end, token{ref, s[start:end]}
		case s[end] == '/':
			return end, token{ref, s[start:end]}
		case s[end] == '+':
			return end, token{ref, s[start:end]}
		case s[end] == '-':
			return end, token{ref, s[start:end]}
		case s[end] == ')':
			return end, token{ref, s[start:end]}
		case s[end] == '(':
			return end, token{ref, s[start:end]}
		}
	}
	return len(s), token{ref, s[start:len(s)]}
}

func parse(s string) (Expression, error) {
	return parseInner(tokenize(s), 0)
}

func parseInner(c <-chan token, level int) (Expression, error){
	precedence := map[string]int{"+": 5, "-": 5, "*": 10, "/": 10}
	ops := []operation{}
	output := []Expression{}
	for t := range c {
		switch t.t {
		case number:
			output = append(output, parseNumber(t))
		case operator:
			for len(ops) > 0 && precedence[ops[len(ops) - 1].operator] > precedence[t.repr] {
				op := ops[len(ops) - 1]
				n := len(output)
				op.right = output[n - 1]
				op.left = output[n - 2]
				output = append(output[:n-2], op)
				ops = ops[:len(ops) - 1]
			}
			op := operation{operator: t.repr}
			ops = append(ops, op)
		case open:
			tmp, err := parseInner(c, level+1)
			if err != nil {
				return tmp, err
			}
			output = append(output, tmp)
		case closed:
			if level == 0 {
				// Make sure the queue is emptied...
				go func(){
					for _ = range c {}
				}()
				return constant{-1.0}, errors.New("Unexpected close parenthesis.")
			}
			for len(ops) > 0 {
				op := ops[len(ops) - 1]
				n := len(output)
				op.right = output[n - 1]
				op.left = output[n - 2]
				output = append(output[:n - 2], op)
				ops = ops[:len(ops) - 1]
			}
			return output[0], nil
		case ref:
			output = append(output, reference{t.repr})
		}
	}
	for len(ops) > 0 {
		op := ops[len(ops) - 1]
		n := len(output)
		op.right = output[n - 1]
		op.left = output[n - 2]
		output = append(output[:n - 2], op)
		ops = ops[:len(ops) - 1]
	}
	return output[0], nil
}

func parseNumber(t token) constant {
	var v float64
	cnt, err := fmt.Sscan(t.repr, &v)
	if cnt == 1 && err == nil {
		return constant{v}
	}
	return constant{-1.0}
}
