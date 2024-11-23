package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/devwithpeet/content-checker/pkg"
)

type Command string

const Version = "0.1.11"

const (
	PrintCommand   Command = "print"
	ErrorsCommand  Command = "errors"
	StatsCommand   Command = "stats"
	VersionCommand Command = "version"
)

func main() {
	action := PrintCommand
	if len(os.Args) > 1 {
		action = Command(os.Args[1])
	}

	root := "."
	if len(os.Args) > 2 {
		root = os.Args[2]
	}

	// collect markdown files
	files, err := findFiles(root)
	if err != nil {
		panic("cannot find files in root: " + root + ", error: " + err.Error())
	}

	// fetch markdown files
	courses, count := CrawlMarkdownFiles(files, maxErrors)

	Prepare(courses)

	switch action {
	case VersionCommand:
		fmt.Println("Version:", Version)

	case PrintCommand:
		statesAllowed := map[pkg.State]struct{}{
			pkg.Complete:   {},
			pkg.Incomplete: {},
			pkg.Stub:       {},
		}

		printIndex := false
		printNonIndex := true

		if len(os.Args) > 3 {
			for _, arg := range os.Args[3:] {
				switch arg {
				case "--without-non-index":
					printNonIndex = false
				case "--with-index":
					printIndex = true
				case "complete":
					statesAllowed = map[pkg.State]struct{}{
						pkg.Complete: {},
					}
				case "incomplete":
					statesAllowed = map[pkg.State]struct{}{
						pkg.Incomplete: {},
					}
				case "stub":
					statesAllowed = map[pkg.State]struct{}{
						pkg.Stub: {},
					}
				}
			}
		}

		Print(count, courses, statesAllowed, printIndex, printNonIndex)

	case ErrorsCommand:
		Errors(count, courses)

	case StatsCommand:
		courses.Stats()

	default:
		panic("unknown command: " + string(action))
	}
}

func findFiles(root string) ([]string, error) {
	pattern := filepath.Join(root, "content") + "/**/**/*.md"

	return filepath.Glob(pattern)
}

const maxErrors = 10

func CrawlMarkdownFiles(matches []string, maxErrors int) (pkg.Courses, int) {
	if maxErrors < 0 {
		maxErrors = math.MaxInt
	}

	result := make(pkg.Courses, 0, len(matches))

	var count, errCount int

	for _, filePath := range matches {
		if errCount >= maxErrors {
			break
		}

		parts := strings.Split(filePath, "/")

		if len(parts) < 3 {
			fmt.Println("Skipping:", filePath)
			continue
		}

		course := parts[len(parts)-3]
		chapter := parts[len(parts)-2]
		page := parts[len(parts)-1]

		rawContent, err := os.ReadFile(filePath)
		if err != nil {
			panic("cannot open file: " + filePath)
		}

		content, err := pkg.ParseMarkdown(string(rawContent))
		if err != nil {
			panic("cannot parse markdown: " + filePath + ", err: " + err.Error())
		}

		result = result.Add(filePath, course, chapter, page, content)

		if len(content.GetIssues(filePath)) > 0 {
			errCount++
		}

		count++
	}

	return result, count
}

func Prepare(courses pkg.Courses) {
	for _, course := range courses {
		course.Prepare()
	}
}

func Print(count int, courses pkg.Courses, statesAllowed map[pkg.State]struct{}, printIndex, printNonIndex bool) {
	fmt.Println("Processed", count, "markdown files")

	for _, course := range courses {
		fmt.Print(course.String(statesAllowed, printIndex, printNonIndex))
	}
}

func Errors(count int, courses pkg.Courses) {
	fmt.Println("Processed", count, "markdown files")

	errorsFound := false

	for _, course := range courses {
		errors := course.GetErrors()
		if len(errors) == 0 {
			continue
		}

		errorsFound = true

		fmt.Println(strings.Join(errors, "\n"))
	}

	if errorsFound {
		os.Exit(1)
	}
}
