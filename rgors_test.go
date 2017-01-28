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
