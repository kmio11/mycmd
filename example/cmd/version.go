package cmd

import (
	"fmt"

	fv "github.com/kmio11/flag-validator/pflag-validator"
	"github.com/kmio11/mycmd"
)

// VersionCommand is an example command which has neither flags or arguments.
type VersionCommand struct {
	*mycmd.Base
}

func NewVersionCmd() *VersionCommand {
	cmd := &VersionCommand{
		Base: mycmd.NewBase(
			"version",
			mycmd.BaseConfig{
				ShortDescription: "print version",
				ShortUsage:       "",
			},
		),
	}

	// set validation rules
	cmd.FS().SetValidationRules(
		fv.NumberOfArgs(0),
	)

	return cmd
}

func (c VersionCommand) Execute() int {
	c.Print(fmt.Sprintln("v1.0.0"))
	return 0
}
