package mycmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"text/template"
)

var _ interface {
	ParentCommand
	HelpSupported
	HiddenSupported
} = (*ParentBase)(nil)

type ParentBase struct {
	*Base
	commands      []Command
	help          *Help
	parsedCommand Command
}

func NewParentBase(name string, cfg BaseConfig) *ParentBase {
	c := &ParentBase{
		Base: NewBase(name, cfg),
	}
	c.help = NewHelp(c)
	return c
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

const parentBaseUsageTemplate = `
Usage:

  {{.FullName}} {{.ShortUsage}}
{{ if gt (len .Commands) 0}}
Commands:
{{ range .Commands}}
  {{.}}
{{- end}}
{{- end}}

Use '{{.FullName}} {{.Help}} <command>' for more details on a command.
`

func (c *ParentBase) commandsWithShortDescription() []string {
	const adjuster = "!#!"
	maxCmdNameLen := 0
	lines := []string{}
	for _, sub := range c.commands {
		if v, ok := sub.(HiddenSupported); ok && v.Hidden() {
			continue
		}

		if maxCmdNameLen < len(sub.Name()) {
			maxCmdNameLen = len(sub.Name())
		}
		lines = append(lines, fmt.Sprintf("%s%s%s", sub.Name(), adjuster, sub.ShortDescription()))
	}

	alignedLines := []string{}
	for _, line := range lines {
		spacing := strings.Repeat(" ", maxCmdNameLen-strings.Index(line, adjuster)+3)
		alignedLines = append(alignedLines,
			strings.Replace(line, adjuster, spacing, 1),
		)
	}

	return alignedLines
}

func (c *ParentBase) Usage() string {
	usageData := map[string]any{
		"FullName":   strings.Join(FullName(c), " "),
		"ShortUsage": "<command> [flags] [arguments]",
		"Commands":   c.commandsWithShortDescription(),
		"Help":       c.help.Name(),
	}

	tmpl := template.Must(template.New("ParentBaseUsage").Parse(parentBaseUsageTemplate))
	var buf bytes.Buffer
	tmpl.Execute(&buf, usageData)

	return buf.String()
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

// Execute executes the main processing of the command.
func (c *ParentBase) Execute() int {
	return c.parsedCommand.Execute()
}

// ExecuteContext is the same as Execute() but accept a context as an argument.
func (c *ParentBase) ExecuteContext(ctx context.Context) int {
	return c.parsedCommand.ExecuteContext(ctx)
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

// SetOutWriter sets the standard output writer.
func (c *ParentBase) SetOutWriter(w io.Writer) {
	c.Base.SetOutWriter(w)
	c.passWriterToSubCmds()
}

// SetErrWriter sets the error output writer.
func (c *ParentBase) SetErrWriter(w io.Writer) {
	c.Base.SetErrWriter(w)
	c.passWriterToSubCmds()
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
