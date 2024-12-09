package pkg

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gosimple/slug"
)

type State string

const (
	Incomplete State = "incomplete"
	Complete   State = "complete"
	Stub       State = "stub"
)

type Badge string

const (
	Alternative Badge = "alternative"
	Extra       Badge = "extra"
	Fun         Badge = "fun"
	Hint        Badge = "hint"
	MustSee     Badge = "must-see"
	FullCourse  Badge = "full-course"
	DeepDive    Badge = "deep-dive"
	Summary     Badge = "summary"
	Unchecked   Badge = "unchecked"
	NoEmbed     Badge = "no-embed"
	Audio       Badge = "audio"
	Easy        Badge = "easy"
	Medium      Badge = "medium"
	Hard        Badge = "hard"
)

type Badges []Badge

func (b Badges) Has(badges ...Badge) bool {
	for _, badge := range badges {
		for _, item := range b {
			if item == badge {
				return true
			}
		}
	}

	return false
}

func (b Badges) String() string {
	items := make([]string, len(b))

	for i, item := range b {
		items[i] = string(item)
	}

	return strings.Join(items, ", ")
}

type Audience string

const (
	All               Audience = "all"
	AllProfessionals  Audience = "all professionals"
	LinuxUsers        Audience = "Linux users"
	WindowsUsers      Audience = "Windows users"
	MacUsers          Audience = "Mac users"
	AllDevelopers     Audience = "all developers"
	WebDevelopers     Audience = "web developers"
	MobileDevelopers  Audience = "mobile developers"
	DesktopDevelopers Audience = "desktop developers"
	GameDevelopers    Audience = "game developers"
	SysAdmins         Audience = "sysadmins"
)

var validAudiences = map[Audience]struct{}{
	All:               {},
	AllProfessionals:  {},
	LinuxUsers:        {},
	WindowsUsers:      {},
	MacUsers:          {},
	AllDevelopers:     {},
	WebDevelopers:     {},
	MobileDevelopers:  {},
	DesktopDevelopers: {},
	GameDevelopers:    {},
	SysAdmins:         {},
}

type Importance string

const (
	Critical  Importance = "critical"
	Essential Importance = "essential"
	Important Importance = "important"
	Relevant  Importance = "relevant"
	Optional  Importance = "optional"
)

func (i Importance) Level() int {
	switch i {
	case Critical:
		return 5
	case Essential:
		return 4
	case Important:
		return 3
	case Relevant:
		return 2
	case Optional:
		return 1
	}

	return -1
}

type MainStatus string

const (
	VideoMissing       MainStatus = "missing"
	VideoReallyMissing MainStatus = "really missing"
	VideoPresent       MainStatus = "present"
	VideoProblem       MainStatus = "problem"
)

type Main struct {
	Status MainStatus
	Videos Videos
}

func (m Main) Has(badges ...Badge) bool {
	return m.Videos.Has(badges...)
}

func (m Main) GetIssues() []string {
	return m.Videos.GetIssues()
}

type Video struct {
	Badges  Badges
	Issues  []string
	Minutes int
	Valid   bool
}

type Videos []Video

func (v Videos) GetIssues() []string {
	var issues []string

	for _, item := range v {
		issues = append(issues, item.Issues...)
	}

	return issues
}

func (v Videos) Has(badges ...Badge) bool {
	for _, item := range v {
		for _, badge := range badges {
			for _, itemBadge := range item.Badges {
				if itemBadge == badge {
					return true
				}
			}
		}
	}

	return false
}

func isOrderedCorrectly(goldenMap map[string]int, givenSlice []string) (string, bool) {
	found := make(map[string]struct{}, len(givenSlice))

	lastIndex := -1
	for _, item := range givenSlice {
		// duplicate section
		if _, foundAlready := found[item]; foundAlready {
			return item, false
		}

		found[item] = struct{}{}

		// right order
		if index, exists := goldenMap[item]; exists {
			if index < lastIndex {
				return item, false
			}

			lastIndex = index

			continue
		}

		// unexpected section
		return item, false
	}

	return "", true
}

type Body interface {
	GetIssues(state State) []string
	CalculateState() (State, error)
	IsSlugForced() bool
}

type Content struct {
	Title             string
	State             State
	Body              Body
	Slug              string
	Weight            string
	Audience          Audience
	Importance        Importance
	OutsideImportance Importance
	Tags              []string
	EmptySections     []string
	Links             map[string]string
}

var regexDashes = regexp.MustCompile(`-+-`)
var regexSlugReduce = regexp.MustCompile(`[:,/?! ]`)
var regexSlugRemove = regexp.MustCompile(`[.'"\\]`)
var regexAbbreviation = regexp.MustCompile(` ([A-Z])\. `)

func slugify(title string) string {
	for _, match := range regexAbbreviation.FindAllStringSubmatch(title, -1) {
		title = strings.Replace(title, match[1]+".", match[1], -1)
	}

	title = strings.ToLower(title)
	title = strings.Replace(title, "c++", "cpp", -1)
	title = strings.Replace(title, "(", "", -1)
	title = strings.Replace(title, ")", "", -1)
	title = strings.Replace(title, "i/o", "io", -1)
	title = strings.Replace(title, "#", "-sharp-", -1)
	title = strings.Replace(title, "&", "-and-", -1)
	title = strings.TrimSuffix(title, ".")
	title = strings.Replace(title, ".", "-dot-", -1)

	// title = strings.Replace(title, ".", "", -1)
	title = regexSlugRemove.ReplaceAllString(title, "")
	title = regexSlugReduce.ReplaceAllString(title, "-")
	title = regexDashes.ReplaceAllString(title, "-")

	title = slug.Make(title)

	return strings.Trim(title, "-")
}

func (c Content) GetIssues(filePath, course, chapter, page string) []string {
	issues := c.Body.GetIssues(c.State)

	slug := slugify(c.Title)

	_, isIndex := c.Body.(*IndexBody)
	if !isIndex {
		if !strings.HasPrefix(page, c.Weight) {
			issues = append(issues, fmt.Sprintf("file name is not prefixed with the weight of the page, file name: %s, weight: %s", page, c.Weight))
		}

		if fmt.Sprintf("%s-%s.md", c.Weight, c.Slug) != page {
			issues = append(issues, fmt.Sprintf("file name does not match the dash joined weight and slug, file name: %s, weight: %s", page, c.Weight))
		}

		if !c.Body.IsSlugForced() && c.Slug != slug {
			issues = append(issues, fmt.Sprintf("slug does not match the lowercase title with dashes (`%s`, `%s`)", c.Slug, slug))
		}
	} else {
		if chapter != slug {
			issues = append(issues, fmt.Sprintf("chapter does not match the slug, file name: %s, chapter: %s, slug: %s", page, chapter, slug))
		}
	}

	if c.State == Complete && len(c.EmptySections) > 0 {
		issues = append(issues, fmt.Sprintf("empty sections: %s", strings.Join(c.EmptySections, ", ")))
	}

	if _, exists := validAudiences[c.Audience]; !exists {
		issues = append(issues, "invalid audience: "+string(c.Audience))
	}

	if c.Importance.Level() < c.OutsideImportance.Level() {
		issues = append(issues, "importance is lower than outside importance")
	}

	if c.OutsideImportance == "" && c.Audience != All {
		issues = append(issues, "outside importance is invalid")
	}

	if c.Audience == All && c.OutsideImportance != "" {
		issues = append(issues, "audience is 'all', outside importance must be empty")
	}

	for _, tag := range c.Tags {
		if tag == "unsorted" {
			issues = append(issues, "tag is 'unsorted'")
		}
		if strings.ToLower(tag) != tag {
			issues = append(issues, "tag is not lowercase: "+tag)
		}
		if strings.Replace(tag, " ", "", 1) != tag {
			issues = append(issues, "tag contains spaces: "+tag)
		}
	}

	return issues
}

type Page struct {
	Title    string
	Content  Content
	Course   string
	Chapter  string
	FileName string
}

func (p Page) GetIssues() []string {
	issues := p.Content.GetIssues(p.FileName, p.Course, p.Chapter, p.Title)

	return issues
}

func (p Page) GetErrors() []string {
	var errors []string

	for _, issue := range p.GetIssues() {
		errors = append(errors, fmt.Sprintf("%s - %s", p.FileName, issue))
	}

	return errors
}

func (p Page) GetWeight() int {
	rawWeight := p.Content.Weight

	val, err := strconv.Atoi(rawWeight)
	if err != nil {
		return 0
	}

	return val
}

func (p Page) GetState() State {
	return p.Content.State
}

func (p Page) String() string {
	color := cliRed

	switch p.GetState() {
	case Complete:
		color = cliGreen
	case Incomplete:
		color = cliYellow
	}

	issues := p.GetIssues()
	if len(issues) > 0 {
		color = cliRed
	}

	result := fmt.Sprintln("    ", color, p.FileName, cliReset, "-", p.Content.State)

	for _, issue := range issues {
		result += fmt.Sprintln("        - ", issue)
	}

	return result
}

func (p Page) GetInternalLink() string {
	filePath := p.FileName

	if strings.HasPrefix(filePath, "content/") {
		filePath = strings.TrimPrefix(filePath, "content")
	}

	if strings.HasSuffix(filePath, "_index.md") {
		return strings.TrimSuffix(filePath, "_index.md")
	}

	filename := filepath.Base(filePath)

	return strings.Replace(filePath, filename, p.Content.Slug, 1) + "/"
}

type Pages []Page

func (p Pages) Add(filePath, courseFN, chapterFN, pageFN string, content Content) Pages {
	return append(p, Page{FileName: filePath, Course: courseFN, Chapter: chapterFN, Title: pageFN, Content: content})
}

type Chapter struct {
	Course  string
	Chapter string
	Pages   Pages
}

func (c *Chapter) String(statesAllowed map[State]struct{}, printIndex, printNonIndex bool) string {
	result := fmt.Sprintln("  ", c.Chapter)

	for _, page := range c.Pages {
		if !printNonIndex && !strings.HasSuffix(page.FileName, "_index.md") {
			continue
		}

		if !printIndex && strings.HasSuffix(page.FileName, "_index.md") {
			continue
		}

		if statesAllowed != nil {
			if _, ok := statesAllowed[page.GetState()]; !ok {
				continue
			}
		}

		result += page.String()
	}

	return result
}

func (c *Chapter) GetWeight() int {
	for _, page := range c.Pages {
		if page.Title != "_index.md" {
			continue
		}

		rawWeight := page.Content.Weight

		val, err := strconv.Atoi(rawWeight)
		if err != nil {
			return 0
		}

		return val
	}

	return 0
}

func (c *Chapter) GetErrors() []string {
	var errors []string

	for _, page := range c.Pages {
		errors = append(errors, page.GetErrors()...)
	}

	return errors
}

func (c *Chapter) GetPageOrderIssues() []string {
	seen := make(map[int][]string, len(c.Pages))
	largestWeight := 0
	var issues []string

	for _, page := range c.Pages {
		if page.Title == "_index.md" {
			continue
		}

		weight := page.GetWeight()
		seen[weight] = append(seen[weight], page.FileName)

		if weight > largestWeight {
			largestWeight = weight
		}

		if weight%10 != 0 {
			issues = append(issues, fmt.Sprintf("weird weight: %d (%s)", weight, c.Chapter))
		}
	}

	if largestWeight < 1 {
		issues = append(issues, fmt.Sprintf("no pages found in chapter (%s)", c.Chapter))
	}

	var missing []int
	for i := 1; i <= largestWeight; i++ {
		if _, ok := seen[i]; !ok {
			if i%10 == 0 {
				missing = append(missing, i)
				continue
			}
		}

		if len(seen[i]) > 1 {
			issues = append(issues, fmt.Sprintf("duplicate pages with weight %d: %s (%s)", i, strings.Join(seen[i], ", "), c.Chapter))
		}
	}

	if len(missing) > 0 {
		issues = append(issues, fmt.Sprintf("missing pages with weight %v (%s)", missing, c.Chapter))
	}

	return issues
}

func (c *Chapter) GetLinks() map[string]string {
	links := make(map[string]string)

	for _, page := range c.Pages {
		for index, link := range page.Content.Links {
			links[page.FileName+":"+index] = link
		}
	}

	return links
}

type Chapters []*Chapter

func (c Chapters) Add(filePath, courseFN, chapterFN, pageFN string, content Content) Chapters {
	for i, chapter := range c {
		if chapter.Chapter == chapterFN {
			c[i].Pages = c[i].Pages.Add(filePath, courseFN, chapterFN, pageFN, content)
			return c
		}
	}

	return append(
		c,
		&Chapter{
			Course:  courseFN,
			Chapter: chapterFN,
			Pages: Pages{
				{
					FileName: filePath,
					Course:   courseFN,
					Chapter:  chapterFN,
					Title:    pageFN,
					Content:  content,
				},
			},
		})
}

type Course struct {
	Course   string
	Chapters Chapters
}

func (c Course) String(statesAllowed map[State]struct{}, printIndex, printNonIndex bool) string {
	result := fmt.Sprintln(c.Course)

	for _, chapter := range c.Chapters {
		result += chapter.String(statesAllowed, printIndex, printNonIndex)
	}

	return result
}

func (c Course) GetChapterOrderIssues() []string {
	seen := make(map[int][]string, len(c.Chapters))
	largestWeight := 0

	for _, chapter := range c.Chapters {
		weight := chapter.GetWeight()
		seen[weight] = append(seen[weight], chapter.Chapter)

		if weight > largestWeight {
			largestWeight = weight
		}
	}

	var issues []string

	if largestWeight < 1 {
		issues = append(issues, fmt.Sprintf("no chapters found in course (%s)", c.Course))
	}

	var missing []int
	for i := 1; i <= largestWeight; i++ {
		if _, ok := seen[i]; !ok {
			missing = append(missing, i)

			continue
		}

		if len(seen[i]) > 1 {
			issues = append(issues, fmt.Sprintf("duplicate chapters with weight %d: %s (%s)", i, strings.Join(seen[i], ", "), c.Course))
		}
	}

	if len(missing) > 0 {
		issues = append(issues, fmt.Sprintf("missing chapter with weight %v (%s)", missing, c.Course))
	}

	return issues
}

func (c Course) GetPageOrderIssues() []string {
	var issues []string

	for _, chapter := range c.Chapters {
		issues = append(issues, chapter.GetPageOrderIssues()...)
	}

	return issues
}

func (c Course) GetLinks() map[string]string {
	allLinks := make(map[string]string)

	for _, chapter := range c.Chapters {
		for page, link := range chapter.GetLinks() {
			allLinks[page] = link
		}
	}

	return allLinks
}

func (c Course) GetErrors() []string {
	var issues []string

	for _, chapter := range c.Chapters {
		issues = append(issues, chapter.GetErrors()...)
	}

	return issues
}

func (c Course) Stats() (int, int, int, int, int) {
	var (
		total, stub, incomplete, complete, errors int
	)

	for _, chapter := range c.Chapters {
		for _, page := range chapter.Pages {
			switch page.GetState() {
			case Stub:
				stub++
			case Incomplete:
				incomplete++
			case Complete:
				complete++
			}

			if len(page.GetIssues()) > 0 {
				errors++
			}
		}

		total += len(chapter.Pages)
	}

	return total, stub, incomplete, complete, errors
}

type Courses []Course

func (c Courses) Add(filePath, courseFN, chapterFN, pageFN string, content Content) Courses {
	for i, course := range c {
		if course.Course == courseFN {
			c[i].Chapters = c[i].Chapters.Add(filePath, courseFN, chapterFN, pageFN, content)
			return c
		}
	}

	return append(
		c,
		Course{
			Course: courseFN,
			Chapters: Chapters{
				{
					Chapter: chapterFN,
					Course:  courseFN,
					Pages: Pages{
						{
							Title:    pageFN,
							Content:  content,
							Course:   courseFN,
							Chapter:  chapterFN,
							FileName: filePath,
						},
					},
				},
			},
		})
}

func (c Courses) GetValidInternalLinks() map[string]struct{} {
	pages := make(map[string]struct{})

	for _, course := range c {
		for _, chapter := range course.Chapters {
			for _, page := range chapter.Pages {
				pages[page.GetInternalLink()] = struct{}{}
			}
		}
	}

	return pages
}

func column(raw interface{}, width int, color Color) string {
	content := fmt.Sprint(raw)

	if len(content) > width {
		return content[:width]
	}

	if content == "0" {
		color = cliReset
	}

	return fmt.Sprint(color, content, cliReset, strings.Repeat(" ", width-len(content)))
}
