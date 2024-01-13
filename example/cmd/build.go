package cmd

import (
	"fmt"

	fv "github.com/kmio11/flag-validator/pflag-validator"
	"github.com/kmio11/mycmd"
)

// BuildCommand is an example command which has flags and arguments.
type BuildCommand struct {
	*mycmd.Base

	flagOut    *string
	flagRace   *bool
	argPackage *string
}

func NewBuildCommand() *BuildCommand {
	cmd := &BuildCommand{
		Base: mycmd.NewBase(
			"build",
			mycmd.BaseConfig{
				ShortDescription: "compile packages and dependencies",
				ShortUsage:       "--out output [--race] <packages>",
			},
		),
	}

	// set flags
	cmd.flagOut = cmd.FS().StringP("out", "o", "", "write the resulting executable to the named output file")
	cmd.flagRace = cmd.FS().Bool("race", false, "enable data race detection")

	// set arguments
	cmd.argPackage = cmd.FS().ArgString(0, "packages", "the packages named by the import paths")

	// set validation rules
	cmd.FS().SetValidationRules(
		fv.Flag("out").Required(),
		fv.NumberOfArgs(1),
	)

	return cmd
}

func (c BuildCommand) Execute() int {
	c.Print(fmt.Sprintf(
		"Build successful. package=<%s> out=<%s>\n",
		*c.argPackage, *c.flagOut,
	))
	return 0
}
