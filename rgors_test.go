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

func TestEval(t *testing.T) {
	parser := Parser{}
	program, err := parser.ParseString("(+ 2 3) (- 10 2) (+ 2 (+ 2 5))")
	if err != nil {
		t.Errorf("parser fail: %s", err)
	}
	for _, exp := range program {
		ret, err := exp.SimpleEval()
		if err != nil {
			t.Errorf("eval fail: %s", err)
		}
		fmt.Printf("%v => %v\n", exp, ret)
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
