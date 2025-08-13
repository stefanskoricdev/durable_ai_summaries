package shared

import (
	"github.com/nlpodyssey/openai-agents-go/agents"
	"github.com/openai/openai-go"
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

func NewTriageAgent() *agents.Agent {
	refine_agent := NewRefineAgent()
	search_agent := NewSearchAgent()
	return agents.New("Triage agent").
		WithInstructions(TRIAGE_AGENT_INSTRUCTIONS).
		WithAgentHandoffs(refine_agent, search_agent).
		WithModel(openai.ChatModelGPT4oMini)
}
