package lispy

import (
	"fmt"
)

type Environment struct {
	Frame     map[string]LObj
	Enclosing *Environment
}

func EvalProgram(program []LObj) (LObj, error) {
	var err error
	env := Environment{Frame: map[string]LObj{"a": LObj{Type: LispNumber, Value: 12}}}
	for i, expr := range program {
		program[i], err = Eval(expr, env)
	}
	return program[len(program)-1], err
}

func Eval(expr LObj, env Environment) (LObj, error) {
	var err error
	switch expr.Type {
	case LispPair:
		return EvalPair(expr, env)
	case LispSymbol:
		return EvalSymbol(expr, env)
	case LispProcedure:
		return LObj{}, fmt.Errorf("illegal non-atomic object: %v", expr)
	default:
		return expr, err
	}
}

func EvalPair(expr LObj, env Environment) (LObj, error) {
	var err error
	return expr, err
}
func EvalSymbol(expr LObj, env Environment) (LObj, error) {
	var err error
	return env.Frame[expr.Value.(string)], err
}
