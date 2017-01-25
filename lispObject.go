package lispy

import (
	"fmt"
)

// Type of Lisp Object
const (
	LispBoolean = -(iota + 1)
	LispSymbol
	LispChar
	LispVector
	LispProcedure // built in
	LispClosure   // compound
	LispPair
	LispNumber
	LispString
	LispPort
	LispNil
)

type LObj struct {
	Type  int
	Value interface{}
	Car   *LObj
	Cdr   *LObj
}

func (obj LObj) pairString() string {
	var text string
	text = fmt.Sprintf("%v", *obj.Car)
	switch obj.Cdr.Type {
	case LispNil:
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
	case LispNil:
		text = "()"
	case LispChar:
		text = string(obj.Value.(rune))
	default:
		text = fmt.Sprintf("%v", obj.Value)
	}
	return text
}

func (obj LObj) ToBool() bool {
	return !(obj.Type == LispBoolean && !obj.Value.(bool))
}

func (obj LObj) IsBoolean() bool {
	return obj.Type == LispBoolean
}

func (obj LObj) IsPair() bool {
	return obj.Type == LispPair
}

func (obj LObj) IsNull() bool {
	return obj.Type == LispNil
}

func (obj LObj) IsList() bool {
	if obj.IsPair() {
		return obj.Cdr.IsList()
	}
	return obj.IsNull()
}
