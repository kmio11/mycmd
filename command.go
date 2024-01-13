package mycmd

import (
	"context"
	"fmt"
	"io"
)

type (
	Command interface {
		Name() string
		ShortDescription() string
		Usage() string
		Parse(args []string) error
		IsHelpRequested(err error) bool
		Execute() int
		ExecuteContext(ctx context.Context) int
		Print(message string)
		PrintError(message string)
		SetOutWriter(w io.Writer)
		SetErrWriter(w io.Writer)
	}

	SubCommand interface {
		Command
		Parent() Command
		SetParent(parent Command)
	}

	ParentCommand interface {
		Command
		Commands() []Command
	}

	HelpSupported interface {
		FullHelpCommandName() string
	}

	HiddenSupported interface {
		Hidden() bool
	}
)

func parseCommand(c Command, args []string) (int, error) {
	err := c.Parse(args)
	if err != nil {
		if c.IsHelpRequested(err) {
			c.Print(c.Usage())
			return 0, err
		}

		c.PrintError(fmt.Sprintf("ERROR : %s\n", err))
		if cc, isHelpSupported := c.(HelpSupported); isHelpSupported {
			c.PrintError(fmt.Sprintf("Run '%s' for usage.\n", cc.FullHelpCommandName()))
		}

		return 2, err
	}
	return 0, nil
}

// RunCommand parses and executes the command.
func RunCommand(c Command, args []string) int {
	ret, err := parseCommand(c, args)
	if err != nil {
		return ret
	}
	return c.Execute()
}

// RunCommandContext parses and executes the command with context.
func RunCommandContext(ctx context.Context, c Command, args []string) int {
	ret, err := parseCommand(c, args)
	if err != nil {
		return ret
	}
	return c.ExecuteContext(ctx)
}
