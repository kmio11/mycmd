package mycmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kmio11/mycmd/wflag"

	"github.com/spf13/pflag"
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

var _ interface {
	SubCommand
	HiddenSupported
} = (*Base)(nil)

var _ interface {
	ParentCommand
	HelpSupported
	HiddenSupported
} = (*ParentBase)(nil)

type (
	Base struct {
		fs               *wflag.FlagSet
		name             string
		shortDescription string
		shortUsage       string
		hidden           bool
		outWriter        io.Writer
		errWriter        io.Writer
		parent           Command
	}

	ParentBase struct {
		*Base
		shortUsage    string
		commands      []Command
		help          *Help
		parsedCommand Command
	}

	Config struct {
		ShortDescription string
		ShortUsage       string
		Hidden           bool
	}
)

func NewBase(name string, cfg Config) *Base {
	s := &Base{
		fs:               wflag.NewFlagSet(name, pflag.ContinueOnError),
		name:             name,
		shortDescription: cfg.ShortDescription,
		shortUsage:       cfg.ShortUsage,
		hidden:           cfg.Hidden,

		outWriter: os.Stdout,
		errWriter: os.Stderr,
	}
	s.fs.SortFlags = false

	return s
}

func NewParentBase(name string, cfg Config) *ParentBase {
	c := &ParentBase{
		Base:       NewBase(name, cfg),
		shortUsage: "<command> [flags] [arguments]",
	}
	c.help = NewHelp(c)
	return c
}

// // FullName returns full name of command (rootName subName subName ...)
func FullName(c Command) []string {
	fullName := []string{}
	if v, ok := c.(SubCommand); ok {
		parent := v.Parent()
		if parent != nil {
			fullName = FullName(parent)
		}
	}
	fullName = append(fullName, c.Name())
	return fullName
}

// AddCommands add subcommands
func (c *ParentBase) AddCommands(commands ...Command) *ParentBase {
	c.commands = commands
	for _, command := range commands {
		if sub, ok := command.(SubCommand); ok {
			sub.SetParent(c)
		}
		command.SetOutWriter(c.outWriter)
		command.SetErrWriter(c.errWriter)
	}

	return c
}

// FS returns FlagSet
func (c Base) FS() *wflag.FlagSet {
	return c.fs
}

func (c Base) Name() string {
	return c.name
}

func (c Base) ShortDescription() string {
	return c.shortDescription
}

func (c *Base) commandNameAndFlags(identNum int) string {
	ident := strings.Repeat(" ", identNum)
	shortUsage := strings.ReplaceAll(
		strings.ReplaceAll(c.shortUsage, "\t", ""),
		"\n", fmt.Sprintf("\n%s", ident),
	)
	return fmt.Sprintf("%s%s %s", ident, strings.Join(FullName(c), " "), shortUsage)
}

func (c *Base) Usage() string {
	var buf = new(bytes.Buffer)

	fmt.Fprintln(buf)
	fmt.Fprintln(buf, "Usage:")
	fmt.Fprintln(buf)
	fmt.Fprintf(buf, "%s\n", c.commandNameAndFlags(2))

	flagUsages := c.FS().FlagUsages()
	if flagUsages != "" {
		fmt.Fprintln(buf)
		fmt.Fprintln(buf, "Flags:")
		fmt.Fprintln(buf)
		fmt.Fprint(buf, flagUsages) // flagUsage has a empty line at the final line. So not using Fprintln but Fprint.
	}

	argUsages := c.FS().ArgUsages()
	if argUsages != "" {
		fmt.Fprintln(buf)
		fmt.Fprintln(buf, "Arguments:")
		fmt.Fprintln(buf)
		fmt.Fprint(buf, argUsages)
	}

	return buf.String()
}

func (c *ParentBase) Usage() string {
	fullName := strings.Join(FullName(c), " ")
	var buf = new(bytes.Buffer)

	fmt.Fprintln(buf)
	fmt.Fprintln(buf, "Usage:")
	fmt.Fprintln(buf)
	fmt.Fprintf(buf, "  %s %s\n", fullName, c.shortUsage)
	fmt.Fprintln(buf)
	fmt.Fprintln(buf, "Commands:")
	fmt.Fprintln(buf)

	const adjuster = "!#!"
	lines := []string{}
	maxCmdNameLen := 0
	for _, sub := range c.commands {
		if v, ok := sub.(HiddenSupported); ok && v.Hidden() {
			continue
		}

		if maxCmdNameLen < len(sub.Name()) {
			maxCmdNameLen = len(sub.Name())
		}
		lines = append(lines, fmt.Sprintf("%s%s%s", sub.Name(), adjuster, sub.ShortDescription()))
	}
	for _, line := range lines {
		spacing := strings.Repeat(" ", maxCmdNameLen-strings.Index(line, adjuster)+3)
		fmt.Fprintf(buf, "  %s\n", strings.Replace(line, adjuster, spacing, 1))
	}

	fmt.Fprintln(buf)
	fmt.Fprintf(buf, "Use '%s %s <command>' for more details on a command.\n", fullName, c.help.Name())

	return buf.String()
}

// Parse parses the flags
func (c Base) Parse(args []string) error {
	err := c.fs.Parse(args)
	if err != nil {
		return err
	}
	return nil
}

// Parse parses the flags
func (c *ParentBase) Parse(args []string) error {
	if len(args) == 0 {
		c.help.Parse(args)
		c.parsedCommand = c.help
		return nil
	}

	subcommand := args[0]
	if subcommand == c.help.Name() {
		c.help.Parse(args[1:])
		c.parsedCommand = c.help
		return nil
	}

	for _, sub := range c.commands {
		sub := sub
		if subcommand == sub.Name() {
			c.parsedCommand = sub
			break
		}
	}

	if c.parsedCommand != nil {
		err := c.parsedCommand.Parse(args[1:])
		if err != nil {
			if c.parsedCommand.IsHelpRequested(err) {
				c.help.Parse(args)
				c.parsedCommand = c.help
				return nil
			}
			return err
		}
		return nil
	}

	return fmt.Errorf("unknown command (%s)", subcommand)
}

// IsHelpRequested will return true when the help was requested in Parse.
func (c Base) IsHelpRequested(err error) bool {
	return errors.Is(err, wflag.ErrHelp)
}

// Execute executes the main processing of the command.
func (c Base) Execute() int {
	panic(fmt.Errorf("%s is not implemented", c.name))
}

// Execute executes the main processing of the command.
func (c *ParentBase) Execute() int {
	return c.parsedCommand.Execute()
}

// ExecuteContext is the same as Execute() but accept a context as an argument.
func (c Base) ExecuteContext(ctx context.Context) int {
	panic(fmt.Errorf("%s is not implemented", c.name))
}

// ExecuteContext is the same as Execute() but accept a context as an argument.
func (c *ParentBase) ExecuteContext(ctx context.Context) int {
	return c.parsedCommand.ExecuteContext(ctx)
}

// OutWriter returns the standard output writer.
func (c *Base) OutWriter() io.Writer {
	return c.outWriter
}

func (c *Base) SetOutWriter(w io.Writer) {
	c.outWriter = w
}

func (c *ParentBase) SetOutWriter(w io.Writer) {
	c.Base.SetOutWriter(w)
	c.passWriterToSubCmds()
}

// OutWriter returns the error output writer.
func (c *Base) ErrWriter() io.Writer {
	return c.errWriter
}

func (c *Base) SetErrWriter(w io.Writer) {
	c.errWriter = w
}

func (c *ParentBase) SetErrWriter(w io.Writer) {
	c.Base.SetErrWriter(w)
	c.passWriterToSubCmds()
}

// passWriterToSubCmds sets the same Writer as itself to its subcommands
func (c *ParentBase) passWriterToSubCmds() {
	if c.help != nil {
		c.help.SetOutWriter(c.outWriter)
		c.help.SetErrWriter(c.errWriter)
	}
	for _, s := range c.commands {
		s.SetOutWriter(c.outWriter)
		s.SetErrWriter(c.errWriter)
	}
}

func (c *Base) Parent() Command {
	return c.parent
}

func (c *Base) SetParent(parent Command) {
	c.parent = parent
}

// If Hidden is true, this command will not be displayed in Usage.
func (c *Base) Hidden() bool {
	return c.hidden
}

// Commands returns subcommands.
func (c *ParentBase) Commands() []Command {
	return c.commands
}

// FullHelpCommandName returns full name of help command (rootName helpName subName)
func (c *ParentBase) FullHelpCommandName() string {
	if c.parsedCommand != nil {
		if v, ok := c.parsedCommand.(HelpSupported); ok {
			childHelpMessage := v.FullHelpCommandName()
			if childHelpMessage != "" {
				return childHelpMessage
			}
		}
		if c.help != nil {
			return strings.Join(append(FullName(c.help), c.parsedCommand.Name()), " ")
		}
	}
	if c.help != nil {
		return strings.Join(append(FullName(c), c.help.Name()), " ")
	}
	return ""
}

func (c *Base) PrintError(msg string) {
	fmt.Fprintf(c.errWriter, "ERROR: %s\n", msg)
}
