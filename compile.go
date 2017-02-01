package rgors

func (next *LObj) isTail() bool {
	return next.Car.Eq(NewSymbol("return"))
}

func (x *LObj) comp(next LObj) LObj {
	if x.IsSymbol() {
		return NewList(*NewSymbol("refer"), *x, next)
	}
	if x.IsPair() {
		switch x.Car.String() {
		case "quote":
			obj, _ := x.ListRef(1)
			return NewList(*NewSymbol("constant"), obj, next)
		case "lambda":
			vars, _ := x.ListRef(1)
			body, _ := x.ListRef(2)
			return NewList(*NewSymbol("close"), vars,
				body.comp(NewList(*NewSymbol("return"))), next)
		case "if":
			test, _ := x.ListRef(1)
			then, _ := x.ListRef(2)
			els, _ := x.ListRef(3)
			thenc := then.comp(next)
			elsec := els.comp(next)
			return test.comp(NewList(*NewSymbol("test"), thenc, elsec))
		case "set!":
			varsym, _ := x.ListRef(1)
			x, _ := x.ListRef(2)
			return x.comp(NewList(*NewSymbol("assign"), varsym, next))
		case "call/cc":
			x, _ := x.ListRef(1)
			c := NewList(*NewSymbol("argument"), x.comp(NewList(*NewSymbol("apply"))))
			c = NewList(*NewSymbol("conti"), c)
			if next.isTail() {
				return c
			} else {
				return NewList(*NewSymbol("frame"), next, c)
			}
		default:
			args := x.Cdr
			c := x.Car.comp(NewList(*NewSymbol("apply")))
			for {
				if args.IsNull() {
					if next.isTail() {
						return c
					} else {
						return NewList(*NewSymbol("frame"), next, c)
					}
				}
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
