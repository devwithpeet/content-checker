package pkg

import "errors"

type IndexBody struct {
	HasEpisodes bool
	State       State
}

func (ib *IndexBody) GetIssues(_ State) []string {
	return nil
}

func (ib *IndexBody) CalculateState() (State, error) {
	if ib.HasEpisodes {
		return ib.State, nil
	}

	return Stub, errors.New("no episodes")
}

func (ib *IndexBody) SetState(state State) {
	ib.State = state
}

func (ib *IndexBody) IsSlugForced() bool {
	return false
}
