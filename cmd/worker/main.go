package main

import (
	"api/internal/summary/activity"
	"api/internal/summary/workflow"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/openai/openai-go"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"google.golang.org/protobuf/types/known/durationpb"
)

func RegisterNamespace(ctx context.Context, hostPort, namespace string) error {
	nsClient, err := client.NewNamespaceClient(client.Options{HostPort: hostPort})
	if err != nil {
		return fmt.Errorf("unable to create namespace client: %w", err)
	}
	defer nsClient.Close()

	if _, err := nsClient.Describe(ctx, namespace); err == nil {
		return nil
	}

	retention := &durationpb.Duration{
		Seconds: int64(30) * int64(24*time.Hour) / int64(time.Second),
	}

	req := &workflowservice.RegisterNamespaceRequest{
		Namespace:                        namespace,
		WorkflowExecutionRetentionPeriod: retention,
	}

	if err := nsClient.Register(ctx, req); err != nil {
		return fmt.Errorf("failed to register namespace %q: %w", namespace, err)
	}

	return nil
}

func main() {
	temporalClient, err := client.Dial(client.Options{
		HostPort:  client.DefaultHostPort,
		Namespace: os.Getenv("TEMPORAL_SUMMARIZE_NAMESPACE"),
	})

	if err != nil {
		log.Fatalln("Unable to create Temporal Client", err.Error())
	}

	defer temporalClient.Close()

	err = RegisterNamespace(
		context.Background(),
		client.DefaultHostPort,
		os.Getenv("TEMPORAL_SUMMARIZE_NAMESPACE"),
	)

	if err != nil {
		log.Fatalln("Failed to register Temporal namespace:", err.Error())
	}

	openAPIClient := openai.NewClient()

	w := worker.New(temporalClient, os.Getenv("TEMPORAL_SUMMARIZE_QUEUE_NAME"), worker.Options{})
	/* Register Workflows */
	w.RegisterWorkflow(workflow.SummarizeWorkflow)
	w.RegisterWorkflow(workflow.InteractiveWorkflow)

	/* Register Activities */
	w.RegisterActivity(activity.RetrieveAudio)

	audioProcessingActivities := activity.NewAudioProcessActivities(openAPIClient)
	w.RegisterActivity(audioProcessingActivities)

	w.RegisterActivity(activity.CreateSummaryOutputFile)
	w.RegisterActivity(activity.OutputSummaryToFile)
	w.RegisterActivity(activity.Refine)
	w.RegisterActivity(activity.Search)

	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalln("Worker failed to start", err)
	}
}
