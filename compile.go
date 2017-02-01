package rgors

func (x *LObj) Compile(next LObj) LObj {
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
				body.Compile(NewList(NewSymbol("return"))), next)
		case "if":

		}
	}
}
