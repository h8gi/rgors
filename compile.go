package rgors

func (next *LObj) isTail() bool {
	return next.Car.Eq(NewSymbol("return"))
}

// compile to continuation passing style
func (x *LObj) comp(next LObj) (LObj, error) {
	if x.IsSymbol() {
		return NewList(*NewSymbol("refer"), *x, next), nil
	}
	if x.IsPair() {
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
			body, err = body.comp(NewList(*NewSymbol("return")))
			if err != nil {
				return body, err
			}
			return NewList(*NewSymbol("close"), vars, body, next), nil
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
			thenc, err := then.comp(next)
			if err != nil {
				return thenc, err
			}
			elsec, err := els.comp(next)
			if err != nil {
				return elsec, err
			}
			return test.comp(NewList(*NewSymbol("test"), thenc, elsec))
		case "set!": // (set! var x)
			varsym, err := x.ListRef(1)
			if err != nil {
				return varsym, err
			}
			x, err := x.ListRef(2)
			if err != nil {
				return x, err
			}
			return x.comp(NewList(*NewSymbol("assign"), varsym, next))
		case "call/cc": // (call/cc x)
			x, err := x.ListRef(1) // x should be proc
			if err != nil {
				return x, err
			}
			c, err := x.comp(NewList(*NewSymbol("apply")))
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
			c, err := x.Car.comp(NewList(*NewSymbol("apply")))
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
				c, err = args.Car.comp(NewList(*NewSymbol("argument"), c))
				if err != nil {
					return c, err
				}
				args = args.Cdr
			}
		}
	}
	return NewList(*NewSymbol("constant"), *x, next), nil
}

func (x *LObj) Compile() (LObj, error) {
	return x.comp(NewList(*NewSymbol("halt")))
}
