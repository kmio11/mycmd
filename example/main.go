package main

import (
	"os"

	"github.com/kmio11/mycmd"
	"github.com/kmio11/mycmd/example/cmd"
)

func main() {
	rootCmd := NewRootCommand()
	os.Exit(rootCmd.ParseAndExecute(os.Args[1:]))
}

func NewRootCommand() *mycmd.Root {
	return mycmd.NewRoot("example").AddCommands(
		cmd.NewVersionCmd(),
		cmd.NewBuildCommand(),
		cmd.NewModCommand(),
	)
}
