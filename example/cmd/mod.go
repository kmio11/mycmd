package cmd

import (
	"fmt"

	fv "github.com/kmio11/flag-validator/pflag-validator"
	"github.com/kmio11/mycmd"
)

type (
	// ModCommand is an example command which has a sub command.
	ModCommand struct {
		*mycmd.ParentBase
	}

	// EditCommand is an sub command of ModCommand.
	EditCommand struct {
		*mycmd.Base

		flagFmt   *bool
		flagPrint *bool
		flagJSON  *bool
	}
)

func NewModCommand() *ModCommand {
	cmd := &ModCommand{
		ParentBase: mycmd.NewParentBase(
			"mod",
			mycmd.BaseConfig{
				ShortDescription: "provides access to operations on modules.",
			},
		).AddCommands(
			NewEditCommand(),
		),
	}

	return cmd
}

func NewEditCommand() *EditCommand {
	cmd := &EditCommand{
		Base: mycmd.NewBase(
			"edit",
			mycmd.BaseConfig{
				ShortDescription: "edit a file from tools or scripts",
				ShortUsage:       "[-fmt|-print|-json]",
			},
		),
	}

	// set flags
	cmd.flagFmt = cmd.FS().Bool("fmt", false, "reformats the file without making other changes")
	cmd.flagPrint = cmd.FS().Bool("print", false, "prints the file in its text format")
	cmd.flagJSON = cmd.FS().Bool("json", false, "prints the file in JSON format")

	// set validation rules
	cmd.FS().SetValidationRules(
		fv.NumberOfArgs(0),
		fv.MutuallyExclusive(
			fv.Flag("fmt"),
			fv.Flag("print"),
			fv.Flag("json"),
		),
	)

	return cmd
}

func (c EditCommand) Execute() int {
	if *c.flagFmt {
		c.Print(fmt.Sprintln("formatted!!"))
		return 0
	}
	if *c.flagPrint {
		c.Print(fmt.Sprintln("printed in text format!!"))
		return 0
	}
	if *c.flagJSON {
		c.Print(fmt.Sprintln("printed in JSON format!!"))
		return 0
	}

	c.Print(fmt.Sprintln("edited!!"))
	return 0
}
