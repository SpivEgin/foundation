package main

import (
	"io"
	"strings"

	"github.com/robertkrimen/otto"
	"github.com/chzyer/readline"
)

func startInteraction() error {
	vm := otto.New()

	prompt := "ottemo> "
	cmdline, err := readline.NewEx(
		&readline.Config {
			Prompt:       prompt,
			AutoComplete: nil,
		})
	if err != nil {
		return err
	}

	var multiline []string

	for {
		line, err := cmdline.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				if multiline != nil {
					multiline = nil

					cmdline.SetPrompt(prompt)
					cmdline.Refresh()

					continue
				}

				break
			}

			return err
		}

		if line == "" {
			continue
		}

		multiline = append(multiline, line)

		jsCode, err := vm.Compile("repl", strings.Join(multiline, "\n"))
		if err != nil {
			cmdline.SetPrompt(strings.Repeat(" ", len(prompt)))
		} else {
			cmdline.SetPrompt(prompt)

			multiline = nil

			v, err := vm.Eval(jsCode)
			if err != nil {
				if ottoErr, ok := err.(*otto.Error); ok {
					io.Copy(cmdline.Stdout(), strings.NewReader(ottoErr.String()))
				} else {
					io.Copy(cmdline.Stdout(), strings.NewReader(err.Error()))
				}
			} else {
				cmdline.Stdout().Write([]byte(v.String() + "\n"))
			}
		}

		cmdline.Refresh()
	}

	return cmdline.Close()

}