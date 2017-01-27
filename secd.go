package lispy

import (
	"fmt"
)

func NewEnv() *LObj {
	env := Cons(lispNull, lispNull)
	return &env
}

// about environment (list of alist)
func (env *LObj) LookUp(sym LObj) (LObj, error) {
	if env.IsNull() {
		return lispFalse, fmt.Errorf("unbound variable: %v", sym)
	}
	currentEnv, err := env.SafeCar()
	if err != nil {
		return *env, err
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

// destructive
func (env *LObj) Define(sym LObj, val LObj) {
	env.Car.Push(Cons(sym, val))
}

// return new extended env
func (parent *LObj) Extend(child *LObj) *LObj {
	result := Cons(*child, *parent)
	return &result
}

func InitialEnv() *LObj {
	var lispAdd2 = LObj{
		Type: LispBuiltin,
		Value: func(obj1, obj2 LObj) (LObj, error) {
			if obj1.IsNumber() && obj2.IsNumber() {
				return LObj{Type: Number, Value: obj1.Value.(float64) + obj2.Value.(float64)}, nil
			} else {
				return lispFalse, fmt.Errorf("not number's %v + %v", obj1, obj2)
			}
		},
	}
	sym, _ := NewSymbol("+")
	env := NewEnv()
	env.Define(sym, lispAdd2)
	return env
}

// closure
func NewClosure(code LObj, env LObj) LObj {
	return LObj{
		Type: LispClosure,
		Car:  &code,
		Cdr:  &env,
	}
}

func (closure *LObj) Code() *LObj {
	return closure.Car
}
func (closure *LObj) Env() *LObj {
	return closure.Cdr
}

// var lispPlus = LObj{Type: LispBuiltin, Value: LispAdd2}
// var initFrame = lispNull

// initFrame.Push(Cons(NewSymbol("+"), lispPlus))
// var InitEnv = Cons(initFrame, lispNull)

type SECD struct {
	Stack       *LObj
	Environment *LObj
	Code        *LObj
	Dump        *LObj
}

// nil: falseをスタックにプッシュ
// ldc: 定数オペランドをスタックにプッシュ
// ld: 変数の値をスタックにプッシュ。変数はオペランドで環境レベルと順番で指定される。例えば "(1 . 3)" なら、現在の関数レベルで3番めの引数を意味する。
// sel: 2つのリストをオペランドに持ち、スタックから値を1つポップする。ポップした値が nil でない場合、先頭のリストを実行し、そうでなければ2番めのリストを実行する。いずれかのリストへのポインタが Cレジスタに格納される前に、sel命令の次の命令へのポインタがダンプにセーブされる。
// join: ダンプからリスト参照をポップし、それをCレジスタにセットする。これはsel命令で選択されたリストの実行が完了したときに実行される。
// ldf: 関数を表す1つのリストをオペランドに持つ。クロージャ（関数と現在の環境のペア）を構築し、それをスタックにプッシュする。
// ap: クロージャと引数（の値）リストをスタックからポップする。クロージャを現在の環境として設定し、引数に適用する。引数リストを環境に設定し、スタックをクリアしてCレジスタにクロージャ内にある関数ポインタをセットする。以前のSとEレジスタの値、Cの次の値はダンプにセーブしておく。
// ret: スタックからリターン値をポップし、ダンプからS、E、Cをリストアする。そしてリターン値を新たな現在のスタックにプッシュする。
// dum: ダミーを環境リストの先頭にプッシュする。ダミーとは空リストである。
// rap: ap命令と類似しているが、ダミー環境と組み合わせて、再帰関数を実現するのに使われる。
// car、cdr、リスト構築、整数の加算、入出力といった基本的な関数も命令として存在する。これらは必要な引数をスタックから得る。
// stop: stop 命令
func (secd *SECD) step() error {
	sym, err := secd.Code.Pop()
	if err != nil {
		return err
	}
	switch sym.String() {
	case "nil":
		secd.Stack.Push(lispFalse)
	case "ldc":
		cst, err := secd.Code.Pop()
		if err != nil {
			return err
		}
		secd.Stack.Push(cst)
	case "ld":
		sym, err := secd.Code.Pop()
		if err != nil {
			return err
		}
		val, err := secd.Environment.LookUp(sym)
		if err != nil {
			return err
		}
		secd.Stack.Push(val)
	case "sel":
		flag, _ := secd.Stack.Pop()
		truecode, _ := secd.Code.Pop()
		falsecode, _ := secd.Code.Pop()
		secd.Dump.Push(*secd.Code)
		if flag.ToBool() {
			*secd.Code = truecode
		} else {
			*secd.Code = falsecode
		}
	case "join":
		c, err := secd.Dump.Pop()
		if err != nil {
			return err
		}
		*secd.Code = c
	case "ldf":
		code, err := secd.Code.Pop()
		if err != nil {
			return err
		}
		closure := NewClosure(code, *secd.Environment)
		secd.Stack.Push(closure)

	}
}
