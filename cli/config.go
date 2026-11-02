package cli

import (
	"fmt"
	"strings"

	"github.com/chzyer/readline"
)

const argNameValSeperator = "="

type Config struct {
	lastInput      []string
	commands       map[string]cliCommand
	rlAutoComplete *readline.PrefixCompleter
	Rl             *readline.Instance
}

func InitConfig() *Config {
	c := Config{}
	c.Rl, c.rlAutoComplete = InitReadline()
	c.commands = map[string]cliCommand{}

	return &c
}

type CommandArg struct {
	Name     string
	Val      string
	Required bool
}

type cliCommand struct {
	name     string
	function func([]CommandArg) error
	args     []CommandArg
}

func (c *Config) AddCommand(name string, function func([]CommandArg) error, args ...CommandArg) {
	newCommand := cliCommand{
		name:     name,
		function: function,
		args:     args,
	}

	c.commands[name] = newCommand

	c.rlAutoComplete.Children = append(c.rlAutoComplete.Children,
		readline.PcItem(name),
	)

	argsChildren := []readline.PrefixCompleterInterface{}
	for _, arg := range args {
		argsChildren = append(argsChildren,
			readline.PcItem(arg.Name+"="),
		)
	}
	c.rlAutoComplete.Children[len(c.rlAutoComplete.Children)-1].SetChildren(argsChildren)
}

func (c *Config) commandLookup(input string) (cliCommand, bool) {
	cmd, ok := c.commands[input]
	return cmd, ok
}

func (c *Config) CommandExe(input string) error {
	cleanInputAndStore(c, input)

	if len(c.lastInput) == 0 {
		return nil
	}

	command, args, err := c.parseInput()
	if err != nil {
		if err.Error() == "no input" {
			return nil
		} else {
			return err
		}
	}

	err = c.commands[command].function(args)
	if err != nil {
		return err
	}

	return nil

}

func (c *Config) parseInput() (string, []CommandArg, error) {
	//-- Handle no input given. Error should be handled by continuing readline loop:
	if len(c.lastInput) == 0 {
		return "", nil, fmt.Errorf("no input")
	}

	//-- Find valid command in input:
	command := ""
	argStrings := []string{}
	for i := len(c.lastInput); i >= 0; i-- {
		potentialCommand := strings.Join(c.lastInput[0:i], " ")
		if _, okay := c.commands[potentialCommand]; okay {
			command = potentialCommand
			if i < len(c.lastInput) {
				argStrings = c.lastInput[i:]
			}
			break
		}
	}

	//-- Concatenate multi word args
	tempArgs := argStrings
	argStrings = []string{}
	for i := 0; i < len(tempArgs); i++ {
		if startsWithQuote, wordNoQuote := isStartQuoteWord(tempArgs[i]); !startsWithQuote {
			argStrings = append(argStrings, tempArgs[i])
		} else {
			argSlices := []string{wordNoQuote}
			i++
			endsWithQuote, wordNoEndQuote := isEndQuoteWord(tempArgs[i])
			for !endsWithQuote && i+1 < len(tempArgs) {
				argSlices = append(argSlices, tempArgs[i])
				i++
				endsWithQuote, wordNoEndQuote = isEndQuoteWord(tempArgs[i])
			}
			if !endsWithQuote {
				argSlices = append(argSlices, tempArgs[i])
			} else {
				argSlices = append(argSlices, wordNoEndQuote)

			}
			argStrings = append(argStrings, strings.Join(argSlices, " "))
		}
	}

	if command == "" {
		return "", nil, fmt.Errorf("error. command not found")
	}

	args := []CommandArg{}
	expectedArgs := c.commands[command].args

	//-- Handle case of no arguments expected + none given and too many arguments given:
	switch len(expectedArgs) {
	case 0:
		switch len(argStrings) {
		case 0:
			return command, args, nil

		default:
			return "", nil, fmt.Errorf("error. %s command takes no arguments", command)
		}

	default:
		if len(argStrings) > len(expectedArgs) {
			return "", nil, fmt.Errorf("error. too many arguments. %s command takes a max of %d arguments. %d arguments given", command, len(expectedArgs), len(argStrings))
		}
	}

	//-- Attempt to match entered arguments to expected arguments:
	remainingArgStrings := []string{}
	remainingExpectedArgs := []CommandArg{}

	for _, argString := range argStrings {
		if strings.Contains(argString, argNameValSeperator) {
			splitArgString := strings.Split(argString, argNameValSeperator)
			matchedArg := false

			for i, expectedArg := range expectedArgs {
				if splitArgString[0] == expectedArg.Name {
					args = append(args, CommandArg{
						Name:     expectedArg.Name,
						Val:      splitArgString[1],
						Required: expectedArg.Required,
					})
					matchedArg = true
					remainingExpectedArgs = append(remainingExpectedArgs, expectedArgs[i+1:]...)
					break
				} else {
					remainingExpectedArgs = append(remainingExpectedArgs, expectedArg)
				}
			}
			expectedArgs = remainingExpectedArgs
			remainingExpectedArgs = []CommandArg{}

			if !matchedArg {
				remainingArgStrings = append(remainingArgStrings, argString)
			}
		} else {
			remainingArgStrings = append(remainingArgStrings, argString)
		}
	}
	argStrings = remainingArgStrings

	//-- For any entered arguments left, match them to remaining expected arguments:
	for _, argString := range argStrings {
		arg := expectedArgs[0]
		expectedArgs = expectedArgs[1:]
		arg.Val = argString

		args = append(args, arg)
	}

	//-- Append any remaining unmatched expected arguments:
	args = append(args, expectedArgs...)

	//-- Check to make sure all required arguments have a value:
	for _, arg := range args {
		if arg.Required && arg.Val == "" {
			return "", nil, fmt.Errorf("error. %s argument required for %s command", arg.Name, command)
		}
	}

	return command, args, nil
}

func (c *Config) parseInputForCommand() string {
	nonArgs := []string{}

	for _, i := range c.lastInput {
		if !strings.Contains(i, "=") {
			nonArgs = append(nonArgs, i)
		}
	}

	if len(nonArgs) == 0 {
		return ""
	}

	return strings.Join(nonArgs, " ")

}

func (c *Config) parseInputForArgs() []CommandArg {
	args := []string{}

	for _, i := range c.lastInput {
		if strings.Contains(i, "=") {
			args = append(args, i)
		}
	}

	commandArgs := []CommandArg{}
	for _, arg := range args {
		splitArg := strings.Split(arg, "=")
		commandArgs = append(commandArgs, CommandArg{
			Name: splitArg[0],
			Val:  splitArg[1],
		})
	}

	return commandArgs
}

func cleanInput(text string) []string {
	var textWords []string
	text = strings.TrimSpace(text)
	firstPass := strings.Split(text, " ")

	for _, word := range firstPass {
		if word != "" {
			textWords = append(textWords, word)
		}
	}

	return textWords
}

func cleanInputAndStore(c *Config, input string) {
	c.lastInput = cleanInput(input)
}

func isStartQuoteWord(word string) (bool, string) {
	if !strings.Contains(word, "\"") {
		return false, ""
	}

	wordSlices := []rune(word)
	if wordSlices[0] != '"' {
		return false, ""
	}

	wordNoQuote := string(wordSlices[1:])
	return true, wordNoQuote
}

func isEndQuoteWord(word string) (bool, string) {
	if !strings.Contains(word, "\"") {
		return false, ""
	}

	wordSlices := []rune(word)
	lastIndex := len(wordSlices) - 1
	if wordSlices[lastIndex] != '"' {
		return false, ""
	}

	wordNoQuote := string(wordSlices[:lastIndex])
	return true, wordNoQuote
}
