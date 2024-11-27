package pkg

import "errors"

type IndexBody struct {
	HasEpisodes   bool
	CompleteState State
}

func (ib *IndexBody) GetIssues(_ State) []string {
	return nil
}

func (ib *IndexBody) CalculateState() (State, error) {
	if ib.HasEpisodes {
		return ib.CompleteState, nil
	}

	return Stub, errors.New("no episodes")
}

func (ib *IndexBody) SetCompleteState(state State) {
	ib.CompleteState = state
}

func (ib *IndexBody) IsSlugForced() bool {
	return false
}
