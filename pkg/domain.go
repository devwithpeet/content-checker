package pkg

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

type Color string

const (
	cliRed    Color = "\x1b[31m"
	cliYellow Color = "\x1b[33m"
	cliGreen  Color = "\x1b[32m"
	cliBlue   Color = "\x1b[34m"
	cliReset  Color = "\x1b[0m"
	cliBold   Color = "\x1b[1m"
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

type DefaultBody struct {
	Main               Main
	HasSummary         bool
	HasTopics          bool
	HasExercises       bool
	RelatedVideos      Videos
	HasRelatedLinks    bool
	UsefulWithoutVideo bool
	SlugForced         bool
	Project            bool
	SectionTitles      []string
}

var defaultBodySectionMap = map[string]int{
	sectionRoot:            0,
	sectionMainVideo:       1,
	sectionSummary:         2,
	sectionTopics:          3,
	sectionCode:            4,
	sectionRelatedLessons:  5,
	sectionRelatedVideos:   6,
	sectionRelatedArticles: 7,
	sectionRelatedLinks:    8,
	sectionExercises:       9,
	sectionNotes:           10,
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

func (db DefaultBody) GetIssues(state State) []string {
	issues := db.Main.GetIssues()
	issues = append(issues, db.RelatedVideos.GetIssues()...)

	switch db.Main.Status {
	case VideoReallyMissing:
		if db.UsefulWithoutVideo {
			issues = append(issues, "main video is NOT REALLY missing (Remove the useful-without-video tag?")
		}
	case VideoMissing:
		if !db.RelatedVideos.Has(Alternative, DeepDive, FullCourse) && !db.UsefulWithoutVideo {
			issues = append(issues, "main video is REALLY missing (Add a useful-without-video tag?")
		}
	}

	if state != db.CalculateState() {
		issues = append(issues, fmt.Sprintf("state mismatch. got: %s, want: %s", state, db.CalculateState()))
	}

	if item, ok := isOrderedCorrectly(defaultBodySectionMap, db.SectionTitles); !ok {
		issues = append(issues, "sections are not in the correct order, first out of order: "+item)
	}

	if !db.Project {
		if !db.HasSummary {
			issues = append(issues, "summary section is missing")
		}

		if !db.HasTopics {
			issues = append(issues, "topics section is missing")
		}
	}

	return issues
}

func (db DefaultBody) CalculateState() State {
	if db.isComplete() {
		return Complete
	}

	if db.isIncomplete() {
		return Incomplete
	}

	return Stub
}

func (db DefaultBody) isComplete() bool {
	if db.Main.Status != VideoPresent {
		return false
	}

	if !db.HasSummary {
		return false
	}

	if !db.HasExercises {
		return false
	}

	if db.Main.Has(Unchecked) {
		return false
	}

	if db.RelatedVideos.Has(Unchecked) {
		return false
	}

	return true
}

func (db DefaultBody) isIncomplete() bool {
	if db.Main.Status == VideoPresent || db.UsefulWithoutVideo {
		return true
	}

	if db.RelatedVideos.Has(Alternative, DeepDive, FullCourse) {
		return true
	}

	return false
}

func (db DefaultBody) GetStatus() MainStatus {
	return db.Main.Status
}

func (db DefaultBody) IsIndex() bool {
	return false
}

func (db DefaultBody) IsSlugForced() bool {
	return db.SlugForced
}

type IndexBody struct {
	HasEpisodes   bool
	CompleteState State
}

func (ib *IndexBody) GetIssues(_ State) []string {
	return nil
}

func (ib *IndexBody) CalculateState() State {
	if ib.HasEpisodes {
		return ib.CompleteState
	}

	return Stub
}

func (ib *IndexBody) SetCompleteState(state State) {
	ib.CompleteState = state
}

func (ib *IndexBody) IsSlugForced() bool {
	return false
}

type PracticeBody struct {
	HasDescription           bool
	HasRecommendedChallenges bool
	HasAdditionalChallenges  bool
}

func (pb PracticeBody) GetIssues(_ State) []string {
	return nil
}

func (pb PracticeBody) CalculateState() State {
	if !pb.HasDescription {
		return Stub
	}

	if pb.HasRecommendedChallenges && pb.HasAdditionalChallenges {
		return Complete
	}

	return Incomplete
}

func (pb PracticeBody) IsSlugForced() bool {
	return false
}

type Body interface {
	GetIssues(state State) []string
	CalculateState() State
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
}

var regexDashes = regexp.MustCompile(`-+-`)
var regexSlugReduce = regexp.MustCompile(`[:,/?! ]`)
var regexSlugRemove = regexp.MustCompile(`[.'"\\]`)

func slugify(title string) string {
	title = strings.ToLower(title)
	title = strings.Replace(title, "i/o", "io", -1)
	title = strings.Replace(title, "#", "-sharp-", -1)
	title = strings.Replace(title, "&", "-and-", -1)
	title = strings.TrimSuffix(title, ".")
	title = strings.Replace(title, ".", "-dot-", -1)

	// title = strings.Replace(title, ".", "", -1)
	title = regexSlugRemove.ReplaceAllString(title, "")
	title = regexSlugReduce.ReplaceAllString(title, "-")
	title = regexDashes.ReplaceAllString(title, "-")

	return strings.Trim(title, "-")
}

func (c Content) GetIssues(filePath string) []string {
	issues := c.Body.GetIssues(c.State)

	_, isIndex := c.Body.(*IndexBody)
	if !isIndex {
		filename := filepath.Base(filePath)
		if !strings.HasPrefix(filename, c.Weight) {
			issues = append(issues, "file name is not prefixed with the weight of the page")
		}
		if fmt.Sprintf("%s-%s.md", c.Weight, c.Slug) != filename {
			issues = append(issues, "file name does not match the dash joined weight and slug")
		}
		if !c.Body.IsSlugForced() && c.Slug != slugify(c.Title) {
			issues = append(issues, fmt.Sprintf("slug does not match the lowercase title with dashes (`%s`, `%s`)", c.Slug, slugify(c.Title)))
		}
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
	FilePath string
	Title    string
	Content  Content
}

func (p Page) GetIssues() []string {
	issues := p.Content.GetIssues(p.FilePath)

	return issues
}

func (p Page) GetErrors() []string {
	var errors []string

	for _, issue := range p.GetIssues() {
		errors = append(errors, fmt.Sprintf("%s - %s", p.FilePath, issue))
	}

	return errors
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

	result := fmt.Sprintln("    ", color, p.FilePath, "-", p.Content.State, cliReset)

	for _, issue := range issues {
		result += fmt.Sprintln("        - ", issue)
	}

	return result
}

type Pages []Page

func (p Pages) Add(filePath, pageFN string, content Content) Pages {
	return append(p, Page{FilePath: filePath, Title: pageFN, Content: content})
}

type Chapter struct {
	Title    string
	Pages    Pages
	prepared bool
}

func (c *Chapter) Prepare() {
	if c.prepared {
		return
	}

	c.prepared = true

	var (
		indexPage  *IndexBody
		pagesExist = false
		incomplete = false
	)

	for _, page := range c.Pages {
		chapter, ok := page.Content.Body.(*IndexBody)
		if ok {
			indexPage = chapter

			continue
		}

		pagesExist = true
		if page.GetState() != Complete {
			incomplete = true
		}

		if incomplete && indexPage != nil {
			break
		}
	}

	if indexPage == nil || !pagesExist || incomplete {
		return
	}

	indexPage.SetCompleteState(Complete)
}

func (c *Chapter) String(statesAllowed map[State]struct{}, printIndex, printNonIndex bool) string {
	result := fmt.Sprintln("  ", c.Title)

	c.Prepare()

	for _, page := range c.Pages {
		if !printNonIndex && !strings.HasSuffix(page.FilePath, "_index.md") {
			continue
		}

		if !printIndex && strings.HasSuffix(page.FilePath, "_index.md") {
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

func (c *Chapter) GetErrors() []string {
	var errors []string

	for _, page := range c.Pages {
		errors = append(errors, page.GetErrors()...)
	}

	return errors
}

type Chapters []*Chapter

func (c Chapters) Add(filePath, chapterFN, pageFN string, content Content) Chapters {
	for i, chapter := range c {
		if chapter.Title == chapterFN {
			c[i].Pages = c[i].Pages.Add(filePath, pageFN, content)
			return c
		}
	}

	return append(c, &Chapter{Title: chapterFN, Pages: Pages{{FilePath: filePath, Title: pageFN, Content: content}}})
}

type Course struct {
	Title    string
	Chapters Chapters
}

func (c Course) Prepare() {
	for _, chapter := range c.Chapters {
		chapter.Prepare()
	}
}

func (c Course) String(statesAllowed map[State]struct{}, printIndex, printNonIndex bool) string {
	result := fmt.Sprintln(c.Title)

	for _, chapter := range c.Chapters {
		result += chapter.String(statesAllowed, printIndex, printNonIndex)
	}

	return result
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
		if course.Title == courseFN {
			c[i].Chapters = c[i].Chapters.Add(filePath, chapterFN, pageFN, content)
			return c
		}
	}

	return append(c, Course{Title: courseFN, Chapters: Chapters{{Title: chapterFN, Pages: Pages{{FilePath: filePath, Title: pageFN, Content: content}}}}})
}

type CourseStat struct {
	Title      string
	Total      int
	Stub       int
	Incomplete int
	Complete   int
	Errors     int
}

func (cs *CourseStat) Print(columnWidths [7]int, columnColors [7]Color, total int) {
	fmt.Printf(
		"%s | %s | %s | %s | %s | %s | %s\n",
		column(cs.Title, columnWidths[0], columnColors[0]),
		column(cs.Total, columnWidths[1], columnColors[1]),
		column(cs.Stub, columnWidths[2], columnColors[2]),
		column(cs.Incomplete, columnWidths[3], columnColors[3]),
		column(cs.Complete, columnWidths[4], columnColors[4]),
		column(cs.Errors, columnWidths[5], columnColors[5]),
		column(cs.Total*100/total, columnWidths[6], columnColors[6]),
	)
}

func (cs *CourseStat) PrintHead(columnWidths [7]int, columnColors [7]Color) {
	fmt.Printf(
		"%s | %s | %s | %s | %s | %s | %s\n",
		column("Course", columnWidths[0], columnColors[0]),
		column("All", columnWidths[1], columnColors[1]),
		column("Stub", columnWidths[2], columnColors[2]),
		column("Incomplete", columnWidths[3], columnColors[3]),
		column("Complete", columnWidths[4], columnColors[4]),
		column("Errors", columnWidths[5], columnColors[5]),
		column("Percent", columnWidths[6], columnColors[6]),
	)
}

func (cs *CourseStat) Line(columnWidths [7]int) {
	for i, width := range columnWidths {
		if i == 0 {
			fmt.Print(strings.Repeat("-", width+1))

			continue
		}

		fmt.Print("+", strings.Repeat("-", width+2))
	}

	fmt.Println()
}

func (cs *CourseStat) Add(stat CourseStat) {
	cs.Total += stat.Total
	cs.Stub += stat.Stub
	cs.Incomplete += stat.Incomplete
	cs.Complete += stat.Complete
	cs.Errors += stat.Errors
}

func NewCourseStat(title string, total, stub, incomplete, complete, errors int) CourseStat {
	return CourseStat{Title: title, Total: total, Stub: stub, Incomplete: incomplete, Complete: complete, Errors: errors}
}

func (c Courses) Stats() {
	columnWidths := [7]int{15, 5, 4, 10, 8, 6, 7}
	columnColors := [7]Color{cliBold, cliBold, cliBlue, cliYellow, cliGreen, cliRed, cliBold}
	totalStat := NewCourseStat("Total", 0, 0, 0, 0, 0)

	totalStat.PrintHead(columnWidths, columnColors)
	totalStat.Line(columnWidths)

	stats := make([]CourseStat, 0, len(c))
	for _, course := range c {
		courseAll, courseStub, courseIncomplete, courseComplete, courseErrors := course.Stats()

		newStats := NewCourseStat(course.Title, courseAll, courseStub, courseIncomplete, courseComplete, courseErrors)

		stats = append(stats, newStats)

		totalStat.Add(newStats)
	}

	for _, stat := range stats {
		stat.Print(columnWidths, columnColors, totalStat.Total)
	}

	totalStat.Line(columnWidths)
	totalStat.Print(columnWidths, columnColors, totalStat.Total)
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
