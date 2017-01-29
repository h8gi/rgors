package rgors

import (
	"fmt"
)

func InitialEnv() *LObj {
	env := NewEnv()
	var lispAdd2 = LObj{
		Type: DTPrimitive,
		Value: func(obj1, obj2 LObj) (LObj, error) {
			if obj1.IsNumber() && obj2.IsNumber() {
				return LObj{Type: DTNumber, Value: obj1.Value.(int) + obj2.Value.(int)}, nil
			} else {
				return LispFalse, fmt.Errorf("+: not number: %v, %v", obj1, obj2)
			}
		},
	}
	sym := NewSymbol("+")
	env.Define(sym, &lispAdd2)

	var lispSub2 = LObj{
		Type: DTPrimitive,
		Value: func(obj1, obj2 LObj) (LObj, error) {
			if obj1.IsNumber() && obj2.IsNumber() {
				return LObj{Type: DTNumber, Value: obj1.Value.(int) - obj2.Value.(int)}, nil
			} else {
				return LispFalse, fmt.Errorf("-: not number: %v, %v", obj1, obj2)
			}
		},
	}
	sym = NewSymbol("-")
	env.Define(sym, &lispSub2)

	return env
}

var env = InitialEnv()

func (obj *LObj) SimpleEval() (LObj, error) {
	switch {
	case obj.IsSelfEvaluating():
		return *obj, nil
	case obj.IsSymbol():
		return env.LookUp(obj)
	case obj.IsList():
		op, err := obj.ListRef(0)
		if err != nil {
			return *obj, err
		}
		op, err = op.SimpleEval()
		if err != nil {
			return *obj, err
		}

		arg1, err := obj.ListRef(1)
		if err != nil {
			return *obj, err
		}
		arg1, err = arg1.SimpleEval()
		if err != nil {
			return *obj, err
		}

		arg2, err := obj.ListRef(2)
		if err != nil {
			return *obj, err
		}
		arg2, err = arg2.SimpleEval()
		if err != nil {
			return *obj, err
		}
		return (op.Value.(func(LObj, LObj) (LObj, error)))(arg1, arg2)
	default:
		return LispFalse, fmt.Errorf("eval error: %v", obj)
	}
}
