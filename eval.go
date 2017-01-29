package rgors

import (
	"fmt"
)

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
