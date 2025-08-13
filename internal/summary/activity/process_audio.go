package activity

import (
	"context"
	"fmt"
	"os"

	"github.com/openai/openai-go"
)

type AudioProcessActivities struct {
	opanAPIClient openai.Client
}

func NewAudioProcessActivities(opanAPIClient openai.Client) *AudioProcessActivities {
	return &AudioProcessActivities{
		opanAPIClient,
	}
}

func (apa *AudioProcessActivities) TranscribeAudio(ctx context.Context, filePath string) (string, error) {
	file, err := os.Open(filePath)

	if err != nil {
		return "", err
	}

	transcription, err := apa.opanAPIClient.Audio.Transcriptions.New(ctx, openai.AudioTranscriptionNewParams{
		Model: openai.AudioModelWhisper1,
		File:  file,
	})

	if err != nil {
		return "", err
	}

	return transcription.Text, nil
}

const summaryFormat = "md"

func CreateSummaryOutputFile(ctx context.Context, fileName string) (string, error) {
	dir, _ := os.Getwd()
	outputPath := fmt.Sprintf("%s/output/%s.%s", dir, fileName, summaryFormat)

	transcriptionFile, err := os.Create(outputPath)

	if err != nil {
		return "", err
	}

	defer transcriptionFile.Close()

	return outputPath, nil
}

func (apa *AudioProcessActivities) SummarizeTranscription(
	ctx context.Context,
	transcription string,
) (string, error) {

	prompt := fmt.Sprintf(`Could you provide a concise and comprehensive summary of the given text in markdown format? Send only summary content without any other comments from you like confirmation message or any questions after you finish with content, also don't wrap your answer in "'''markdown'''". The summary should capture the main points and key details of the text while conveying the author's intended meaning accurately. Please ensure that the summary is well-organized and easy to read, with clear headings and subheadings to guide the reader through each section. The length of the summary should be appropriate to capture the main points and key details of the text, without including unnecessary information or becoming overly long. Text: %s`,
		transcription)

	params := openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		Model: openai.ChatModelGPT4o,
		Seed:  openai.Int(0),
	}

	completion, err := apa.opanAPIClient.Chat.Completions.New(ctx, params)

	if err != nil {
		return "", err
	}

	return completion.Choices[0].Message.Content, nil
}

func OutputSummaryToFile(ctx context.Context, summary string, outputFilePath string) (bool, error) {
	err := os.WriteFile(outputFilePath, []byte(summary), 0644)

	if err != nil {
		return false, err
	}

	return true, nil
}
