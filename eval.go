package lispy

import (
	"fmt"
)

type Environment struct {
	Frame     map[string]LObj
	Enclosing *Environment
}

func (env *Environment) LookUp(sym LObj) (LObj, error) {
	obj, ok := env.Frame[sym.String()]
	if ok {
		return obj, nil
	}
	nextenv := env.Enclosing
	if nextenv == nil {
		return sym, fmt.Errorf("unbound variable: %v", sym)
	}
	return nextenv.LookUp(sym)
}

func EvalProgram(program []LObj) (LObj, error) {
	var err error
	var result LObj
	env := Environment{Frame: map[string]LObj{"a": LObj{Type: LispNumber, Value: 12}}}
	for _, expr := range program {
		result, err = Eval(expr, env)
		if err != nil {
			break
		}
	}
	return result, err
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
	default: //  self evaluating
		return expr, err
	}
}

func EvalPair(pair LObj, env Environment) (LObj, error) {
	var err error
	var fun LObj
	switch pair.Car.Type {
	case LispBoolean, LispChar, LispVector, LispNumber, LispString, LispPort, LispNil:
		return *pair.Car, fmt.Errorf("call of non-procedure: %v", pair.Car)
	case LispProcedure:
		return Apply(*pair.Car, *pair.Cdr)
	default:
		fun, err = Eval(*pair.Car, env)
		return Apply(fun, *pair.Cdr)
	}
	return pair, err
}
func EvalSymbol(sym LObj, env Environment) (LObj, error) {
	return env.LookUp(sym)
}

// List functions
func ListRef(pair LObj, n int) {

}

func Apply(fun LObj, args LObj) (LObj, error) {
	var result LObj
	var err error
	return result, err
}
