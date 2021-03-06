package rgors

import (
	"fmt"
)

// Lisp object is used as AST, Lisp code, and secd machine code
// type of Lisp Object
const (
	DTBoolean = -(iota + 1)
	DTSymbol
	DTChar
	DTVector
	DTPrimitive // built in, Value is go function
	DTClosure   // compound car is code, cdr is env
	DTPair
	DTNumber
	DTString
	DTPort
	DTNull
)

// car & cdr is only used when Type is DTPair
type LObj struct {
	Type  int
	Value interface{}
	Car   *LObj
	Cdr   *LObj
}

// utilities
var LispFalse = LObj{Type: DTBoolean, Value: false}
var LispTrue = LObj{Type: DTBoolean, Value: true}
var LispNull = LObj{Type: DTNull}

// pair to string (recursive)
func (obj LObj) pairString() string {
	var text string
	text = fmt.Sprintf("%v", *obj.Car)
	switch obj.Cdr.Type {
	case DTNull:
		text += ")"
	case DTPair:
		text += " " + obj.Cdr.pairString()
	default:
		text += " . " + obj.Cdr.String() + ")"
	}
	return text
}

func (obj LObj) String() (text string) {
	switch obj.Type {
	case DTBoolean:
		if obj.Value == true {
			text = "#t"
		} else {
			text = "#f"
		}
	case DTPair:
		text = fmt.Sprintf("(%v", obj.pairString())
	case DTString:
		text = fmt.Sprintf("\"%v\"", obj.Value)
	case DTNull:
		text = "()"
	case DTChar:
		text = string(obj.Value.(rune))
	case DTPrimitive:
		text = "<primitive>"
	case DTClosure:
		text = fmt.Sprintf("(^ %v : %v)", obj.Body(), obj.Env())
	default:
		text = fmt.Sprintf("%v", obj.Value)
	}
	return text
}

// convert lisp object to go bool
func (obj *LObj) ToBool() bool {
	return !(obj.Type == DTBoolean && !obj.Value.(bool))
}

// predicates
func (obj *LObj) IsBoolean() bool {
	return obj.Type == DTBoolean
}

func (obj *LObj) IsPair() bool {
	return obj.Type == DTPair
}

func (obj *LObj) IsSymbol() bool {
	return obj.Type == DTSymbol
}

func (obj *LObj) IsNull() bool {
	return obj.Type == DTNull
}

func (obj *LObj) IsNumber() bool {
	return obj.Type == DTNumber
}

func (obj *LObj) IsList() bool {
	if obj.IsPair() {
		return obj.Cdr.IsList()
	}
	return obj.IsNull()
}

func (obj *LObj) IsSelfEvaluating() bool {
	switch obj.Type {
	case DTBoolean, DTChar, DTString, DTNumber, DTVector:
		return true
	default:
		return false
	}
}

func (obj *LObj) IsClosure() bool {
	return obj.Type == DTClosure
}

func (obj *LObj) IsPrimitive() bool {
	return obj.Type == DTPrimitive
}

func (obj *LObj) IsProcedure() bool {
	return obj.IsClosure() || obj.IsPrimitive()
}

// List utilities

// car with type check
func (obj *LObj) SafeCar() (LObj, error) {
	if obj.IsPair() {
		return *obj.Car, nil
	} else {
		return LispFalse, fmt.Errorf("car: %v is not pair", obj)
	}
}

// cdr with type check
func (obj *LObj) SafeCdr() (LObj, error) {
	if obj.IsPair() {
		return *obj.Cdr, nil
	} else {
		return LispFalse, fmt.Errorf("cdr: %v is not pair", obj)
	}
}

func (obj *LObj) SetCar(car LObj) error {
	if obj.IsPair() {
		*obj.Car = car
		return nil
	} else {
		return fmt.Errorf("car: bad argument type: %v", obj)
	}
}

func (obj *LObj) SetCdr(cdr LObj) error {
	if obj.IsPair() {
		*obj.Cdr = cdr
		return nil
	} else {
		return fmt.Errorf("cdr: bad argument type: %v", obj)
	}
}

func Cons(car, cdr LObj) LObj {
	return LObj{Type: DTPair, Car: &car, Cdr: &cdr}
}

func (obj *LObj) ListRef(n int) (LObj, error) {
	var err error
	// range check
	if n < 0 {
		return LispFalse, fmt.Errorf("list-ref: out of range, %d", n)
	}
	// null check
	if obj.IsNull() {
		return LispFalse, fmt.Errorf("list-ref: null value")
	}
	// cdr down loop
	var elem LObj = *obj
	for {
		if n == 0 {
			return elem.SafeCar()
		}
		n -= 1                     // decrement
		elem, err = elem.SafeCdr() // cdr down
		if err != nil {
			return elem, err
		}
	}
}

// destructive cdr and return car
func (pair *LObj) Pop() (LObj, error) {
	ret, err := pair.SafeCar()
	if err != nil {
		return ret, err
	}
	*pair, err = pair.SafeCdr()
	return ret, err
}

// obj -> (a obj)
func (obj *LObj) Push(car *LObj) {
	*obj = Cons(*car, *obj)
}

// compare by pointer
func (obj1 *LObj) Eq(obj2 *LObj) bool {
	return *obj1 == *obj2
}

// utility
func (obj *LObj) CarEq(s string) bool {
	return obj.Car.Eq(NewSymbol(s))
}

// reuturn: pair or #f
func (alist *LObj) Assq(sym LObj) (LObj, error) {
	copiedobj := *alist
	for {
		if copiedobj.IsNull() {
			return LispFalse, nil
		}
		pair, err := copiedobj.SafeCar()
		if err != nil {
			return copiedobj, err
		}
		compsym, err := pair.SafeCar()
		if err != nil {
			return pair, err
		}
		if sym.Eq(&compsym) {
			return pair, nil
		}
		copiedobj = *copiedobj.Cdr
	}
}

var symbolTable map[string]*LObj = make(map[string]*LObj, 0)

func NewSymbol(s string) *LObj {
	sym, ok := symbolTable[s] // search intern table
	if ok {
		return sym
	}
	symbolTable[s] = &LObj{Type: DTSymbol, Value: s} // intern
	return symbolTable[s]                            // return symbol
}

func NewList(objs ...LObj) LObj {
	pair := LispNull
	for i := len(objs) - 1; i >= 0; i-- {
		pair = Cons(objs[i], pair)
	}
	return pair
}

func (obj *LObj) Length() (int, error) {
	if !obj.IsList() {
		return 0, fmt.Errorf("not a list: %v", obj)
	}
	count := 0
	copiedobj := *obj
	for {
		if copiedobj.IsNull() {
			return count, nil
		}
		count += 1
		copiedobj = *copiedobj.Cdr
	}
}

func NewVector(objs ...LObj) LObj {
	return LObj{Type: DTVector, Value: objs}
}
