package lispy

import (
	"fmt"
)

// Lisp object is used as AST, Lisp code, and secd machine code
// Type of Lisp Object
const (
	LispTBoolean = -(iota + 1)
	LispTSymbol
	LispTChar
	LispTVector
	LispTBuiltin // built in, Value is go function
	LispTClosure // compound car is code, cdr is env
	LispTPair
	LispTNumber
	LispTString
	LispTPort
	LispTNull
)

// car & cdr is only used when Type is LispTPair
type LObj struct {
	Type  int
	Value interface{}
	Car   *LObj
	Cdr   *LObj
}

// utilities
var LispFalse = LObj{Type: LispTBoolean, Value: false}
var LispTrue = LObj{Type: LispTBoolean, Value: true}
var LispNull = LObj{Type: LispTNull}

// pair to string (recursive)
func (obj LObj) pairString() string {
	var text string
	text = fmt.Sprintf("%v", *obj.Car)
	switch obj.Cdr.Type {
	case LispTNull:
		text += ")"
	case LispTPair:
		text += " " + obj.Cdr.pairString()
	default:
		text += " . " + obj.Cdr.String() + ")"
	}
	return text
}

func (obj LObj) String() (text string) {
	switch obj.Type {
	case LispTBoolean:
		if obj.Value == true {
			text = "#t"
		} else {
			text = "#f"
		}
	case LispTPair:
		text = fmt.Sprintf("(%v", obj.pairString())
	case LispTString:
		text = fmt.Sprintf("\"%v\"", obj.Value)
	case LispTNull:
		text = "()"
	case LispTChar:
		text = string(obj.Value.(rune))
	default:
		text = fmt.Sprintf("%v", obj.Value)
	}
	return text
}

// convert lisp object to go bool
func (obj LObj) ToBool() bool {
	return !(obj.Type == LispTBoolean && !obj.Value.(bool))
}

// predicates
func (obj LObj) IsBoolean() bool {
	return obj.Type == LispTBoolean
}

func (obj LObj) IsPair() bool {
	return obj.Type == LispTPair
}

func (obj LObj) IsSymbol() bool {
	return obj.Type == LispTSymbol
}

func (obj LObj) IsLispNull() bool {
	return obj.Type == LispTNull
}

func (obj LObj) IsNumber() bool {
	return obj.Type == LispTNumber
}

func (obj LObj) IsList() bool {
	if obj.IsPair() {
		return obj.Cdr.IsList()
	}
	return obj.IsLispNull()
}

// List utilities

// car with type check
func (obj LObj) SafeCar() (LObj, error) {
	if obj.IsPair() {
		return *obj.Car, nil
	} else {
		return LispFalse, fmt.Errorf("car: %v is not pair", obj)
	}
}

// cdr with type check
func (obj LObj) SafeCdr() (LObj, error) {
	if obj.IsPair() {
		return *obj.Cdr, nil
	} else {
		return LispFalse, fmt.Errorf("cdr: %v is not pair", obj)
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
	return LObj{Type: LispTPair, Car: &car, Cdr: &cdr}
}

func (obj LObj) ListRef(n int) (LObj, error) {
	var err error
	// range check
	if n < 0 {
		return LispFalse, fmt.Errorf("list-ref: out of range, %d", n)
	}
	// null check
	if obj.IsLispNull() {
		return LispFalse, fmt.Errorf("list-ref: null value")
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

// compare by pointer
func (obj1 LObj) Eq(obj2 LObj) bool {
	return obj1 == obj2
}

// reuturn: pair or #f
func (alist LObj) Assq(sym LObj) (LObj, error) {
	if alist.IsLispNull() {
		return LispFalse, nil
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

var symbolTable map[string]LObj = make(map[string]LObj, 0)

func NewSymbol(s string) LObj {
	sym, ok := symbolTable[s] // search intern table
	if ok {
		return sym
	}
	symbolTable[s] = LObj{Type: LispTSymbol, Value: s} // intern
	return symbolTable[s]                              // return symbol
}

func NewList(objs ...LObj) LObj {
	pair := LispNull
	for i := len(objs) - 1; i >= 0; i-- {
		pair = Cons(objs[i], pair)
	}
	return pair
}
