package rgors

import (
	"fmt"
	"testing"
)

func TestParser(t *testing.T) {
	parser := Parser{}
	program, err := parser.ParseString("(* 2 3) '(hello) \"foo bar\"")
	if err != nil {
		t.Errorf("parser fail %s", err)
	}

	for _, exp := range program {
		fmt.Println(exp)
	}
}
