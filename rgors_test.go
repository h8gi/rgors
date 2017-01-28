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
