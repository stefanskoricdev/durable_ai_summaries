package main

import (
	"api/internal/summary/workflow"
	"api/internal/util"
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	markdown "github.com/Klaus-Tockloth/go-term-markdown"
	"github.com/fatih/color"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

var questionStr = color.New(color.FgGreen, color.Bold).Add(color.Underline).SprintfFunc()
var dangerStr = color.New(color.FgRed, color.Bold).SprintfFunc()

func main() {
	ctx := context.Background()
	topic := util.StringPrompt(questionStr("✨ Ready to discover something new? \n📖 Please tell me what topic you'd like me to search and summarize? "))

	temporalClient, err := client.Dial(client.Options{
		HostPort:  client.DefaultHostPort,
		Namespace: "summarize",
		Logger:    util.CustomLogger{},
	})

	if err != nil {
		dangerStr("Unable to create Temporal Client: %v\n", err)
	}

	defer temporalClient.Close()

	var workflowId string

	workflowStatus := enums.WORKFLOW_EXECUTION_STATUS_UNSPECIFIED

	if workflowId == "" {
		workflowId = fmt.Sprintf("interaction-workflow-%v", time.Now())
	}

	resp, err := temporalClient.DescribeWorkflowExecution(ctx, workflowId, "")

	if err == nil {
		workflowStatus = resp.WorkflowExecutionInfo.Status
	}

	var iwf client.WorkflowRun

	if workflowStatus != enums.WORKFLOW_EXECUTION_STATUS_RUNNING {
		iwf, err = temporalClient.ExecuteWorkflow(
			ctx,
			client.StartWorkflowOptions{
				ID:        workflowId,
				TaskQueue: "summarize",
			},
			workflow.InteractiveWorkflow,
			struct{ InitQuery string }{
				InitQuery: topic,
			},
		)

		if err != nil {
			fmt.Printf("Failed to create new workflow: %v", err.Error())
			panic("Failed to create new workflow")
		}
	}

	util.LogInfo(
		"Just a sec! ⏳ I'm quickly assessing your topic to tailor the best follow-up questions for you. 🚀",
	)

	selectedUrl := ""

	for iwf != nil {
		stateResult, err := temporalClient.QueryWorkflow(
			ctx,
			iwf.GetID(),
			iwf.GetRunID(),
			workflow.QueryCheckState,
		)

		var workflowState workflow.InteractiveWorkflowState

		if err != nil {
			panic("Could not retrieve workflow state")
		}

		if workflowState.Status == workflow.StatusPending {
			time.Sleep(time.Second * 2)
			continue
		}

		if workflowState.Status == workflow.StatusCompleted {
			break
		}

		stateResult.Get(&workflowState)

		if workflowState.Status == workflow.StatusAwaitsRefinement {
			answers := make([]string, 0)

			for i, q := range workflowState.RefinementQuestions {
				a := util.StringPrompt(questionStr("%s", q))
				if i == len(workflowState.RefinementQuestions)-1 {
					fmt.Println("")
				}
				answers = append(answers, a)
			}

			_, err := temporalClient.QueryWorkflow(ctx, iwf.GetID(), iwf.GetRunID(), workflow.QueryAnswer, answers)

			if err != nil {
				panic("Failed to make workflow query with answers")
			}

			util.LogInfo(
				"The hunt is on! 🕵️‍♀️ I'm busy finding exciting video results tailored just for you. Get ready for some insights! ✨",
			)

			continue
		}

		if workflowState.Status == workflow.StatusAwaitsSelection && len(workflowState.SearchResults) > 0 {
			question := questionStr(
				"Results are in! 🎬 Which video would you like me to summarize? Select one from the list below: \n",
			)

			for i, v := range workflowState.SearchResults {
				question += questionStr("[%d]: %s \n", i+1, v.Title)
			}

			answer := util.StringPrompt(question)

			selected, err := strconv.ParseInt(answer, 10, 64)

			if err != nil {
				fmt.Printf("Invalid input, please enter a number: %v\n", err)
				continue
			}

			_, err = temporalClient.QueryWorkflow(
				ctx,
				iwf.GetID(),
				iwf.GetRunID(),
				workflow.QuerySearchSelection,
				selected,
			)

			if err != nil {
				panic("Failed to query search selection")
			}

			selectedUrl = workflowState.SearchResults[selected-1].URL
			iwf = nil
			continue
		}
	}

	util.LogInfo(
		"That's the one! 🎯 Alright, consider it done. \nI'm now going to 📥 grab that video, ✍️ listen to every word to write it all down, and then pull out the most important points for your summary.\nAlmost there! ⬇️ 🎧 ✍️ 💡",
	)

	outputPath, err := workflow.ExecuteSummarizeWorkflow(
		ctx,
		temporalClient,
		workflow.SummarizeWorkflowParams{URL: selectedUrl},
	)

	if err != nil {
		panic("Failed to execute summarize workflow")
	}

	util.LogInfo(
		"Mission complete! ✨\nThe summary of your chosen video is now ready for you to explore. Check below to discover the highlights! 👇\n\n\n",
	)

	source, err := os.ReadFile(outputPath)

	if err != nil {
		panic(err)
	}

	result := markdown.Render(string(source), 80, 6)
	fmt.Println(string(result))
}
