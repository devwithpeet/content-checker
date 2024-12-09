package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/devwithpeet/content-checker/pkg"
	sm "github.com/peteraba/sortedmap"
)

type Command string

const Version = "0.7.1"

const (
	PrintCommand             Command = "print"
	ErrorsCommand            Command = "errors"
	StatsCommand             Command = "stats"
	VersionCommand           Command = "version"
	CheckPageOrderCommand    Command = "check-page-order"
	CheckChapterOrderCommand Command = "check-chapter-order"
	CheckLinksCommand        Command = "check-links"
)

func getArgs(args []string) (Command, string, map[pkg.State]struct{}, bool, bool, bool, string, int, []string, bool) {
	var err error

	action := PrintCommand
	if len(args) > 1 {
		action = Command(args[1])
	}

	statesAllowed := map[pkg.State]struct{}{}

	checkExternal := false
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
			case "--check-external", "-check-external":
				checkExternal = true
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

	return action, root, statesAllowed, verbose, printIndex, printNonIndex, courseWanted, maxErrors, tagsWanted, checkExternal
}

func main() {
	action, root, statesAllowed, verbose, printIndex, printNonIndex, courseWanted, maxErrors, tagsWanted, checkExternal := getArgs(os.Args)

	// collect markdown files
	files, err := findFiles(root, courseWanted, verbose)
	if err != nil {
		panic("cannot find files in root: " + root + ", error: " + err.Error())
	}

	if action == VersionCommand {
		fmt.Println("Version:", Version)

		return
	}

	// fetch markdown files
	courses, count := CrawlMarkdownFiles(files, maxErrors, tagsWanted, verbose)
	fmt.Println("Processed", count, "markdown files.")

	switch action {
	case PrintCommand:
		Print(count, courses, statesAllowed, printIndex, printNonIndex)

	case ErrorsCommand:
		Errors(count, courses)

	case StatsCommand:
		pkg.PrintStats(courses)

	case CheckChapterOrderCommand:
		CheckChapterOrder(count, courses)

	case CheckPageOrderCommand:
		CheckPageOrder(count, courses)

	case CheckLinksCommand:
		if courseWanted != "" {
			fmt.Println("cannot check links for a specific course")

			return
		}
		if len(tagsWanted) > 0 {
			fmt.Println("cannot check links for a specific tag")

			return
		}

		CheckLinks(count, courses, checkExternal, verbose)

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

		if len(rawContent) == 0 {
			panic("empty file: " + filePath)
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

func Print(count int, courses pkg.Courses, statesAllowed map[pkg.State]struct{}, printIndex, printNonIndex bool) {
	for _, course := range courses {
		fmt.Print(course.String(statesAllowed, printIndex, printNonIndex))
	}
}

func CheckChapterOrder(count int, courses pkg.Courses) {
	for _, course := range courses {
		issues := course.GetChapterOrderIssues()

		if len(issues) == 0 {
			continue
		}

		fmt.Println(course.Course)
		fmt.Println(strings.Repeat("=", len(course.Course)))
		fmt.Println(strings.Join(issues, "\n"))
		fmt.Println()
		fmt.Println(strings.Join(issues, "\n"))
		fmt.Println()
	}
}

func CheckPageOrder(count int, courses pkg.Courses) {
	for _, course := range courses {
		issues := course.GetPageOrderIssues()

		if len(issues) == 0 {
			continue
		}

		fmt.Println(course.Course)
		fmt.Println(strings.Repeat("=", len(course.Course)))
		fmt.Println()
		fmt.Println(strings.Join(issues, "\n"))
		fmt.Println()
	}
}

var linkRegex = regexp.MustCompile(`https?://([^//]+)/?[^ ]*`)

func CheckLinks(count int, courses pkg.Courses, checkExternal, verbose bool) {
	fileLinks := make(map[string][]string)
	internalLinks := sm.New[string, []string]()
	externalLinks := sm.New[string, *sm.SortedMap[string, []string]]()
	for _, course := range courses {
		for page, link := range course.GetLinks() {
			matches := linkRegex.FindStringSubmatch(link)

			if len(matches) < 2 {
				ext := filepath.Ext(link)

				if ext != "" {
					fileLinks[link] = append(fileLinks[link], page)
				} else if internalLinks.Has(link) {
					internalLinks.Set(link, append(internalLinks.MustGet(link), page))
				} else {
					internalLinks.Set(link, []string{page})
				}

				continue
			}

			domain := matches[1]

			if !externalLinks.Has(domain) {
				externalLinks.Set(domain, sm.New[string, []string]())
			}

			domainLinks := externalLinks.MustGet(domain)

			if !domainLinks.Has(link) {
				domainLinks.Set(link, []string{})
			}

			domainLinks.Set(link, append(domainLinks.MustGet(link), page))
		}
	}

	checkInternalLinks(internalLinks, courses, verbose)
	checkExternalLinks(externalLinks, checkExternal, verbose)
	checkFileLinks(fileLinks)
}

func checkInternalLinks(links *sm.SortedMap[string, []string], courses pkg.Courses, verbose bool) {
	validInternalLinks := courses.GetValidInternalLinks()

	notFound := 0
	for link, pages := range links.Items() {
		if _, ok := validInternalLinks[link]; ok {
			continue
		}

		if _, ok := validInternalLinks[link+"/"]; ok {
			continue
		}

		notFound++
		fmt.Printf("- '%s' NOT FOUND\n", link)
		for _, page := range pages {
			fmt.Printf("    - %s\n", page)
		}
	}

	if notFound > 0 {
		fmt.Println("Not found", notFound, "internal links.")
	} else {
		fmt.Println("All internal links found.")
	}

	if verbose {
		for link := range validInternalLinks {
			fmt.Printf("Found link: '%s'\n", link)
		}
	}
}

var skipDomains = []string{
	"codeforces.com",
	"developer.android.com",
	"leetcode.com",
	"linux.die.net",
	"marketplace.visualstudio.com",
	"udemy.com",
	"youtube.com",
	"www.amazon.com",
	"www.canva.com",
	"www.cloudflare.com",
	"www.java.com",
	"www.linux.org",
	"www.make.com",
	"www.mercurial-scm.org",
	"www.skillshare.com",
	"www.softwaretestinghelp.com",
	"www.whatsapp.com",
	"www.youtube.com",
}

func checkExternalLinks(links *sm.SortedMap[string, *sm.SortedMap[string, []string]], checkExternal, verbose bool) {
	if !checkExternal {
		return
	}

	var wg sync.WaitGroup
	var lock sync.Mutex

	for domain, domainLinks := range links.Items() {
		skip := false
		for _, skipDomain := range skipDomains {
			if domain == skipDomain {
				skip = true
				fmt.Println("Skipping domain:", domain)
			}
		}
		if skip {
			continue
		}

		wg.Add(1)

		written := false
		go func() {
			defer wg.Done()

			domainClient := pkg.NewDomainClient(domain, 5*time.Second, 3, 1*time.Second)

			results := domainClient.FetchPages(domainLinks.Keys())
			lock.Lock()
			defer lock.Unlock()

			for _, result := range results {
				if result.Code == 200 {
					continue
				}

				fmt.Printf("Domain: %s, Status code: %d, URL: %s\n", domain, result.Code, result.URL)
				for _, content := range domainLinks.MustGet(result.URL) {
					fmt.Println("  -", content)
					written = true
				}
			}

			if verbose {
				fmt.Printf("Domain: %s, Count: %d\n\n", domain, domainLinks.Len())
			}

			if verbose || written {
				fmt.Println()
			}
		}()
	}

	wg.Wait()
}

func checkFileLinks(links map[string][]string) {
	found := 0
	notFound := 0
	for link, pages := range links {
		filePath := "static/" + link
		if _, err := os.Stat(filePath); err == nil {
			found++

			continue
		}

		filePath = "content/" + link

		if _, err := os.Stat(filePath); err == nil {
			found++

			continue
		}

		notFound++

		fmt.Printf("- %s NOT FOUND\n", link)
		for _, page := range pages {
			fmt.Printf("    - %s\n", page)
		}
	}

	if found > 0 {
		fmt.Println("Found", found, "file links.")
	}
	if notFound > 0 {
		fmt.Println("Not found", notFound, "file links.")
	}
}

func Errors(count int, courses pkg.Courses) {

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
