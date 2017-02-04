package rgors

import (
	"fmt"
)

// check tail call
func (next *LObj) isTail() bool {
	return next.Car.Eq(NewSymbol("return"))
}

func (e *LObj) Extend(r LObj) LObj {
	return Cons(r, *e)
}

func (env *LObj) CompileLookUp(varsym *LObj) (pair LObj, err error) {
	var rib, elt int = 0, 0
	for {
		if env.IsNull() {
			return LispFalse, fmt.Errorf("unbound variable: %v", varsym)
		}
		vars := env.Car
		if !vars.IsPair() {
			return *vars, fmt.Errorf("lambda vars not pair: %v", vars)
		}
		for {
			// goto next rib
			if vars.IsNull() {
				rib += 1
				elt = 0
				break
			}
			// found!
			if vars.Car.Eq(varsym) {
				return Cons(
					LObj{Type: DTNumber, Value: rib},
					LObj{Type: DTNumber, Value: elt}), nil
			}
			// next
			vars = vars.Cdr
			elt += 1
		}
		// not found in this rib
		env = env.Cdr
	}
}

// compile to continuation passing style
func (x *LObj) comp(next, env LObj) (LObj, error) {
	if x.IsSymbol() { // symbol
		pair, err := env.CompileLookUp(x)
		if err != nil {
			return pair, err
		}
		return NewList(*NewSymbol("refer"), pair, next), nil
	} else if x.IsPair() { // pair
		switch x.Car.String() {
		case "quote": // (quote obj)
			obj, err := x.ListRef(1)
			if err != nil {
				return obj, err
			}
			return NewList(*NewSymbol("constant"), obj, next), nil
		case "lambda": // (lambda (var ...) body)
			vars, err := x.ListRef(1)
			if err != nil {
				return vars, err
			}
			body, err := x.ListRef(2)
			if err != nil {
				return body, err
			}
			body, err = body.comp(NewList(*NewSymbol("return")), env.Extend(vars))
			if err != nil {
				return body, err
			}
			return NewList(*NewSymbol("close"), body, next), nil
		case "if": // (if test then else)
			test, err := x.ListRef(1)
			if err != nil {
				return test, err
			}
			then, err := x.ListRef(2)
			if err != nil {
				return test, err
			}
			els, err := x.ListRef(3)
			if err != nil {
				return els, err
			}
			thenc, err := then.comp(next, env)
			if err != nil {
				return thenc, err
			}
			elsec, err := els.comp(next, env)
			if err != nil {
				return elsec, err
			}
			return test.comp(NewList(*NewSymbol("test"), thenc, elsec), env)
		case "set!": // (set! var x)
			varsym, err := x.ListRef(1)
			if err != nil {
				return varsym, err
			}
			x, err := x.ListRef(2)
			if err != nil {
				return x, err
			}
			access, err := env.CompileLookUp(&varsym)
			if err != nil { // not found
				return access, err
			}
			return x.comp(NewList(*NewSymbol("assign"), access, next), env)
		case "call/cc": // (call/cc x)
			x, err := x.ListRef(1) // x should be proc
			if err != nil {
				return x, err
			}
			c, err := x.comp(NewList(*NewSymbol("apply")), env)
			if err != nil {
				return c, err
			}
			// (conti (argument c))
			c = NewList(*NewSymbol("conti"), NewList(*NewSymbol("argument"), c))
			// tail call
			if next.isTail() {
				return c, nil
			} else {
				// save call frame (next is return address)
				return NewList(*NewSymbol("frame"), next, c), nil
			}
		default:
			// apply function
			args := x.Cdr
			// c's last inst is apply
			c, err := x.Car.comp(NewList(*NewSymbol("apply")), env)
			if err != nil {
				return c, err
			}
			for {
				if args.IsNull() {
					if next.isTail() {
						return c, nil
					} else {
						// next is return address
						return NewList(*NewSymbol("frame"), next, c), nil
					}
				}
				// cons up c with argument
				c, err = args.Car.comp(NewList(*NewSymbol("argument"), c), env)
				if err != nil {
					return c, err
				}
				args = args.Cdr
			}
		}
	} else if x.IsSelfEvaluating() {
		return NewList(*NewSymbol("constant"), *x, next), nil
	} else {
		return LispFalse, fmt.Errorf("not atomic: %v", x)
	}

}

func (x *LObj) Compile() (LObj, error) {
	return x.comp(NewList(*NewSymbol("halt")), LispNull)
}
