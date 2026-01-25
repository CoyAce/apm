// Example: Noise suppression
//
// This example demonstrates how to use go-webrtc-apm for noise suppression.
// It reads audio from a file, applies noise suppression, and writes the result.
//
// Usage:
//
//	go run main.go -input noisy.raw -output clean.raw
//
// The audio files should be raw 48kHz mono float32 PCM.
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	apm "github.com/CoyAce/apm"
)

func main() {
	inputFile := flag.String("input", "", "Input raw audio file (48kHz mono float32)")
	outputFile := flag.String("output", "", "Output raw audio file")
	level := flag.String("level", "high", "Noise suppression level: low, moderate, high, veryhigh")
	flag.Parse()

	if *inputFile == "" || *outputFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Map level string to enum
	var nsLevel apm.NoiseSuppressionLevel
	switch *level {
	case "low":
		nsLevel = apm.NoiseSuppressionLow
	case "moderate":
		nsLevel = apm.NoiseSuppressionModerate
	case "high":
		nsLevel = apm.NoiseSuppressionHigh
	case "veryhigh":
		nsLevel = apm.NoiseSuppressionVeryHigh
	default:
		log.Fatalf("Invalid noise suppression level: %s", *level)
	}

	// Create processor with noise suppression
	processor, err := apm.New(apm.Config{
		NumChannels: 1,
		NoiseSuppression: &apm.NoiseSuppressionConfig{
			Enabled: true,
			Level:   nsLevel,
		},
		HighPassFilter: true, // Also enable high-pass filter to remove DC offset
	})
	if err != nil {
		log.Fatalf("Failed to create processor: %v", err)
	}
	defer processor.Close()

	// Open input file
	in, err := os.Open(*inputFile)
	if err != nil {
		log.Fatalf("Failed to open input file: %v", err)
	}
	defer in.Close()

	// Create output file
	out, err := os.Create(*outputFile)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer out.Close()

	// Process audio frame by frame
	frameSize := apm.NumSamplesPerFrame
	inputBuffer := make([]float32, frameSize)
	frameCount := 0

	fmt.Printf("Processing audio with noise suppression (level: %s)...\n", *level)

	for {
		// Read a frame of audio
		err := binary.Read(in, binary.LittleEndian, inputBuffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Failed to read audio: %v", err)
		}

		// Process the frame
		outputBuffer, _, err := processor.ProcessCapture(inputBuffer)
		if err != nil {
			log.Fatalf("Failed to process frame: %v", err)
		}

		// Write the processed frame
		err = binary.Write(out, binary.LittleEndian, outputBuffer)
		if err != nil {
			log.Fatalf("Failed to write audio: %v", err)
		}

		frameCount++
	}

	duration := float64(frameCount) * float64(apm.FrameMs) / 1000.0
	fmt.Printf("Processed %d frames (%.2f seconds)\n", frameCount, duration)
	fmt.Printf("Output written to: %s\n", *outputFile)
}
