package lispy

type Environment struct {
	Frame     map[string]LObj
	Enclosing *Environment
}

func EvalProgram(program []LObj, env Environment) (LObj, Environment) {
	for i, expr := range program {
		program[i], env = Eval(expr, env)
	}
	return program[len(program)-1], env
}

func Eval(expr LObj, env Environment) (LObj, Environment) {
	switch expr.Type {
	case LispPair:
		return EvalPair(expr, env)
	case LispSymbol:
		return EvalSymbol(expr, env)
	default:
		return expr, env
	}
}

func EvalPair(expr LObj, env Environment) (LObj, Environment) {
	return expr, env
}
func EvalSymbol(expr LObj, env Environment) (LObj, Environment) {
	return expr, env
}
