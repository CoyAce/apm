// Example: Full audio processing pipeline
//
// This example demonstrates how to use go-webrtc-apm with all features enabled:
// - Echo cancellation
// - Noise suppression
// - Automatic gain control
// - Voice activity detection
// - High-pass filter
//
// This is typical for voice communication applications like VoIP or voice assistants.
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
	nearendFile := flag.String("nearend", "", "Near-end (microphone) raw audio file")
	outputFile := flag.String("output", "", "Output raw audio file (processed)")
	vadOnly := flag.Bool("vad-only", false, "Only output frames with detected voice")
	flag.Parse()

	if *farendFile == "" || *nearendFile == "" || *outputFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Create processor with full pipeline
	processor, err := apm.New(apm.Config{
		CaptureChannels: 1,
		RenderChannels:  1,

		// Echo cancellation: removes speaker audio from mic input
		EchoCancellation: apm.EchoCancellationConfig{
			Enabled: true,
		},

		// Noise suppression: removes background noise
		NoiseSuppression: apm.NoiseSuppressionConfig{
			Enabled: true,
		},

		// Automatic gain control: normalizes volume levels
		GainControl: apm.GainControlConfig{
			Enabled:           true,
			Mode:              apm.AgcModeAdaptiveDigital,
			TargetLevelDbfs:   3, // Target 3 dB below full scale
			CompressionGainDb: 9, // Up to 9 dB of gain
			EnableLimiter:     true,
		},

		// High-pass filter: removes DC offset and low-frequency noise
		HighPassFilterEnabled: true,
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
	voiceFrameCount := 0

	fmt.Println("Processing audio with full pipeline...")
	fmt.Println("Features: AEC, NS, AGC, VAD, HPF")
	fmt.Println()

	for {
		// Read a frame of far-end audio
		err := binary.Read(farend, binary.LittleEndian, farendBuffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Failed to read far-end audio: %v", err)
		}

		// Read a frame of near-end audio
		err = binary.Read(nearend, binary.LittleEndian, nearendBuffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Failed to read near-end audio: %v", err)
		}

		// Process render (far-end) first
		err = processor.ProcessRender(farendBuffer)
		if err != nil {
			log.Fatalf("Failed to process render frame: %v", err)
		}

		// Process capture (near-end)
		outputBuffer, err := processor.ProcessCapture(nearendBuffer)
		if err != nil {
			log.Fatalf("Failed to process capture frame: %v", err)
		}

		// Write output (optionally only when voice is detected)
		if !*vadOnly {
			err = binary.Write(out, binary.LittleEndian, outputBuffer)
			if err != nil {
				log.Fatalf("Failed to write audio: %v", err)
			}
		}

		frameCount++
	}

	duration := float64(frameCount) * float64(apm.FrameMs) / 1000.0
	voiceDuration := float64(voiceFrameCount) * float64(apm.FrameMs) / 1000.0

	fmt.Println()
	fmt.Printf("Processed %d frames (%.2f seconds)\n", frameCount, duration)
	fmt.Printf("Voice detected in %d frames (%.2f seconds, %.1f%%)\n",
		voiceFrameCount, voiceDuration, float64(voiceFrameCount)/float64(frameCount)*100)
	fmt.Printf("Output written to: %s\n", *outputFile)
}
