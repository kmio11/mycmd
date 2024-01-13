package main

import (
	"testing"

	"github.com/kmio11/mycmd"
	"github.com/kmio11/mycmd/testutils"
)

func TestRoot_ParseAndExecute(t *testing.T) {
	testdata := testutils.NewTestData(t, t.Name())
	tests := []testutils.TestCaseRootParseAndExecute{
		{
			Name: "version",
			Args: []string{
				"version",
			},
			Want: 0,
		},
		{
			Name: "help_version",
			Args: []string{
				"help", "version",
			},
			Want: 0,
		},
		{
			Name: "build",
			Args: []string{
				"build", "--out", "output", "--race", "packages",
			},
			Want: 0,
		},
		{
			Name: "help_build",
			Args: []string{
				"help", "build",
			},
			Want: 0,
		},
		{
			Name: "mod_edit",
			Args: []string{
				"mod", "edit", "--fmt",
			},
			Want: 0,
		},
		{
			Name: "help_mod",
			Args: []string{
				"help", "mod",
			},
			Want: 0,
		},
		{
			Name: "mod_help_edit",
			Args: []string{
				"mod", "help", "edit",
			},
			Want: 0,
		},
	}

	testutils.RunTestRoot_ParseAndExecute(
		t, tests, testdata,
		func() mycmd.Command {
			return NewRootCommand()
		},
		nil,
	)
}
