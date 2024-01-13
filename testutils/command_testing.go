package testutils

import (
	"bytes"
	"flag"
	"fmt"
	"testing"

	"github.com/kmio11/mycmd"
	"github.com/stretchr/testify/assert"
)

var update = flag.Bool("update", false, "update golden files")

type (
	Factory func() mycmd.Command

	CustomAssertion[T any] func(
		t *testing.T,
		tt T,
		update bool,
		cmd mycmd.Command,
		actual []any,
	)

	SetupFunc[T any] func(t *testing.T, tt T)
)

func setup[T any](t *testing.T, tt T, newCmd Factory, setupFunc SetupFunc[T]) (cmd mycmd.Command, outWriter, errWriter *bytes.Buffer) {
	if setupFunc != nil {
		setupFunc(t, tt)
	}

	cmd = newCmd()

	outWriter, errWriter = new(bytes.Buffer), new(bytes.Buffer)
	cmd.SetOutWriter(outWriter)
	cmd.SetErrWriter(errWriter)

	return
}

func AssertStdOutAndStdErr(t *testing.T, testdata *TestData, ttName string, outWriter, errWriter *bytes.Buffer) {
	// assert stdout
	testdata.CompareWithGolden(t, *update,
		testdata.FileName(t,
			"golden", ttName, "stdout.txt",
		),
		outWriter.Bytes(),
	)

	// assert stderr
	testdata.CompareWithGolden(t, *update,
		testdata.FileName(t,
			"golden", ttName, "stderr.txt",
		),
		errWriter.Bytes(),
	)
}

type (
	TestCaseUsage struct {
		Name  string
		Setup SetupFunc[TestCaseUsage]
	}
)

func RunTestCommand_Usage(
	t *testing.T,
	tests []TestCaseUsage,
	testdata *TestData,
	newCmd Factory,
	customAssertion CustomAssertion[TestCaseUsage],
) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			cmd, outWriter, errWriter := setup[TestCaseUsage](t, tt, newCmd, tt.Setup)

			actual := cmd.Usage()

			//
			// assertion
			//

			// assert usage string
			testdata.CompareWithGolden(t, *update,
				testdata.FileName(t,
					"golden", tt.Name, "usage.txt",
				),
				[]byte(actual),
			)

			// assert stdout and stderr
			AssertStdOutAndStdErr(t, testdata, tt.Name, outWriter, errWriter)

			if customAssertion != nil {
				customAssertion(t, tt, *update, cmd, []any{actual})
			}
		})
	}
}

type (
	TestCaseParse struct {
		Name      string
		Args      []string
		WantError bool
		Setup     SetupFunc[TestCaseParse]
	}
)

func RunTestCommand_Parse(
	t *testing.T,
	tests []TestCaseParse,
	testdata *TestData,
	newCmd Factory,
	customAssertion CustomAssertion[TestCaseParse],
) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			cmd, outWriter, errWriter := setup[TestCaseParse](t, tt, newCmd, tt.Setup)

			actual := cmd.Parse(tt.Args)

			//
			// assertion
			//
			assert.Equal(
				t, tt.WantError, actual != nil,
				"wantError is %v, but actual was %v", tt.WantError, actual != nil,
			)

			// assert error message
			testdata.CompareWithGolden(t, *update,
				testdata.FileName(t,
					"golden", tt.Name, "error.txt",
				),
				[]byte(fmt.Sprintf("%v", actual)),
			)

			// assert stdout and stderr
			AssertStdOutAndStdErr(t, testdata, tt.Name, outWriter, errWriter)
		})
	}
}

type (
	TestCaseExecute struct {
		Name  string
		Args  []string
		Want  int
		Setup SetupFunc[TestCaseExecute]
	}
)

func RunTestCommand_Execute(
	t *testing.T,
	tests []TestCaseExecute,
	testdata *TestData,
	newCmd Factory,
	customAssertion CustomAssertion[TestCaseExecute],
) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			cmd, outWriter, errWriter := setup[TestCaseExecute](t, tt, newCmd, tt.Setup)

			err := cmd.Parse(tt.Args)
			if err != nil {
				t.Fatal(err)
			}

			actual := cmd.Execute()

			//
			// assertion
			//

			// assert status code
			assert.Equal(t, tt.Want, actual)

			// assert stdout and stderr
			AssertStdOutAndStdErr(t, testdata, tt.Name, outWriter, errWriter)

			if customAssertion != nil {
				customAssertion(t, tt, *update, cmd, []any{actual})
			}
		})
	}
}
