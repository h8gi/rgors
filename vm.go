package rgors

import (
	"fmt"
)

type VM struct {
	a LObj // the accumulator
	x LObj // the next expression (list)
	e LObj // the current environment
	r LObj // the current value rib
	s LObj // the current stack
}

func NewVM() *VM {
	vm := &VM{
		a: LispNull,
		x: LispNull,
		e: LispNull,
		r: LispNull,
		s: LispNull,
	}
	// vars := NewList(*NewSymbol("+"), *NewSymbol("-"))
	// vals := NewList(
	// 	LObj{
	// 		Type: DTPrimitive,
	// 		Value: func(args ...LObj) LObj {
	// 			ret := 0
	// 			for _, elem := range args {
	// 				ret += elem.Value.(int)
	// 			}
	// 			return LObj{
	// 				Type:  DTNumber,
	// 				Value: ret,
	// 			}
	// 		}},
	// 	LObj{
	// 		Type: DTPrimitive,
	// 		Value: func(args ...LObj) LObj {
	// 			var ret int = 0
	// 			if len(args) == 1 {
	// 				ret = -args[0].Value.(int)
	// 			} else {
	// 				ret = args[0].Value.(int)
	// 				args = args[1:len(args)]
	// 				for _, elem := range args {
	// 					ret -= elem.Value.(int)
	// 				}
	// 			}
	// 			return LObj{
	// 				Type:  DTNumber,
	// 				Value: ret,
	// 			}
	// 		},
	// 	},
	// )
	// vm.e = vm.e.Extend(vars, vals)
	return vm
}

func (vm *VM) Load(obj LObj) {
	vm.x = obj
}

func (vm VM) String() string {
	return fmt.Sprintf("a: %v\nx: %v\ne: %v\nr: %v\ns: %v\n", vm.a, vm.x, vm.e, vm.r, vm.s)
}

func (vm *VM) Run() (LObj, error) {
	// TODO: errorcheck
Loop:
	for {
		fmt.Println(vm)
		switch vm.x.Car.String() {
		case "halt": // (halt)
			// finish computation, return value
			break Loop
		case "refer": // (refer var next-x)
			// get variable name
			varsym, _ := vm.x.ListRef(1)
			// set x to next-x
			vm.x, _ = vm.x.ListRef(2)
			// set accumulator to variable's value
			vals, err := vm.e.LookUp(&varsym)
			if err != nil {
				return *vals, err
			}
			vm.a = *vals.Car
		case "constant": // (constant obj next-x)
			//  set! accumulator constant value
			vm.a, _ = vm.x.ListRef(1)
			// set x to next
			vm.x, _ = vm.x.ListRef(2)
		case "close": // (close vars body next-x)
			// get lambda variables
			vars, _ := vm.x.ListRef(1)
			// get lambda body
			body, _ := vm.x.ListRef(2)
			// set x to next-x
			vm.x, _ = vm.x.ListRef(3)
			// set accumulator to closure
			vm.a = NewClosure(vars, body, vm.e)
		case "test": // (test then else)
			thenobj, _ := vm.x.ListRef(1)
			elseobj, _ := vm.x.ListRef(2)
			// if accumulator is true
			if vm.a.ToBool() {
				vm.x = thenobj
			} else {
				vm.x = elseobj
			}
		case "assign": // (assign var next-x)
			varsym, _ := vm.x.ListRef(1)
			// set x to next-x
			vm.x, _ = vm.x.ListRef(2)
			vals, err := vm.e.LookUp(&varsym)
			if err != nil {
				return *vals, err
			}
			// assing var to value
			vals.SetCar(vm.a)
		case "conti": // (conti x)
			// later, x takes one argument from accumulater
			vm.x, _ = vm.x.ListRef(1)
			// make continuation from stack
			vm.a = NewContinuation(vm.s)
		case "naute": // (naute s var)
			// restore s
			vm.s, _ = vm.x.ListRef(1)
			// set accumulator to var's value
			varsym, _ := vm.x.ListRef(2)
			vals, _ := vm.e.LookUp(&varsym)
			vm.a = *vals.Car
			// next is (return)
			vm.x = NewList(*NewSymbol("return"))
		case "frame": // (frame ret next-x)
			ret, _ := vm.x.ListRef(1)
			// set x to next-x
			vm.x, _ = vm.x.ListRef(2)
			vm.s = NewCallFrame(ret, vm.e, vm.r, vm.s)
			vm.r = LispNull
		case "argument": // (argument x)
			vm.x, _ = vm.x.ListRef(1)
			vm.r = Cons(vm.a, vm.r)
		case "apply": // (apply)
			// accumulator is closure or primitive
			if vm.a.IsClosure() {
				body := vm.a.Body()
				e := vm.a.Env()
				vars := vm.a.Vars()
				// next inst is body
				vm.x = body // body's cont is (return)
				// extend env with arguments
				vm.e = e.Extend(vars, vm.r)
				vm.r = LispNull
			} else if vm.a.IsPrimitive() {
				vm.a = vm.a.PrimitiveApply(vm.r)
				vm.r = LispNull
				vm.x = NewList(*NewSymbol("return"))
			} else {
				return LispFalse, fmt.Errorf("not procedure: %v", vm.a)
			}
		case "return":
			// resets from stack
			vm.x, _ = vm.s.ListRef(0)
			vm.e, _ = vm.s.ListRef(1)
			vm.r, _ = vm.s.ListRef(2)
			vm.s, _ = vm.s.ListRef(3)
		}
	}
	ret := vm.a
	vm.a = LispNull
	return ret, nil
}

// VM support functions
//
// environment
func (env *LObj) LookUp(sym *LObj) (*LObj, error) {
	for {
		// env exhausted
		if env.IsNull() {
			return &LispFalse, fmt.Errorf("unbound variable: %v", sym)
		}
		vars := env.Car.Car
		vals := env.Car.Cdr
		for {
			// goto next rib
			if vars.IsNull() {
				break
			}
			// found!
			if vars.Car.Eq(sym) {
				return vals, nil
			}
			// next
			vars = vars.Cdr
			vals = vals.Cdr
		}
		env = env.Cdr
	}
}

func (env *LObj) Extend(vars, vals LObj) LObj {
	return Cons(Cons(vars, vals), *env)
}

// closure
func NewClosure(vars, body, env LObj) LObj {
	return LObj{
		Type:  DTClosure,
		Car:   &vars,
		Cdr:   &body,
		Value: env,
	}
	// return NewList(body, env, vars)
}
func (closure *LObj) Vars() LObj {
	return *closure.Car
}
func (closure *LObj) Body() LObj {
	return *closure.Cdr
}
func (closure *LObj) Env() LObj {
	return closure.Value.(LObj)
}

// continuation
func NewContinuation(s LObj) LObj {
	symv := NewSymbol("v")
	vars := NewList(*symv)
	body := NewList(*NewSymbol("naute"), s, *symv)
	env := LispNull
	return NewClosure(vars, body, env)
}

// call frame
func NewCallFrame(x, e, r, s LObj) LObj {
	return NewList(x, e, r, s)
}

func (obj *LObj) PrimitiveApply(arglist LObj) LObj {
	args := make([]LObj, 0)
	for {
		if arglist.IsNull() {
			break
		}
		elem, _ := arglist.Pop()
		args = append(args, elem)
	}
	f := obj.Value.(func(...LObj) LObj)
	return f(args...)
}
