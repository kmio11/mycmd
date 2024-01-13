package mycmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
)

var _ SubCommand = (*Help)(nil)

type Help struct {
	parent    Command
	outWriter io.Writer
	errWriter io.Writer

	target        Command
	unknownTarget string
}

func NewHelp(parent ParentCommand) *Help {
	return &Help{
		parent:    parent,
		outWriter: os.Stdout,
		errWriter: os.Stderr,
	}
}

func (c *Help) Name() string {
	return "help"
}

func (c *Help) ShortDescription() string {
	return ""
}

func (c *Help) Usage() string {
	return ""
}

func (c *Help) Parse(args []string) error {
	if len(args) == 0 {
		// subcommand is not specified. help for parent
		c.target = c.parent
		return nil
	}
	if parent, ok := c.parent.(ParentCommand); ok {
		for _, sub := range parent.Commands() {
			sub := sub
			if args[0] == sub.Name() {
				// help for subcommand
				c.target = sub
				break
			}
		}
	}

	c.unknownTarget = args[0]
	return nil
}

func (c *Help) IsHelpRequested(err error) bool {
	return false
}

func (c *Help) Execute() int {
	ctx := context.Background()
	return c.ExecuteContext(ctx)
}

func (c *Help) ExecuteContext(ctx context.Context) int {
	if c.target != nil {
		fmt.Fprintln(c.outWriter, c.target.Usage())
		return 0
	}
	fullName := strings.Join(FullName(c), " ")
	fmt.Fprintf(
		c.errWriter, "%s %s: unknown help topic. Run '%s'.\n",
		fullName, c.unknownTarget, fullName,
	)
	return 2
}

func (c *Help) Parent() Command {
	return c.parent
}

func (c *Help) SetParent(parent Command) {
	c.parent = parent
}

func (c *Help) SetOutWriter(w io.Writer) {
	c.outWriter = w
}

func (c *Help) SetErrWriter(w io.Writer) {
	c.errWriter = w
}

func (c *Help) Print(msg string) {
	fmt.Fprint(c.outWriter, msg)
}

func (c *Help) PrintError(msg string) {
	fmt.Fprint(c.errWriter, msg)
}
