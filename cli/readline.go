package cli

import (
	"github.com/chzyer/readline"
)

func InitReadline() (*readline.Instance, *readline.PrefixCompleter) {

	completer := readline.NewPrefixCompleter()

	rl, _ := readline.NewEx(&readline.Config{
		Prompt:       "> ",
		AutoComplete: completer,
	})

	return rl, completer
}
