package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/devwithpeet/content-checker/pkg"
)

type Command string

const Version = "0.3.4"

const (
	PrintCommand   Command = "print"
	ErrorsCommand  Command = "errors"
	StatsCommand   Command = "stats"
	VersionCommand Command = "version"
)

func getArgs(args []string) (Command, string, map[pkg.State]struct{}, bool, bool, bool, string, int, []string) {
	var err error

	action := PrintCommand
	if len(args) > 1 {
		action = Command(args[1])
	}

	statesAllowed := map[pkg.State]struct{}{}

	verbose := false
	printIndex := false
	printNonIndex := true
	rootFound := false
	root := "."
	courseWanted := ""
	maxErrors := defaulMaxErrors
	tagsWanted := []string{}

	if len(args) > 2 {
		for i := 2; i < len(args); i++ {
			arg := args[i]

			switch arg {
			case "--without-non-index", "-without-non-index":
				printNonIndex = false
			case "--with-index", "-with-index":
				printIndex = true
			case "--verbose", "-verbose":
				verbose = true
			case "--tags", "-tags":
				if len(args) <= i+1 {
					panic("missing value for --tags")
				}

				for _, tag := range strings.Split(args[i+1], ",") {
					tagsWanted = append(tagsWanted, strings.TrimSpace(tag))
				}

				i++
			case "--max-errors", "-max-errors":
				if len(args) <= i+1 {
					panic("missing value for --max-errors")
				}

				maxErrors, err = strconv.Atoi(args[i+1])
				if err != nil {
					panic(err)
				}

				i++
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

			default:
				if !rootFound {
					root = arg
					rootFound = true
				} else {
					courseWanted = arg
				}
			}
		}
	}

	if len(statesAllowed) == 0 {
		statesAllowed = map[pkg.State]struct{}{
			pkg.Complete:   {},
			pkg.Incomplete: {},
			pkg.Stub:       {},
		}
	}

	return action, root, statesAllowed, verbose, printIndex, printNonIndex, courseWanted, maxErrors, tagsWanted
}

func main() {
	action, root, statesAllowed, verbose, printIndex, printNonIndex, courseWanted, maxErrors, tagsWanted := getArgs(os.Args)

	// collect markdown files
	files, err := findFiles(root, courseWanted, verbose)
	if err != nil {
		panic("cannot find files in root: " + root + ", error: " + err.Error())
	}

	// fetch markdown files
	courses, count := CrawlMarkdownFiles(files, maxErrors, tagsWanted, verbose)

	Prepare(courses)

	switch action {
	case VersionCommand:
		fmt.Println("Version:", Version)

	case PrintCommand:
		Print(count, courses, statesAllowed, printIndex, printNonIndex)

	case ErrorsCommand:
		Errors(count, courses)

	case StatsCommand:
		courses.Stats()

	default:
		panic("unknown command: " + string(action))
	}
}

func findFiles(root, courseWanted string, verbose bool) ([]string, error) {
	if courseWanted == "" {
		courseWanted = "**"
	}

	pattern := filepath.Join(root, "content") + "/" + courseWanted + "/**/*.md"

	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	if verbose {
		fmt.Println("Files found:")
		for _, file := range files {
			fmt.Println(file)
		}
	}

	return files, nil
}

const defaulMaxErrors = -1

func CrawlMarkdownFiles(matches []string, maxErrors int, tagsWanted []string, verbose bool) (pkg.Courses, int) {
	if maxErrors < 0 {
		maxErrors = math.MaxInt
	}

	result := make(pkg.Courses, 0, len(matches))

	var count, errCount int

	for _, filePath := range matches {
		if maxErrors > 0 && errCount >= maxErrors {
			fmt.Println("Max errors reached, stopping")
			break
		}

		parts := strings.Split(filePath, "/")

		if len(parts) < 3 {
			fmt.Println("Skipping:", filePath)
			continue
		}

		course := parts[len(parts)-3]
		chapter := parts[len(parts)-2]
		fileName := parts[len(parts)-1]

		rawContent, err := os.ReadFile(filePath)
		if err != nil {
			panic("cannot open file: " + filePath)
		}

		content, err := pkg.ParseMarkdown(string(rawContent))
		if err != nil {
			panic("cannot parse markdown: " + filePath + ", err: " + err.Error())
		}

		if len(tagsWanted) > 0 {
			found := false

			for _, tag := range content.Tags {
				for _, tagWanted := range tagsWanted {
					if tag == tagWanted {
						found = true
						break
					}

				}

				if found {
					break
				}
			}

			if !found {
				continue
			}
		}

		result = result.Add(filePath, course, chapter, fileName, content)

		if len(content.GetIssues(filePath, course, chapter, fileName)) > 0 {
			errCount++
		}

		count++
	}

	if verbose {
		fmt.Println()
		fmt.Println("Courses:")
		for _, course := range result {
			fmt.Println(course.Course)
		}
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
