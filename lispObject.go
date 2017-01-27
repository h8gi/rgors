package lispy

import (
	"fmt"
)

// Lisp object is used as AST, Lisp code, and secd machine code
// Type of Lisp Object
const (
	LispBoolean = -(iota + 1)
	LispSymbol
	LispChar
	LispVector
	LispProcedure // built in, Value is go function
	LispClosure   // compound car is code, cdr is env
	LispPair
	LispNumber
	LispString
	LispPort
	LispNull
)

// car & cdr is only used when Type is LispPair
type LObj struct {
	Type  int
	Value interface{}
	Car   *LObj
	Cdr   *LObj
}

// utilities
var lispFalse = LObj{Type: LispBoolean, Value: false}
var lispTrue = LObj{Type: LispBoolean, Value: true}
var lispNull = LObj{Type: LispNull}

// pair to string (recursive)
func (obj LObj) pairString() string {
	var text string
	text = fmt.Sprintf("%v", *obj.Car)
	switch obj.Cdr.Type {
	case LispNull:
		text += ")"
	case LispPair:
		text += " " + obj.Cdr.pairString()
	default:
		text += " . " + obj.Cdr.String() + ")"
	}
	return text
}

func (obj LObj) String() (text string) {
	switch obj.Type {
	case LispBoolean:
		if obj.Value == true {
			text = "#t"
		} else {
			text = "#f"
		}
	case LispPair:
		text = fmt.Sprintf("(%v", obj.pairString())
	case LispString:
		text = fmt.Sprintf("\"%v\"", obj.Value)
	case LispNull:
		text = "()"
	case LispChar:
		text = string(obj.Value.(rune))
	default:
		text = fmt.Sprintf("%v", obj.Value)
	}
	return text
}

// convert lisp object to go bool
func (obj LObj) ToBool() bool {
	return !(obj.Type == LispBoolean && !obj.Value.(bool))
}

// predicates
func (obj LObj) IsBoolean() bool {
	return obj.Type == LispBoolean
}

func (obj LObj) IsPair() bool {
	return obj.Type == LispPair
}

func (obj LObj) IsSymbol() bool {
	return obj.Type == LispSymbol
}

func (obj LObj) IsNull() bool {
	return obj.Type == LispNull
}

func (obj LObj) IsList() bool {
	if obj.IsPair() {
		return obj.Cdr.IsList()
	}
	return obj.IsNull()
}

// List utilities

// car with type check
func (obj LObj) SafeCar() (LObj, error) {
	if obj.IsPair() {
		return *obj.Car, nil
	} else {
		return lispFalse, fmt.Errorf("car: %v is not pair", obj)
	}
}

// cdr with type check
func (obj LObj) SafeCdr() (LObj, error) {
	if obj.IsPair() {
		return *obj.Cdr, nil
	} else {
		return lispFalse, fmt.Errorf("cdr: %v is not pair", obj)
	}
}

func (obj LObj) SetCar(car LObj) error {
	if obj.IsPair() {
		*obj.Car = car
		return nil
	} else {
		return fmt.Errorf("car: bad argument type: %v", obj)
	}
}

func (obj LObj) SetCdr(cdr LObj) error {
	if obj.IsPair() {
		*obj.Cdr = cdr
		return nil
	} else {
		return fmt.Errorf("cdr: bad argument type: %v", obj)
	}
}

func Cons(car, cdr LObj) LObj {
	return LObj{Type: LispPair, Car: &car, Cdr: &cdr}
}

func (obj LObj) ListRef(n int) (LObj, error) {
	var err error
	// range check
	if n < 0 {
		return lispFalse, fmt.Errorf("list-ref: out of range, %d", n)
	}
	// null check
	if obj.IsNull() {
		return lispFalse, fmt.Errorf("list-ref: null value")
	}
	// cdr down loop
	for {
		if n == 0 {
			return obj.SafeCar()
		}
		n -= 1                   // decrement
		obj, err = obj.SafeCdr() // cdr down
		if err != nil {
			return obj, err
		}
	}
}

// destructive cdr and return car
func (pair *LObj) Pop() (LObj, error) {
	ret, err := pair.SafeCdr()
	if err != nil {
		return ret, err
	}
	*pair = ret
	return ret, err
}

// obj -> (a obj)
func (obj *LObj) Push(car LObj) {
	*obj = Cons(car, *obj)
}

// compare obj's representation with s
func (sym LObj) TextEq(s string) bool {
	return sym.String() == s
}

func (sym1 LObj) SymEq(sym2 LObj) bool {
	return sym1.IsSymbol() && sym2.IsSymbol() && (sym1.String() == sym2.String())
}

func (obj1 LObj) Eq(obj2 LObj) bool {
	if obj1.IsSymbol() && obj2.IsSymbol() {
		return obj1.Value == obj2.Value
	}
	if obj1.IsNull() && obj2.IsNull() {
		return true
	}
	return &obj1 == &obj2
}

// pair or #f
func (alist LObj) Assq(sym LObj) (LObj, error) {
	if alist.IsNull() {
		return lispFalse, nil
	}

	pair, err := alist.SafeCar()
	if err != nil {
		return alist, err
	}
	compsym, err := pair.SafeCar()
	if err != nil {
		return pair, err
	}
	if sym.Eq(compsym) {
		return pair, nil
	}
	return alist.Cdr.Assq(sym)
}

func (env LObj) LookUp(sym LObj) (LObj, error) {
	if env.IsNull() {
		return lispFalse, fmt.Errorf("unbound variable: %v", sym)
	}
	currentEnv, err := env.SafeCar()
	if err != nil {
		return env, err
	}
	// lookup current environment
	pair, err := currentEnv.Assq(sym)
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
