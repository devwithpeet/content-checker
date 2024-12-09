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
		wantTagsWanted    []string
		wantCheckExternal bool
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
			wantTagsWanted:    []string{},
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
			wantTagsWanted:    []string{},
		},
		{
			name:              "print hello",
			args:              []string{"", "print", "hello"},
			wantCommand:       PrintCommand,
			wantPath:          "hello",
			wantStatesAllowed: defaultStatesAllowed,
			wantVerbose:       false,
			wantPrintIndex:    false,
			wantPrintNonIndex: true,
			wantCourseWanted:  "",
			wantMaxErrors:     -1,
			wantTagsWanted:    []string{},
		},
		{
			name:              "print hello --verbose",
			args:              []string{"", "print", "hello", "--verbose"},
			wantCommand:       PrintCommand,
			wantPath:          "hello",
			wantStatesAllowed: defaultStatesAllowed,
			wantVerbose:       true,
			wantPrintIndex:    false,
			wantPrintNonIndex: true,
			wantCourseWanted:  "",
			wantMaxErrors:     -1,
			wantTagsWanted:    []string{},
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
			wantTagsWanted:    []string{},
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
			wantTagsWanted:    []string{},
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
			wantTagsWanted:    []string{},
		},
		{
			name:        "print . --verbose --max-errors 12 --tags 'foo,bar' stub a1.1",
			args:        []string{"", "print", ".", "--verbose", "--max-errors", "12", "--tags", "foo,bar", "stub", "a1.1"},
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
			wantTagsWanted:    []string{"foo", "bar"},
		},
		{
			name:              "check-links . --check-external",
			args:              []string{"", "check-links", ".", "--check-external"},
			wantCommand:       CheckLinksCommand,
			wantPath:          ".",
			wantStatesAllowed: defaultStatesAllowed,
			wantVerbose:       false,
			wantPrintIndex:    false,
			wantPrintNonIndex: true,
			wantCourseWanted:  "",
			wantMaxErrors:     -1,
			wantTagsWanted:    []string{},
			wantCheckExternal: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// execute
			command, path, statesAllowed, verbose, printIndex, printNonIndex, courseWanted, maxErrors, tagsWanted, checkExternal := getArgs(tt.args)

			// verify
			assert.Equal(t, tt.wantCommand, command, "command")
			assert.Equal(t, tt.wantPath, path, "path")
			assert.Equal(t, tt.wantStatesAllowed, statesAllowed, "statesAllowed")
			assert.Equal(t, tt.wantVerbose, verbose, "verbose")
			assert.Equal(t, tt.wantPrintIndex, printIndex, "printIndex")
			assert.Equal(t, tt.wantPrintNonIndex, printNonIndex, "printNonIndex")
			assert.Equal(t, tt.wantCourseWanted, courseWanted, "courseWanted")
			assert.Equal(t, tt.wantMaxErrors, maxErrors, "maxErrors")
			assert.Equal(t, tt.wantTagsWanted, tagsWanted, "tagsWanted")
			assert.Equal(t, tt.wantCheckExternal, checkExternal, "checkExternal")
		})
	}

}
