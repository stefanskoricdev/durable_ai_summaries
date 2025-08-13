package workflow

import (
	"api/internal/summary/activity"
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
)

type SummarizeWorkflowParams struct {
	URL string
}

func SummarizeWorkflow(ctx workflow.Context, params SummarizeWorkflowParams) (outputPath string, err error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 5,
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	var retrieveAudioResult activity.RetrieveAudioResult

	err = workflow.ExecuteActivity(ctx, activity.RetrieveAudio, params.URL).Get(ctx, &retrieveAudioResult)

	if err != nil {
		return "", err
	}

	var transcriptionText string

	err = workflow.ExecuteActivity(
		ctx,
		(*activity.AudioProcessActivities).TranscribeAudio,
		retrieveAudioResult.OutputPath,
	).Get(ctx, &transcriptionText)

	if err != nil {
		return "", err
	}

	var futures struct {
		summarizeActivity               workflow.Future
		createSummaryOutputFileActivity workflow.Future
	}

	futures.summarizeActivity = workflow.ExecuteActivity(
		ctx,
		(*activity.AudioProcessActivities).SummarizeTranscription,
		transcriptionText,
	)

	futures.createSummaryOutputFileActivity = workflow.ExecuteActivity(
		ctx,
		activity.CreateSummaryOutputFile,
		retrieveAudioResult.FileName,
	)

	var summary string

	err = futures.summarizeActivity.Get(ctx, &summary)

	if err != nil {
		return "", err
	}

	var summaryOutputPath string

	err = futures.createSummaryOutputFileActivity.Get(ctx, &summaryOutputPath)

	if err != nil {
		return "", err
	}

	var isSummarySuccess bool

	err = workflow.ExecuteActivity(
		ctx,
		activity.OutputSummaryToFile,
		summary,
		summaryOutputPath,
	).Get(ctx, &isSummarySuccess)

	if err != nil {
		return "", err
	}

	outputPath = summaryOutputPath
	return outputPath, nil
}

func ExecuteSummarizeWorkflow(
	ctx context.Context,
	temporalClient client.Client,
	params SummarizeWorkflowParams,
) (string, error) {
	res, err := temporalClient.ExecuteWorkflow(
		ctx,
		client.StartWorkflowOptions{
			ID:        fmt.Sprintf("summarize-workflow-%s", time.Now().Format("20060102150405")),
			TaskQueue: "summarize",
		},
		SummarizeWorkflow,
		SummarizeWorkflowParams{URL: params.URL},
	)

	if err != nil {
		return "", err
	}

	var outputPath string
	res.Get(ctx, &outputPath)

	return outputPath, nil
}
