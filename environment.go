package rgors

import (
	"fmt"
)

// env, closure
func NewEnv() *LObj {
	env := Cons(LispNull, LispNull)
	return &env
}

// about environment (list of alist)
func (env *LObj) LookUp(sym *LObj) (LObj, error) {
	if env.IsNull() {
		return LispFalse, fmt.Errorf("unbound variable: %v", sym)
	}
	currentEnv, err := env.SafeCar()
	if err != nil {
		return *env, err
	}
	// lookup current environment
	pair, err := currentEnv.Assq(*sym)
	if err != nil {
		return currentEnv, err
	}
	// found!
	if pair.ToBool() {
		return *pair.Cdr, nil
	}
	// not found
	return env.Cdr.LookUp(sym)
}

// destructive
func (env *LObj) Define(sym *LObj, val *LObj) {
	pair := Cons(*sym, *val)
	env.Car.Push(&pair)
}

// return new extended env
func (parent *LObj) Extend(child *LObj) *LObj {
	result := Cons(*child, *parent)
	return &result
}

func InitialEnv() *LObj {
	env := NewEnv()
	var lispAdd2 = LObj{
		Type: DTPrimitive,
		Value: func(obj1, obj2 LObj) (LObj, error) {
			if obj1.IsNumber() && obj2.IsNumber() {
				return LObj{Type: DTNumber, Value: obj1.Value.(int) + obj2.Value.(int)}, nil
			} else {
				return LispFalse, fmt.Errorf("+: not number: %v, %v", obj1, obj2)
			}
		},
	}
	sym := NewSymbol("+")
	env.Define(sym, &lispAdd2)

	var lispSub2 = LObj{
		Type: DTPrimitive,
		Value: func(obj1, obj2 LObj) (LObj, error) {
			if obj1.IsNumber() && obj2.IsNumber() {
				return LObj{Type: DTNumber, Value: obj1.Value.(int) - obj2.Value.(int)}, nil
			} else {
				return LispFalse, fmt.Errorf("-: not number: %v, %v", obj1, obj2)
			}
		},
	}
	sym = NewSymbol("-")
	env.Define(sym, &lispSub2)

	return env
}

// closure
func NewClosure(code *LObj, env *LObj) *LObj {
	return &LObj{
		Type: DTClosure,
		Car:  code,
		Cdr:  env,
	}
}

func (closure *LObj) Code() *LObj {
	return closure.Car
}
func (closure *LObj) Env() *LObj {
	return closure.Cdr
}
