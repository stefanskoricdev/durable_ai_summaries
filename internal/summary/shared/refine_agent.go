package shared

import (
	"github.com/nlpodyssey/openai-agents-go/agents"
	"github.com/openai/openai-go"
)

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

func NewRefineAgent() *agents.Agent {
	return agents.New("Refine agent").
		WithInstructions(REFINE_AGENT_INSTRUCTIONS).
		WithModel(openai.ChatModelGPT4oMini).
		WithOutputType(agents.OutputType[WithRefineOutput]())
}
