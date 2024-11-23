package main

import (
	"testing"

	"github.com/devwithpeet/content-checker/pkg"
	"github.com/stretchr/testify/assert"
)

// insert test for getArgs
func Test_getArgs(t *testing.T) {
	defaultStatesAllowed := map[pkg.State]struct{}{
		pkg.Complete:   {},
		pkg.Incomplete: {},
		pkg.Stub:       {},
	}

	tests := []struct {
		name              string
		args              []string
		wantCommand       Command
		wantPath          string
		wantStatesAllowed map[pkg.State]struct{}
		wantVerbose       bool
		wantPrintIndex    bool
		wantPrintNonIndex bool
		wantCourseWanted  string
		wantMaxErrors     int
	}{
		{
			name:              "version",
			args:              []string{"", "version"},
			wantCommand:       VersionCommand,
			wantPath:          ".",
			wantStatesAllowed: defaultStatesAllowed,
			wantVerbose:       false,
			wantPrintIndex:    false,
			wantPrintNonIndex: true,
			wantCourseWanted:  "",
			wantMaxErrors:     -1,
		},
		{
			name:              "print",
			args:              []string{"", "print"},
			wantCommand:       PrintCommand,
			wantPath:          ".",
			wantStatesAllowed: defaultStatesAllowed,
			wantVerbose:       false,
			wantPrintIndex:    false,
			wantPrintNonIndex: true,
			wantCourseWanted:  "",
			wantMaxErrors:     -1,
		},
		{
			name:              "print .",
			args:              []string{"", "print", "hello"},
			wantCommand:       PrintCommand,
			wantPath:          "hello",
			wantStatesAllowed: defaultStatesAllowed,
			wantVerbose:       false,
			wantPrintIndex:    false,
			wantPrintNonIndex: true,
			wantCourseWanted:  "",
			wantMaxErrors:     -1,
		},
		{
			name:              "print . --verbose",
			args:              []string{"", "print", ".", "--verbose"},
			wantCommand:       PrintCommand,
			wantPath:          ".",
			wantStatesAllowed: defaultStatesAllowed,
			wantVerbose:       true,
			wantPrintIndex:    false,
			wantPrintNonIndex: true,
			wantCourseWanted:  "",
			wantMaxErrors:     -1,
		},
		{
			name:              "print . --verbose --max-errors 12",
			args:              []string{"", "print", ".", "--verbose", "--max-errors", "12"},
			wantCommand:       PrintCommand,
			wantPath:          ".",
			wantStatesAllowed: defaultStatesAllowed,
			wantVerbose:       true,
			wantPrintIndex:    false,
			wantPrintNonIndex: true,
			wantCourseWanted:  "",
			wantMaxErrors:     12,
		},
		{
			name:              "print . --verbose --max-errors 12 a1.1",
			args:              []string{"", "print", ".", "--verbose", "--max-errors", "12", "a1.1"},
			wantCommand:       PrintCommand,
			wantPath:          ".",
			wantStatesAllowed: defaultStatesAllowed,
			wantVerbose:       true,
			wantPrintIndex:    false,
			wantPrintNonIndex: true,
			wantCourseWanted:  "a1.1",
			wantMaxErrors:     12,
		},
		{
			name:        "print . --verbose --max-errors 12 stub a1.1",
			args:        []string{"", "print", ".", "--verbose", "--max-errors", "12", "stub", "a1.1"},
			wantCommand: PrintCommand,
			wantPath:    ".",
			wantStatesAllowed: map[pkg.State]struct{}{
				pkg.Stub: {},
			},
			wantVerbose:       true,
			wantPrintIndex:    false,
			wantPrintNonIndex: true,
			wantCourseWanted:  "a1.1",
			wantMaxErrors:     12,
		},
		{
			name:        "print . --verbose --max-errors 12 stub a1.1",
			args:        []string{"", "print", ".", "--verbose", "--max-errors", "12", "stub", "a1.1"},
			wantCommand: PrintCommand,
			wantPath:    ".",
			wantStatesAllowed: map[pkg.State]struct{}{
				pkg.Stub: {},
			},
			wantVerbose:       true,
			wantPrintIndex:    false,
			wantPrintNonIndex: true,
			wantCourseWanted:  "a1.1",
			wantMaxErrors:     12,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// execute
			command, path, statesAllowed, verbose, printIndex, printNonIndex, courseWanted, maxErrors := getArgs(tt.args)

			// verify
			assert.Equal(t, tt.wantCommand, command)
			assert.Equal(t, tt.wantPath, path)
			assert.Equal(t, tt.wantStatesAllowed, statesAllowed)
			assert.Equal(t, tt.wantVerbose, verbose)
			assert.Equal(t, tt.wantPrintIndex, printIndex)
			assert.Equal(t, tt.wantPrintNonIndex, printNonIndex)
			assert.Equal(t, tt.wantCourseWanted, courseWanted)
			assert.Equal(t, tt.wantMaxErrors, maxErrors)
		})
	}

}
