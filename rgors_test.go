package rgors

import (
	"fmt"
	"testing"
)

func TestParser(t *testing.T) {
	parser := Parser{}
	program, err := parser.ParseString("(* 2 3) '(hello) \"foo bar\"")
	if err != nil {
		t.Errorf("parser fail: %s", err)
	}

	for _, exp := range program {
		fmt.Println(exp)
	}
}

func TestEq(t *testing.T) {
	parser := Parser{}
	a1, _ := parser.str2expr("a")
	a2, _ := parser.str2expr("a")
	if !a1.Eq(&a2) {
		t.Errorf("fail: symbol compare: %v == %v", a1, a2)
	}
	list1, _ := parser.str2expr("(1 2 3)")
	list2, _ := parser.str2expr("(1 2 3)")
	list3 := list1
	if list1.Eq(&list2) {
		t.Errorf("fail: list compare: %v != %v", list1, list2)
	}
	if !list1.Eq(&list3) {
		t.Errorf("fail: list compare: %v == %v", list1, list3)
	}

	s1, _ := parser.str2expr("\"hello\"")
	s2, _ := parser.str2expr("\"hello\"")
	if !s1.Eq(&s2) {
		t.Error("fail: str compare: %v != %v", s1, s2)
	}

	a := NewSymbol("a")
	b := NewSymbol("b")

	cns1 := Cons(*a, *b)
	cns2 := Cons(*a, *b)
	if cns1.Eq(&cns2) {
		t.Error("fail: cons compare: %v != %v", cns1, cns2)
	}
}

func TestPop(t *testing.T) {
	parser := Parser{}
	list, _ := parser.str2expr("(a b c)")
	sym, _ := list.Pop()
	if !sym.Eq(NewSymbol("a")) {
		t.Error("fail: pop")
	}
	sym, _ = list.Pop()
	if !sym.Eq(NewSymbol("b")) {
		t.Error("fail: pop")
	}
}

func TestLength(t *testing.T) {
	parser := Parser{}

	var testLists map[string]int = map[string]int{
		"(1 2 3)":   3,
		"()":        0,
		"(3 4 5 7)": 4,
	}

	for str, expect := range testLists {
		list, err := parser.str2expr(str)
		fmt.Println("len ", expect, list)
		if err != nil {
			t.Errorf("parser fail: %s", err)
		}
		len, err := list.Length()
		if err != nil {
			t.Errorf("length fail: %s", err)
		}
		if len != expect {
			t.Errorf("expect %d, but %d", expect, len)
		}

		len, err = list.Length()
		if err != nil {
			t.Errorf("length fail: %s", err)
		}
		if len != expect {
			t.Errorf("expect %d, but %d", expect, len)
		}
		fmt.Println("len ", expect, list)
	}

}

func TestList(t *testing.T) {
	parser := Parser{}
	list1, _ := parser.str2expr("(a b c)")
	list2 := NewList(*NewSymbol("a"), *NewSymbol("b"), *NewSymbol("c"))
	if list1.String() != list2.String() {
		t.Errorf("list constructor fail")
	}
	if list1.Eq(&list2) {
		t.Errorf("Eq fail")
	}
}

func TestLookup(t *testing.T) {
	parser := Parser{}
	vars, _ := parser.str2expr("(a b c)")
	vals, _ := parser.str2expr("(1 2 3)")
	env := LispNull.Extend(vars, vals)
	fmt.Println(env)
	fmt.Println(env.LookUp(NewSymbol("c")))
	fmt.Println(env)
	newvals, _ := env.LookUp(NewSymbol("c"))
	newvals.SetCar(LispFalse)
	fmt.Println(env)
}

func TestCompileLookUp(t *testing.T) {
	env := LispNull.CompileExtend(NewList(*NewSymbol("a"), *NewSymbol("b"), *NewSymbol("c")))

	rib, elt, _ := env.CompileLookUp(NewSymbol("c"))
	if !(rib == 0 && elt == 2) {
		t.Errorf("lookup fail")
	}

	env = env.CompileExtend(NewList(*NewSymbol("x"), *NewSymbol("y")))

	rib, elt, _ = env.CompileLookUp(NewSymbol("x"))
	if !(rib == 0 && elt == 0) {
		t.Errorf("lookup fail")
	}

	rib, elt, _ = env.CompileLookUp(NewSymbol("c"))
	if !(rib == 1 && elt == 2) {
		t.Errorf("lookup fail")
	}
	rib, elt, _ = env.CompileLookUp(NewSymbol("a"))
	if !(rib == 1 && elt == 0) {
		t.Errorf("lookup fail")
	}
}
