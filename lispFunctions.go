package lispy

// func IsList(obj LObj) LObj {
// 	if obj.Type == LispNil {
// 		return LObj{Type: LispBoolean, Value: true}
// 	}
// 	if obj.Type == LispPair {
// 		return IsList(*obj.Cdr)
// 	}
// 	return LObj{Type: LispBoolean, Value: false}
// }

// func Cons(car, cdr LObj) LObj {
// 	return LObj{Type: LispPair, Car: car, Cdr: cdr}
// }

// func Car(obj LObj) LObj {
// 	return *obj.Car
// }

// func Cdr(obj LObj) LObj {
// 	return *obj.Cdr
// }

// func IsNil(obj LObj) LObj {
// 	if obj.Type == LispNil {
// 		return LObj{Type: LispBoolean, Value: true}
// 	} else {
// 		return LObj{Type: LispBoolean, Value: false}
// 	}
// }
