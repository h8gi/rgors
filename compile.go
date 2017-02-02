package rgors

func (next *LObj) isTail() bool {
	return next.Car.Eq(NewSymbol("return"))
}

// compile to continuation passing style
func (x *LObj) comp(next LObj) LObj {
	if x.IsSymbol() {
		return NewList(*NewSymbol("refer"), *x, next)
	}
	if x.IsPair() {
		switch x.Car.String() {
		case "quote": // (quote obj)
			obj, _ := x.ListRef(1)
			return NewList(*NewSymbol("constant"), obj, next)
		case "lambda": // (lambda (var ...) body)
			vars, _ := x.ListRef(1)
			body, _ := x.ListRef(2)
			return NewList(*NewSymbol("close"), vars,
				// at last, closure return
				body.comp(NewList(*NewSymbol("return"))), next)
		case "if": // (if test then else)
			test, _ := x.ListRef(1)
			then, _ := x.ListRef(2)
			els, _ := x.ListRef(3)
			thenc := then.comp(next)
			elsec := els.comp(next)
			return test.comp(NewList(*NewSymbol("test"), thenc, elsec))
		case "set!": // (set! var x)
			varsym, _ := x.ListRef(1)
			x, _ := x.ListRef(2)
			return x.comp(NewList(*NewSymbol("assign"), varsym, next))
		case "call/cc": // (call/cc x)
			x, _ := x.ListRef(1) // x should be proc
			c := NewList(*NewSymbol("argument"), x.comp(NewList(*NewSymbol("apply"))))
			c = NewList(*NewSymbol("conti"), c)
			// tail call
			if next.isTail() {
				return c
			} else {
				// save call frame (next is return address)
				return NewList(*NewSymbol("frame"), next, c)
			}
		default:
			// apply function
			args := x.Cdr
			// c's last inst is apply
			c := x.Car.comp(NewList(*NewSymbol("apply")))
			for {
				if args.IsNull() {
					if next.isTail() {
						return c
					} else {
						// next is return address
						return NewList(*NewSymbol("frame"), next, c)
					}
				}
				// cons up c with argument
				c = args.Car.comp(NewList(*NewSymbol("argument"), c))
				args = args.Cdr
			}
		}
	}
	return NewList(*NewSymbol("constant"), *x, next)
}

func (x *LObj) Compile() LObj {
	return x.comp(NewList(*NewSymbol("halt")))
}
