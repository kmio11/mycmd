package mycmd

import "context"

var _ ParentCommand = (*Root)(nil)

type Root struct {
	*ParentBase
}

func NewRoot(name string) *Root {
	c := &Root{
		ParentBase: NewParentBase(name, BaseConfig{}),
	}
	return c
}

func (c *Root) AddCommands(commands ...Command) *Root {
	c.ParentBase = c.ParentBase.AddCommands(commands...)
	return c
}

// ParseAndExecute parses and executes command.
func (c *Root) ParseAndExecute(args []string) int {
	return RunCommand(c, args)
}

// ParseAndExecuteContext parses and executes command with context.
func (c *Root) ParseAndExecuteContext(ctx context.Context, args []string) int {
	return RunCommandContext(ctx, c, args)
}
