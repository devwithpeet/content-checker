package pkg

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	// default
	sectionRoot            = "root"
	sectionMainVideo       = "main video"
	sectionSummary         = "summary"
	sectionTopics          = "topics"
	sectionCode            = "code"
	sectionRelatedLessons  = "related lessons"
	sectionRelatedVideos   = "related videos"
	sectionRelatedArticles = "related articles"
	sectionRelatedLinks    = "related links"
	sectionExercises       = "exercises"
	sectionNotes           = "notes"

	// index
	sectionEpisodes = "episodes"

	// practice
	sectionDescription           = "description"
	sectionRecommendedChallenges = "recommended challenges"
	sectionAdditionalChallenges  = "additional challenges"
)

const (
	maxNonFullCourseLength = 119
	maxExtraLength         = 60
	minDeepDiveLength      = 30
)

const markdownHeaderLength = 3

func ParseMarkdown(rawContent string) (Content, error) {
	if len(rawContent) < markdownHeaderLength*2 {
		return Content{}, errors.New("markdown too short")
	}

	// Convert DOS/Windows line endings (\r\n) into Linux/Unix line endings
	strContent := strings.Replace(rawContent, "\r\n", EOL, -1)

	header, body, err := splitMarkdown(strContent)
	if err != nil {
		return Content{}, fmt.Errorf("markdown header could not be extracted, err: %w", err)
	}

	sections := extractSections(body)
	headers := getHeaderValues(header)
	tags := getTags(headers, nil)

	archetype := getValueWithDefault(headers, "archetype", "")

	var content Content
	if archetype == "chapter" {
		content.Body = sectionsToIndexBody(sections)
	} else if sections.HasNonEmpty(sectionDescription) {
		content.Body = sectionsToPracticeBody(sections)
	} else {
		content.Body = sectionsToDefaultBody(sections, tags)
	}

	content.Slug = getValueWithDefault(headers, "slug", "")
	content.Weight = getValueWithDefault(headers, "weight", "")
	content.Title = getValueWithDefault(headers, "title", "")
	content.State = State(getValueWithDefault(headers, "state", ""))
	content.Audience = Audience(getValueWithDefault(headers, "audience", ""))
	content.Importance = Importance(getValueWithDefault(headers, "audienceImportance", ""))
	content.OutsideImportance = Importance(getValueWithDefault(headers, "outsideImportance", ""))
	content.Tags = tags

	return content, nil
}

func splitMarkdown(in string) (string, string, error) {
	if len(in) < 4 {
		return "", "", errors.New("markdown too short")
	}

	// Handle TOML front matter
	if in[:4] == "+++\n" {
		if idx := strings.Index(in[4:], "\n+++"); idx != -1 {
			return in[4 : idx+4], strings.Trim(in[idx+8:], "\n+"), nil
		}
	}

	return "", "", errors.New("could not split markdown")
}

var regexHeader = regexp.MustCompile(`^(\S+)\s*=\s*(.*)$`)

func getHeaderValues(header string) map[string]string {
	values := make(map[string]string)

	for _, row := range strings.Split(header, "\n") {
		matches := regexHeader.FindStringSubmatch(row)

		if len(matches) != 3 {
			continue
		}

		values[matches[1]] = strings.Trim(matches[2], `'"`)
	}

	return values
}

func getTags(values map[string]string, defaultValue []string) []string {
	tagsRaw, ok := values["tags"]
	if !ok {
		return defaultValue
	}

	var tags []string

	for _, part := range strings.Split(strings.Trim(tagsRaw, `[]`), ",") {
		tags = append(tags, strings.Trim(part, ` '"`))
	}

	return tags
}

func getValueWithDefault(values map[string]string, key, defaultValue string) string {
	if value, ok := values[key]; ok {
		return value
	}

	return defaultValue
}

type Section struct {
	Title   string
	Content string
}

type Sections []Section

func (s Sections) HasNonEmpty(title string) bool {
	for _, section := range s {
		if section.Title == title {
			return len(section.Content) > 0
		}
	}

	return false
}

func (s Sections) Get(title string) string {
	for _, section := range s {
		if section.Title == title {
			return section.Content
		}
	}

	return ""
}

func (s Sections) Titles() []string {
	keys := make([]string, 0, len(s))
	for _, section := range s {
		keys = append(keys, section.Title)
	}

	return keys
}

func extractSections(body string) Sections {
	var sections Sections

	currentSection := "root"
	sectionStart := 0

	rows := strings.Split(body, EOL)
	for i, row := range rows {
		if len(row) >= 3 && row[:3] == "## " {
			content := strings.Trim(strings.Join(rows[sectionStart:i], EOL), " \t\n")
			sections = append(sections, Section{Title: currentSection, Content: content})

			sectionStart = i + 1

			currentSection = strings.ToLower(strings.Trim(row[3:], " \t"))

			continue
		}

		if i > sectionStart && len(row) >= 3 && row[:3] == "---" {
			// if the previous line is empty or non-existent, this is a horizontal rule, not a header
			if i == 0 || len(rows[i-1]) == 0 {
				continue
			}

			content := strings.Trim(strings.Join(rows[sectionStart:i-1], EOL), " \t\n")
			sections = append(sections, Section{Title: currentSection, Content: content})

			sectionStart = i + 1

			currentSection = strings.ToLower(strings.Trim(rows[i-1], " \t"))

			continue
		}
	}

	if currentSection != "root" {
		content := strings.Trim(strings.Join(rows[sectionStart:], EOL), " \t\n")
		sections = append(sections, Section{Title: currentSection, Content: content})
	}

	// Remove the root section if it's empty
	if len(sections) > 0 && sections[0].Content == "" {
		sections = sections[1:]
	}

	return sections
}

var regexMissing = regexp.MustCompile(`{{<\s*main-missing\s*>}}`)
var regexReallyMissing = regexp.MustCompile(`{{<\s*main-really-missing\s*>}}`)
var regexYoutube = regexp.MustCompile(`{{<\s*youtube(-button)?\s+([^>]*)\s*>}}`)

func ExtractMain(content string) Main {
	matchCount := 0
	main := Main{
		Status: VideoProblem,
		Videos: nil,
	}

	matches := regexMissing.FindAllStringSubmatch(content, -1)
	if len(matches) > 0 {
		main.Status = VideoMissing
		matchCount++
	}

	matches = regexReallyMissing.FindAllStringSubmatch(content, -1)
	if len(matches) > 0 {
		main.Status = VideoReallyMissing
		matchCount++
	}

	if matchCount > 1 {
		main.Status = VideoProblem
	}

	matches = regexYoutube.FindAllStringSubmatch(content, -1)
	if len(matches) > 0 {
		main.Videos = ExtractVideos(content, true)

		if matchCount == 0 {
			main.Status = VideoPresent
		}
	}

	return main
}

var regexSubHeader = regexp.MustCompile(`\n####?#? .*\n`)

func ExtractVideos(content string, noBadgeOkay bool) Videos {
	if strings.TrimSpace(content) == "" {
		return nil
	}

	sections := regexSubHeader.Split("\n"+content, -1)

	relatedVideos := make(Videos, 0, len(sections))
	for _, section := range sections {
		if strings.TrimSpace(section) == "" {
			continue
		}

		relatedVideo := extractVideo(section, noBadgeOkay)

		if relatedVideo.Valid {
			relatedVideos = append(relatedVideos, relatedVideo)
		}
	}

	return relatedVideos
}

var regexTime = regexp.MustCompile(`{{<\s*time\s+(\d+)\s*>}}`)

func extractTime(content string) (int, []string) {
	var (
		issues  []string
		minutes int
		err     error
	)

	timeMatches := regexTime.FindAllStringSubmatch(content, -1)
	if len(timeMatches) == 0 {
		issues = append(issues, "missing time shortcode")
	} else {
		minutes, err = strconv.Atoi(timeMatches[0][1])
		if err != nil {
			issues = append(issues, fmt.Sprintf("failed to parse duration: %s", timeMatches[0][1]))
		}
	}
	if len(timeMatches) > 1 {
		issues = append(issues, "multiple time shortcodes found")
	}

	return minutes, issues
}

var regexBadge = regexp.MustCompile(`{{<\s*badge-(\S*)\s*>}}`)

func extractBadges(content string, noBadgeOkay bool) (Badges, bool, []string) {
	var (
		badges = Badges{}
		issues []string
	)

	noEmbed := false
	badgeMatches := regexBadge.FindAllStringSubmatch(content, -1)

	for _, match := range badgeMatches {
		switch badge := Badge(match[1]); badge {
		case Unchecked, Alternative, Extra, Fun, Hint, MustSee, Summary, DeepDive, FullCourse:
			badges = append(badges, badge)
		case NoEmbed:
			noEmbed = true
		case Audio, Easy, Medium, Hard:
			continue
		default:
			issues = append(issues, fmt.Sprintf("Unknown badge: '%s'", badge))
		}
	}

	if len(badges) == 0 {
		if !noBadgeOkay {
			issues = append(issues, "missing badge shortcode")
		}

		return Badges{}, noEmbed, issues
	}

	levelFound := NoEmbed

	for _, badge := range badges {
		if badge == Unchecked || badge == Audio || badge == NoEmbed {
			continue
		}

		if levelFound != NoEmbed {
			issues = append(issues, "unexpected badge shortcode found: "+string(badge))
		}

		levelFound = badge
	}

	return badges, noEmbed, issues
}

func extractYoutube(content string, noEmbed bool) (int, []string) {
	var issues []string

	youtubeMatches := regexYoutube.FindAllStringSubmatch(content, -1)

	switch len(youtubeMatches) {
	case 0:
		if !noEmbed {
			issues = append(issues, "missing youtube shortcode")
		}
	case 1:
		if noEmbed {
			issues = append(issues, "unexpected youtube shortcode together with no-embed badge")
		}
	default:
		issues = append(issues, "multiple youtube shortcodes found")
	}

	return len(youtubeMatches), issues
}

func extractVideo(content string, noBadgeOkay bool) Video {
	var (
		issues  []string
		minutes int
	)

	minutes, timeIssues := extractTime(content)
	issues = append(issues, timeIssues...)

	badges, noEmbed, badgeIssues := extractBadges(content, noBadgeOkay)
	issues = append(issues, badgeIssues...)

	ytCount, ytIssues := extractYoutube(content, noEmbed)
	issues = append(issues, ytIssues...)

	if ytCount == 0 && !noEmbed && len(badges) == 0 && minutes == 0 {
		return Video{}
	}

	if minutes > 0 && len(badges) > 0 && strings.Index(content, "badge") < strings.Index(content, "time") {
		issues = append(issues, "badge should be placed after time")
	}

	if minutes >= maxNonFullCourseLength && !badges.Has(FullCourse, Fun) {
		issues = append(issues, "badges should have full-course, but do not. badges: "+badges.String())
	} else if minutes > maxExtraLength && badges.Has(Extra) {
		issues = append(issues, "badges should have deep-dive, but do not. badges: "+badges.String())
	} else if minutes < minDeepDiveLength && badges.Has(DeepDive) {
		issues = append(issues, "badges should have extra, but do not. badges: "+badges.String())
	}

	return Video{
		Badges:  badges,
		Issues:  issues,
		Minutes: minutes,
		Valid:   true,
	}
}

const (
	tagUsefulWithoutVideo = "useful-without-video"
	tagSlugForced         = "slug-forced"
	tagNoExercise         = "no-exercise"
	tagProjects           = "projects"
)

func sectionsToDefaultBody(sections Sections, tags []string) DefaultBody {
	hasSummary := sections.HasNonEmpty(sectionSummary)
	hasTopics := sections.HasNonEmpty(sectionTopics)
	hasRelatedLinks := sections.HasNonEmpty(sectionRelatedLinks)
	hasExercises := sections.HasNonEmpty(sectionExercises)

	main := ExtractMain(sections.Get(sectionMainVideo))
	relatedVideos := ExtractVideos(sections.Get(sectionRelatedVideos), false)

	if hasExercises && strings.TrimSpace(sections.Get(sectionExercises)) == "" {
		hasExercises = false
	}

	usefulWithoutVideo := false
	isSlugForced := false
	isProject := false
	for _, tag := range tags {
		if tag == tagUsefulWithoutVideo {
			usefulWithoutVideo = true
		}
		if tag == tagNoExercise {
			hasExercises = true
		}
		if tag == tagSlugForced {
			isSlugForced = true
		}
		if tag == tagProjects {
			isProject = true
		}
	}

	return DefaultBody{
		Main:               main,
		HasSummary:         hasSummary,
		HasTopics:          hasTopics,
		RelatedVideos:      relatedVideos,
		HasRelatedLinks:    hasRelatedLinks,
		HasExercises:       hasExercises,
		UsefulWithoutVideo: usefulWithoutVideo,
		SlugForced:         isSlugForced,
		Project:            isProject,
		SectionTitles:      sections.Titles(),
	}
}

func sectionsToIndexBody(sections Sections) *IndexBody {
	return &IndexBody{
		HasEpisodes: sections.HasNonEmpty(sectionEpisodes),
		State:       Incomplete,
	}
}

func sectionsToPracticeBody(sections Sections) *PracticeBody {
	return &PracticeBody{
		HasDescription:           sections.HasNonEmpty(sectionDescription),
		HasRecommendedChallenges: sections.HasNonEmpty(sectionRecommendedChallenges),
		HasAdditionalChallenges:  sections.HasNonEmpty(sectionAdditionalChallenges),
	}
}
