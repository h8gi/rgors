package rgors

// http://www.brics.dk/RS/03/33/BRICS-RS-03-33.pdf

import ()

// type SECD struct {
// 	Stack       *LObj // value stack (list)
// 	Environment *LObj // environment (list)
// 	Control     *LObj // code		 (list)
// 	Dump        *LObj // (s e c) dump (list)
// }

// func (p Program) Compile() Program {
// 	for _, exp := range p {

// 	}
// }

// func (obj *LObj) Compile() LObj {
// 	switch {
// 	case obj.IsSelfEvaluating():
// 		return *obj
// 	case obj.IsSymbol():
// 		return *obj
// 	case obj.IsList():
// 		t1, err := obj.ListRef(0)
// 		t2, err := obj.ListRef(1)
// 		NewList(NewSymbol("app"), t1.Compile(), t2.Compile())
// 	}
// }
