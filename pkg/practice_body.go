package pkg

import "errors"

type PracticeBody struct {
	HasDescription           bool
	HasRecommendedChallenges bool
	HasAdditionalChallenges  bool
}

func (pb PracticeBody) GetIssues(_ State) []string {
	return nil
}

func (pb PracticeBody) CalculateState() (State, error) {
	if !pb.HasDescription {
		return Stub, errors.New("no description")
	}

	if pb.HasRecommendedChallenges && pb.HasAdditionalChallenges {
		return Complete, nil
	}

	return Incomplete, errors.New("missing recommended or additional challenges")
}

func (pb PracticeBody) IsSlugForced() bool {
	return false
}
