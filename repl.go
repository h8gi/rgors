package rgors

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
	var contFlag bool // for multiple lines
	fmt.Println("Lispy Version 0.0.0.0.1")
	fmt.Println("Press Ctrl+c to Exit")
	rl, err := readline.New("rgors> ")
	if err != nil {
		panic(err)
	}

	p := Parser{}

	for {

		// readline
		tmpline, err := rl.Readline()
		if contFlag {
			line = line + "\n" + tmpline
			contFlag = false
			rl.SetPrompt("rgors> ")
		} else {
			line = tmpline
		}
		if err != nil {
			break
		}

		// lexer check
		// display(&p, line)
		// parse
		program, err := p.ParseString(line)
		if err != nil {
			switch err.(type) {
			case *UnclosedError: // continue
				contFlag = true
				rl.SetPrompt("... ")
				continue
			default:
				fmt.Println(err.Error())
				continue
			}
		}
		// eval???
		vm := NewVM()
		for _, expr := range program {
			comp, err := expr.Compile()
			if err != nil {
				fmt.Println("compile error:", err.Error())
				continue
			}
			fmt.Println(comp)

			// eval!!
			vm.Load(comp)
			ans, err := vm.Run()
			if err != nil {
				fmt.Println("vm error:", err.Error())
				continue
			}
			fmt.Println("=>", comp, "\n=>", ans)
		}
	}
}
