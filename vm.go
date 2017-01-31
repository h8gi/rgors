package rgors

type VM struct {
	a *LObj // the accumulator
	x *LObj // the next expression
	e *LObj // the current environment
	r *LObj // the current value rib
	s *LObj // the current stack
}

func (vm *VM) Run() LObj {

}
