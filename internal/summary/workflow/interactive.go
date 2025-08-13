package workflow

import (
	"api/internal/summary/activity"
	"api/internal/summary/shared"
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
)

type InteractiveWorkflowParams struct {
	InitQuery string
}

type Status string

const (
	StatusPending          Status = "pending"
	StatusAwaitsRefinement Status = "awaits_refinement"
	StatusRefined          Status = "refined"
	StatusAwaitsSelection  Status = "awaits_selection"
	StatusError            Status = "error"
	StatusCompleted        Status = "completed"
)

const (
	QueryCheckState      = "state"
	QueryAnswer          = "answer"
	QuerySearchSelection = "search_selection"
)

type InteractiveWorkflowState struct {
	Status              Status
	InitQuery           string
	RefinementQuestions []string
	RefinementAnswers   []string
	SearchResults       []shared.SearchResult
	SearchSelection     *int64
}

func InteractiveWorkflow(ctx workflow.Context, params InteractiveWorkflowParams) (err error) {
	state := InteractiveWorkflowState{
		Status:              StatusPending,
		InitQuery:           "",
		RefinementQuestions: []string{},
		RefinementAnswers:   []string{},
		SearchResults:       []shared.SearchResult{},
		SearchSelection:     nil,
	}

	state.InitQuery = params.InitQuery

	err = workflow.SetQueryHandler(ctx, QueryCheckState, func() (InteractiveWorkflowState, error) {
		return state, nil
	})

	if err != nil {
		return
	}

	err = workflow.SetQueryHandler(ctx, QueryAnswer, func(answers []string) (bool, error) {
		state.RefinementAnswers = answers
		state.Status = StatusRefined
		return true, nil
	})

	if err != nil {
		return
	}

	err = workflow.SetQueryHandler(ctx, QuerySearchSelection, func(choiceIndex int64) (bool, error) {
		state.SearchSelection = &choiceIndex
		state.Status = StatusCompleted

		return true, nil
	})

	if err != nil {
		return
	}

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 5,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	for state.Status != StatusCompleted {

		if state.Status == StatusError {
			break
		}

		if len(state.RefinementAnswers) >= len(state.RefinementQuestions) && state.Status == StatusAwaitsRefinement {
			state.Status = StatusRefined
			continue
		}

		if state.Status == StatusPending {
			var output shared.WithRefineOutput

			err := workflow.ExecuteActivity(ctx, activity.Refine, state.InitQuery).Get(ctx, &output)

			if err != nil {
				state.Status = "error"
				break
			}

			if len(output.RefineQuestions) > 0 {
				state.Status = StatusAwaitsRefinement
				state.RefinementQuestions = output.RefineQuestions
				continue
			}

			state.Status = StatusAwaitsSelection
			state.SearchResults = output.SearchResults
			continue
		}

		if state.Status == StatusRefined {
			var output shared.WithRefineOutput

			enriched_query := fmt.Sprintf(`Original query: %s \n\n Additional context from clarifications: \n`, state.InitQuery)

			for i, answer := range state.RefinementAnswers {
				enriched_query += fmt.Sprintf("- %s? %s \n", state.RefinementQuestions[i], answer)
			}

			err := workflow.ExecuteActivity(ctx, activity.Search, enriched_query).Get(ctx, &output)

			if err != nil {
				state.Status = StatusError
			}

			state.Status = StatusAwaitsSelection
			state.SearchResults = output.SearchResults
			continue
		}

		if state.Status == StatusCompleted {
			workflow.Sleep(ctx, time.Second*1)
			continue
		}

		workflow.Sleep(ctx, time.Second*3)
	}

	return
}
