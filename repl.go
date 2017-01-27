package lispy

import (
	"fmt"
	"github.com/chzyer/readline"
)

func display(p *Parser, line string) {
	p.SetString(line)
	tokens, _ := p.ReadTokens()
	// display innput
	fmt.Println("Lexer------------")
	for i, token := range tokens {
		fmt.Printf("%d: %+v\n", i, token)
		if token.Text == "quit" {
			return
		}
	}
	fmt.Println("------------Lexer")
}

func Repl() {
	var line string
	var err error
	var contFlag bool
	fmt.Println("Lispy Version 0.0.0.0.1")
	fmt.Println("Press Ctrl+c to Exit")
	rl, err := readline.New("lispy> ")
	if err != nil {
		panic(err)
	}

	p := Parser{}

	for {
		tmpline, err := rl.Readline()
		if contFlag {
			line = line + "\n" + tmpline
			contFlag = false
			rl.SetPrompt("lispy> ")
		} else {
			line = tmpline
		}
		if err != nil {
			break
		}
		display(&p, line)
		program, err := p.ParseString(line)

		if err != nil {
			switch err.(type) {
			case *UnclosedError:
				contFlag = true
				rl.SetPrompt("... ")
				continue
			default:
				fmt.Println(err.Error())
				continue
			}
		}

		for _, expr := range program {
			fmt.Println(expr)
			fmt.Println(expr.Assq(LObj{Type: LispSymbol, Value: "foo"}))
		}

		// obj, err := EvalProgram(program)
		// if err != nil {
		// 	fmt.Println(err.Error())
		// 	continue
		// }
		// fmt.Printf("%v\n", obj)
	}
}
