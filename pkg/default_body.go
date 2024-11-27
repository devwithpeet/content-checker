package pkg

import (
	"errors"
	"fmt"
)

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

	calculatedStates, err := db.CalculateState()
	if state != calculatedStates {
		msg := "unknown"
		if err != nil {
			msg = err.Error()
		}

		issues = append(issues, fmt.Sprintf("state mismatch. got: %s, want: %s, reason: %s", state, calculatedStates, msg))
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

func (db DefaultBody) CalculateState() (State, error) {
	err := db.isComplete()
	if err == nil {
		return Complete, nil
	}

	err2 := db.isIncomplete()
	if err2 == nil {
		return Incomplete, err
	}

	return Stub, err2
}

func (db DefaultBody) isComplete() error {
	if db.Main.Status != VideoPresent {
		return errors.New("video not present")
	}

	if !db.HasSummary {
		return errors.New("summary section missing")
	}

	if !db.HasExercises {
		return errors.New("exercises section missing")
	}

	if db.Main.Has(Unchecked) {
		return errors.New("main has unchecked videos")
	}

	if db.RelatedVideos.Has(Unchecked) {
		return errors.New("related videos include unchecked videos")
	}

	return nil
}

func (db DefaultBody) isIncomplete() error {
	if db.Main.Status == VideoPresent || db.UsefulWithoutVideo {
		return nil
	}

	if db.RelatedVideos.Has(Alternative, DeepDive, FullCourse) {
		return nil
	}

	return errors.New("video missing and no alternative, deep-dive or full-course videos")
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
