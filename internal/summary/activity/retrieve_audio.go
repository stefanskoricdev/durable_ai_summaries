package activity

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/google/uuid"
)

const format = "mp3"

type RetrieveAudioResult struct {
	OutputPath string
	FileName   string
}

func RetrieveAudio(URL string) (*RetrieveAudioResult, error) {
	dir, _ := os.Getwd()

	outputDir := fmt.Sprintf("%s/output", dir)
	fileName := uuid.New().String()
	outputPath := fmt.Sprintf("%s/%s.%s", outputDir, fileName, format)

	cmd := exec.Command("yt-dlp",
		"-f", "bestaudio",
		"--extract-audio",
		"--audio-format", format,
		"--audio-quality", "0", // best quality
		"-o", outputPath,
		URL,
	)

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return &RetrieveAudioResult{
		OutputPath: outputPath,
		FileName:   fileName,
	}, nil
}
