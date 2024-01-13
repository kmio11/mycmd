package mycmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/kmio11/mycmd/wflag"
	"github.com/spf13/pflag"
)

var _ interface {
	SubCommand
	HiddenSupported
} = (*Base)(nil)

type Base struct {
	fs               *wflag.FlagSet
	name             string
	shortDescription string
	shortUsage       string
	hidden           bool
	outWriter        io.Writer
	errWriter        io.Writer
	parent           Command
}

type BaseConfig struct {
	ShortDescription string
	ShortUsage       string
	Hidden           bool
}

func NewBase(name string, cfg BaseConfig) *Base {
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

func (c Base) Name() string {
	return c.name
}

func (c Base) ShortDescription() string {
	return c.shortDescription
}

const baseUsageTemplate = `
Usage:

{{.CommandNameAndFlags}}
{{if ne .Flags ""}}
Flags:

{{.Flags}}
{{- printf "\n"}}
{{- end -}}
{{if ne .Arguments ""}}
Arguments:

{{.Arguments}}
{{- printf "\n"}}
{{- end -}}
`

func (c *Base) commandNameAndFlags(identNum int) string {
	ident := strings.Repeat(" ", identNum)
	shortUsage := strings.ReplaceAll(
		strings.ReplaceAll(c.shortUsage, "\t", ""),
		"\n", fmt.Sprintf("\n%s", ident),
	)
	return fmt.Sprintf("%s%s %s", ident, strings.Join(FullName(c), " "), shortUsage)
}

func (c *Base) Usage() string {
	usageData := map[string]any{
		"CommandNameAndFlags": c.commandNameAndFlags(2),
		"Flags":               strings.TrimRight(c.FS().FlagUsages(), "\n"),
		"Arguments":           strings.TrimRight(c.FS().ArgUsages(), "\n"),
	}

	tmpl := template.Must(template.New("BaseUsage").Parse(baseUsageTemplate))
	var buf bytes.Buffer
	tmpl.Execute(&buf, usageData)

	return buf.String()
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

// FS returns FlagSet
func (c Base) FS() *wflag.FlagSet {
	return c.fs
}

// Parse parses the flags
func (c Base) Parse(args []string) error {
	err := c.fs.Parse(args)
	if err != nil {
		return err
	}
	return nil
}

// IsHelpRequested will return true when the help was requested in Parse.
func (c Base) IsHelpRequested(err error) bool {
	return errors.Is(err, wflag.ErrHelp)
}

// Execute executes the main processing of the command.
func (c Base) Execute() int {
	panic(fmt.Errorf("%s is not implemented", c.name))
}

// ExecuteContext is the same as Execute() but accept a context as an argument.
func (c Base) ExecuteContext(ctx context.Context) int {
	panic(fmt.Errorf("%s is not implemented", c.name))
}

// Print writes to OutWriter.
func (c *Base) Print(msg string) {
	fmt.Fprint(c.outWriter, msg)
}

// PrintError writes to ErrWriter.
func (c *Base) PrintError(msg string) {
	fmt.Fprint(c.errWriter, msg)
}

// OutWriter returns the standard output writer.
func (c *Base) OutWriter() io.Writer {
	return c.outWriter
}

// SetOutWriter sets the standard output writer.
func (c *Base) SetOutWriter(w io.Writer) {
	c.outWriter = w
}

// ErrWriter returns the error output writer.
func (c *Base) ErrWriter() io.Writer {
	return c.errWriter
}

// SetErrWriter sets the error output writer.
func (c *Base) SetErrWriter(w io.Writer) {
	c.errWriter = w
}

// Parent returns the parent command.
func (c *Base) Parent() Command {
	return c.parent
}

// SetParent sets the parent command.
func (c *Base) SetParent(parent Command) {
	c.parent = parent
}

// If Hidden is true, this command will not be displayed in Usage.
func (c *Base) Hidden() bool {
	return c.hidden
}
