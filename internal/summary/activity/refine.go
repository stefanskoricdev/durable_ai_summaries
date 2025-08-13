package activity

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

const TRIAGE_AGENT_INSTRUCTIONS = `
	You are a triage agent that determines if a query needs clarifying questions to provide better results.

	Analyze the user's query and decide: 
	**Route to REFINE AGENT if the query:**
	- Lacks specific details about topic
	- Lacks details about search filters like duration in minutes, sort by (relevance or upload date or view count or 	rating)	

	**Route to SEARCH_AGENT if the query:**
	- Is already very specific with clear parameters
	- Contains details like title and search filters
	- Has sufficient context to conduct searching

	• If clarifications needed → call transfer_to_refine_agent
	• If specific enough → call transfer_to_search_agent

	Return exactly ONE function-call.
	`

func newTriageAgent() *agents.Agent {
	refine_agent := newRefineAgent()
	search_agent := newSearchAgent()
	return agents.New("Triage agent").
		WithInstructions(TRIAGE_AGENT_INSTRUCTIONS).
		WithAgentHandoffs(refine_agent, search_agent).
		WithModel(openai.ChatModelGPT4oMini)
}

type SearchResult struct {
	Title string
	URL   string
}

type WithRefineOutput struct {
	RefineQuestions []string       `json:"refineQuestions" jsonschema_description:"A list of refining questions"`
	SearchResults   []SearchResult `json:"searchResults" jsonschema_description:"A list of search results that match query params"`
}

const REFINE_AGENT_INSTRUCTIONS = `
	You are a refining agent that should ask clarification questions to gather more informations to perform search.
	
	Things you should ask if it lacks from a prompt are:
	1. What is a topic we should search for no need to go into details about aspects of it (eg. if user provides "Javascript" only then no need to ask more about this we know what is a topic user wants to search)
	2. what duration of video should we search for - Possible values:
		1. "videoDurationUnspecified",
		2. "any" - Do not filter video search results based on their duration. This
		is the default value.
		3. "short" - Only include videos that are less than four minutes long.
		4. "medium" - Only include videos that are between four and 20 minutes long
		(inclusive).
		5. "long" - Only include videos longer than 20 minutes.
	3. how we should sort them by - Possible values:
		1. "searchSortUnspecified"
		2. "date" - Resources are sorted in reverse chronological order based on the
		date they were created.
		3. "rating" - Resources are sorted from highest to lowest rating.
		4. "viewCount" - Resources are sorted from highest to lowest number of views.
		5. "relevance" (default) - Resources are sorted based on their relevance to
		the search query. This is the default value for this parameter.
		6. "title" - Resources are sorted alphabetically by title.
		7. "videoCount" - Channels are sorted in descending order of their number of
		uploaded videos.

	GUIDELINES:
	1. **Be concise while gathering all necessary information** Ask 2–3 clarifying questions to gather more details for searching.
	- Make sure to gather all the information needed to carry out the search task in a concise, well-structured manner. Use bullet points or numbered lists if appropriate for clarity. Don't ask for unnecessary information, or information that the user has already provided.
`

func newRefineAgent() *agents.Agent {
	return agents.New("Refine agent").
		WithInstructions(REFINE_AGENT_INSTRUCTIONS).
		WithModel(openai.ChatModelGPT4oMini).
		WithOutputType(agents.OutputType[WithRefineOutput]())
}

type SearchParams struct {
	Topic    string `json:"topic"`
	Duration string `json:"duration"`
	Sort_BY  string `json:"sort_by"`
}

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

func newSearchAgent() *agents.Agent {
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

func Refine(ctx context.Context, query string) (*WithRefineOutput, error) {
	triageAgent := newTriageAgent()
	result, err := agents.Run(ctx, triageAgent, query)

	if err != nil {
		fmt.Printf("Issue running triage agent: %s\n", err.Error())
		return nil, err
	}

	output := result.FinalOutput.(WithRefineOutput)

	fmt.Printf("Triage agent output: %v", output)

	return &WithRefineOutput{
		RefineQuestions: output.RefineQuestions,
		SearchResults:   output.SearchResults,
	}, nil
}

func Search(ctx context.Context, query string) (*WithRefineOutput, error) {
	searchAgent := newSearchAgent()
	result, err := agents.Run(ctx, searchAgent, query)

	if err != nil {
		fmt.Printf("Issue while running search agent: %v", err.Error())
		return nil, err
	}

	output := result.FinalOutput.(WithRefineOutput)

	return &WithRefineOutput{
		RefineQuestions: output.RefineQuestions,
		SearchResults:   output.SearchResults,
	}, nil
}
