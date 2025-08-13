package shared

import (
	"context"
	"fmt"
	"os"

	"github.com/nlpodyssey/openai-agents-go/agents"
	"github.com/nlpodyssey/openai-agents-go/modelsettings"
	"github.com/openai/openai-go"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func search(ctx context.Context, params SearchParams) (WithRefineOutput, error) {
	youtubeService, err := youtube.NewService(ctx, option.WithAPIKey(os.Getenv("YOUTUBE_API_KEY")))

	if err != nil {
		panic("Failed to create YoutubeService")
	}

	res, err := youtubeService.Search.
		List([]string{"id", "snippet"}).
		Q(params.Topic).
		Type("video").
		VideoDuration(params.Duration).
		Order(params.Sort_BY).MaxResults(5).
		Do()

	if err != nil || res.HTTPStatusCode < 200 || res.HTTPStatusCode > 299 {
		fmt.Println(err.Error())
		panic("Failed to list data")
	}

	results := make([]SearchResult, 0)

	for _, item := range res.Items {
		var r SearchResult
		r.Title = item.Snippet.Title
		r.URL = fmt.Sprintf("https://www.youtube.com/watch?v=%s", item.Id.VideoId)

		results = append(results, r)
	}

	return WithRefineOutput{
		RefineQuestions: []string{},
		SearchResults:   results,
	}, nil
}

var searchTool = agents.NewFunctionTool(
	"Search", "Search for youtube videos by given topic and filter options like duration, and sort_by",
	search,
)

const SEARCH_AGENT_INSTRUCTIONS = `You are search assistent.`

func NewSearchAgent() *agents.Agent {
	return agents.New("Search agents").
		WithInstructions(SEARCH_AGENT_INSTRUCTIONS).
		WithTools(searchTool).
		WithToolUseBehavior(agents.RunLLMAgain()).
		WithModelSettings(modelsettings.ModelSettings{
			ToolChoice: modelsettings.ToolChoiceRequired,
		}).
		WithModel(openai.ChatModelGPT4oMini).
		WithOutputType(agents.OutputType[WithRefineOutput]())
}
