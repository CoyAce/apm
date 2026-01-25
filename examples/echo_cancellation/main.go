// Example: Echo cancellation
//
// This example demonstrates how to use go-webrtc-apm for echo cancellation.
// It simulates a scenario where speaker output (far-end) leaks into the microphone (near-end).
//
// Usage:
//
//	go run main.go -farend speaker.raw -nearend mic.raw -output clean.raw
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

	"github.com/CoyAce/apm"
)

func main() {
	farendFile := flag.String("farend", "", "Far-end (speaker) raw audio file (48kHz mono float32)")
	nearendFile := flag.String("nearend", "", "Near-end (microphone) raw audio file with echo")
	outputFile := flag.String("output", "", "Output raw audio file (echo cancelled)")
	delayMs := flag.Int("delay", 0, "Stream delay in milliseconds (0 for delay-agnostic mode)")
	flag.Parse()

	if *farendFile == "" || *nearendFile == "" || *outputFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Configure echo cancellation
	aecConfig := apm.EchoCancellationConfig{
		Enabled: true,
	}

	if *delayMs > 0 {
		aecConfig.StreamDelayMs = *delayMs
		fmt.Printf("Using fixed stream delay: %d ms\n", *delayMs)
	} else {
		fmt.Println("Using delay-agnostic echo cancellation")
	}

	// Create processor with echo cancellation
	processor, err := apm.New(apm.Config{
		CaptureChannels:  1,
		EchoCancellation: aecConfig,
	})
	if err != nil {
		log.Fatalf("Failed to create processor: %v", err)
	}
	defer processor.Close()

	// Open input files
	farend, err := os.Open(*farendFile)
	if err != nil {
		log.Fatalf("Failed to open far-end file: %v", err)
	}
	defer farend.Close()

	nearend, err := os.Open(*nearendFile)
	if err != nil {
		log.Fatalf("Failed to open near-end file: %v", err)
	}
	defer nearend.Close()

	// Create output file
	out, err := os.Create(*outputFile)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer out.Close()

	// Process audio frame by frame
	frameSize := apm.NumSamplesPerFrame
	farendBuffer := make([]float32, frameSize)
	nearendBuffer := make([]float32, frameSize)
	frameCount := 0

	fmt.Println("Processing audio with echo cancellation...")

	for {
		// Read a frame of far-end audio (speaker output)
		err := binary.Read(farend, binary.LittleEndian, farendBuffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Failed to read far-end audio: %v", err)
		}

		// Read a frame of near-end audio (microphone input with echo)
		err = binary.Read(nearend, binary.LittleEndian, nearendBuffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Failed to read near-end audio: %v", err)
		}

		// IMPORTANT: Process render (far-end) BEFORE capture (near-end)
		// This provides the reference signal for echo cancellation
		err = processor.ProcessRender(farendBuffer)
		if err != nil {
			log.Fatalf("Failed to process render frame: %v", err)
		}

		// Process capture (near-end) to remove echo
		outputBuffer, err := processor.ProcessCapture(nearendBuffer)
		if err != nil {
			log.Fatalf("Failed to process capture frame: %v", err)
		}

		// Write the processed frame
		err = binary.Write(out, binary.LittleEndian, outputBuffer)
		if err != nil {
			log.Fatalf("Failed to write audio: %v", err)
		}

		frameCount++

		// Print stats every second
		if frameCount%100 == 0 {
			stats := processor.GetStats()
			fmt.Printf("Frame %d: ERL=%.1f dB", frameCount, stats.EchoReturnLoss)
			fmt.Printf(", ERLE=%.1f dB", stats.EchoReturnLossEnhancement)
			fmt.Println()
		}
	}

	duration := float64(frameCount) * float64(apm.FrameMs) / 1000.0
	fmt.Printf("\nProcessed %d frames (%.2f seconds)\n", frameCount, duration)
	fmt.Printf("Output written to: %s\n", *outputFile)
}
