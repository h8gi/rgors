package lispy

import (
	"fmt"
)

type Frame map[string]LObj

type Environment struct {
	Frame
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

type Stack []LObj

type Code Stack

func (s *Stack) Push(obj LObj) {
	*s = append(*s, obj)
}

func (s *Stack) Pop() LObj {
	last := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return last
}

type Secd struct {
	Stack       Stack
	Environment Environment
	Code        Code // LObj pair
	Dump        interface{}
}

func (secd *Secd) Nil() {
	secd.Stack.Push(LObj{Type: Boolean, Value: false})
}

func (secd *Secd) Ldc(obj LObj) {
	secd.Stack.Push(obj)
}

func (secd *Secd) Ld(sym LObj) error {
	var val, err = secd.Environment.LookUp(sym)
	if err != nil {
		return err
	}
	secd.Stack.Push(val)
	return err
}

func (secd *Secd) Sel(cthen, celse Code) {
	var flag = secd.Stack.Pop()
	if flag.ToBool() {
	} else {

	}
}
