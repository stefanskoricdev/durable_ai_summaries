package activity

import (
	"api/internal/summary/shared"
	"context"
	"fmt"

	"github.com/nlpodyssey/openai-agents-go/agents"
)

func Refine(ctx context.Context, query string) (*shared.WithRefineOutput, error) {
	triageAgent := shared.NewTriageAgent()
	result, err := agents.Run(ctx, triageAgent, query)

	if err != nil {
		fmt.Printf("Issue running triage agent: %s\n", err.Error())
		return nil, err
	}

	output := result.FinalOutput.(shared.WithRefineOutput)

	fmt.Printf("Triage agent output: %v", output)

	return &shared.WithRefineOutput{
		RefineQuestions: output.RefineQuestions,
		SearchResults:   output.SearchResults,
	}, nil
}

func Search(ctx context.Context, query string) (*shared.WithRefineOutput, error) {
	searchAgent := shared.NewSearchAgent()
	result, err := agents.Run(ctx, searchAgent, query)

	if err != nil {
		fmt.Printf("Issue while running search agent: %v", err.Error())
		return nil, err
	}

	output := result.FinalOutput.(shared.WithRefineOutput)

	return &shared.WithRefineOutput{
		RefineQuestions: output.RefineQuestions,
		SearchResults:   output.SearchResults,
	}, nil
}
