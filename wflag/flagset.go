package wflag

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	fv "github.com/kmio11/flag-validator/pflag-validator"

	flag "github.com/spf13/pflag"
)

// ErrHelp is the error returned if the flag help is invoked but no such flag is defined.
var ErrHelp = errors.New("help requested")

type FlagSet struct {
	*flag.FlagSet
	args            map[int]Arg
	validationRules *fv.RuleSet

	errorHandling flag.ErrorHandling
}

// Arg represents non-flag arguments.
type Arg struct {
	Index int
	Name  string
	Usage string
	Value *string
}

func NewFlagSet(name string, errorHandling flag.ErrorHandling) *FlagSet {
	fs := flag.NewFlagSet(name, errorHandling)
	// prevent to print to stderr
	fs.Usage = func() {}
	fs.SetOutput(io.Discard)

	return &FlagSet{
		FlagSet:       fs,
		args:          map[int]Arg{},
		errorHandling: errorHandling,
	}
}

func (fs *FlagSet) FlagUsages() string {
	return fs.FlagSet.FlagUsages()
}

func (fs *FlagSet) ArgUsages() string {
	const (
		indentNum  = 2
		spacingNum = 3
		adjuster   = "\x00"
	)
	lines := []string{}
	var maxNameLen int
	for _, arg := range fs.args {
		if len(arg.Name) > maxNameLen {
			maxNameLen = len(arg.Name)
		}
		lines = append(lines,
			fmt.Sprintf("%s%s%s",
				arg.Name, adjuster, arg.Usage,
			),
		)
	}

	var buf = new(bytes.Buffer)
	indent := strings.Repeat(" ", indentNum)
	for _, line := range lines {
		adjusted := strings.Replace(line,
			adjuster,
			strings.Repeat(" ", maxNameLen-strings.Index(line, adjuster)+spacingNum), 1,
		)
		wrapped := strings.ReplaceAll(adjusted,
			"\n",
			"\n"+strings.Repeat(" ", indentNum+maxNameLen+spacingNum),
		)

		fmt.Fprintf(buf, "%s%s\n", indent, wrapped)
	}

	return buf.String()
}

func (fs *FlagSet) Parse(args []string) error {
	err := fs.FlagSet.Parse(args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ErrHelp
		}
		return err
	}
	for i, a := range fs.args {
		if i > fs.NArg()-1 {
			*a.Value = ""
			continue
		}
		iArg := fs.Arg(i)
		*a.Value = iArg
	}
	if fs.validationRules != nil {
		err = fs.validationRules.Validate(fs.FlagSet)
		if err != nil {
			return fs.handleParsingError(err)
		}
	}

	return nil
}

func (fs *FlagSet) handleParsingError(err error) error {
	if err != nil {
		switch fs.errorHandling {
		case flag.ContinueOnError:
			return err
		case flag.ExitOnError:
			fmt.Println(err)
			os.Exit(2)
		case flag.PanicOnError:
			panic(err)
		}
	}
	return nil
}

func (fs *FlagSet) SetValidationRules(rules ...fv.Rule) {
	fs.validationRules = fv.NewRuleSet(rules...)
}

// ArgString returns pointer to set n'th non-flag argument after parsing.
func (fs *FlagSet) ArgString(n int, name string, usage string) *string {
	p := new(string)
	fs.args[n] = Arg{
		Name:  name,
		Usage: usage,
		Index: n,
		Value: p,
	}
	return p
}
