package mycmd

import (
	"fmt"
)

var _ ParentCommand = (*Root)(nil)

type Root struct {
	*ParentBase
}

func NewRoot(name string) *Root {
	c := &Root{
		ParentBase: NewParentBase(name, Config{}),
	}
	return c
}

func (c *Root) AddCommands(commands ...Command) *Root {
	c.ParentBase = c.ParentBase.AddCommands(commands...)
	return c
}

// ParseAndExecute parses and executes command.
func (c *Root) ParseAndExecute(args []string) int {
	err := c.Parse(args)
	if err != nil {
		if c.IsHelpRequested(err) {
			fmt.Fprintln(c.outWriter, c.Usage())
			return 0
		}

		fmt.Fprintf(c.errWriter, "ERROR : %s\n", err)
		fmt.Fprintf(c.errWriter, "Run '%s' for usage.\n", c.FullHelpCommandName())

		return 2
	}

	return c.Execute()
}
