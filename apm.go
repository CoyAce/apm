// Package apm provides a Go wrapper for WebRTC's AudioProcessing module.
//
// This package enables high-quality audio processing including echo cancellation,
// noise suppression, automatic gain control, and voice activity detection.
//
// Example usage:
//
//	processor, err := apm.New(apm.Config{
//	    SampleRateHz: 48000,
//	    NumChannels:  1,
//	    EchoCancellation: &apm.EchoCancellationConfig{
//	        Enabled:          true,
//	        SuppressionLevel: apm.SuppressionHigh,
//	    },
//	    NoiseSuppression: &apm.NoiseSuppressionConfig{
//	        Enabled: true,
//	        Level:   apm.NoiseSuppressionHigh,
//	    },
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer processor.Close()
//
//	// Process audio frames...
//	cleanSamples, hasVoice, err := processor.ProcessCapture(micSamples)
package apm

import (
	"fmt"
	"sync"
)

// Processor is the main APM instance
type Processor struct {
	handle      *Handle
	config      Config
	mu          sync.Mutex
	numChannels int
}

// New creates a new audio processor with the given configuration
func New(config Config) (*Processor, error) {

	handle, err := Create(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create audio processor: %w", err)
	}

	p := &Processor{
		handle:      handle,
		config:      config,
		numChannels: config.CaptureChannels,
	}

	return p, nil
}

func (p *Processor) Initialize() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.handle == nil {
		return
	}

	p.handle.Initialize()
}

// ProcessCapture processes microphone input (near-end signal)
// Returns processed audio and voice activity detection result
// Input samples should be float32 in range [-1.0, 1.0]
func (p *Processor) ProcessCapture(samples []float32) ([]float32, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.handle == nil {
		return nil, fmt.Errorf("processor is closed")
	}

	expectedLen := p.numChannels * NumSamplesPerFrame
	if len(samples) != expectedLen {
		return nil, fmt.Errorf("expected %d samples (%d channels x %d samples/frame), got %d",
			expectedLen, p.numChannels, NumSamplesPerFrame, len(samples))
	}

	// Make a copy to avoid modifying the input
	output := make([]float32, len(samples))
	copy(output, samples)

	err := p.handle.ProcessCaptureFrame(output, p.numChannels)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// ProcessCaptureInt16 processes microphone input with int16 samples
// This is a convenience method that converts from/to int16 format
func (p *Processor) ProcessCaptureInt16(samples []int16) ([]int16, error) {
	// Convert int16 to float32
	floatSamples := make([]float32, len(samples))
	for i, s := range samples {
		floatSamples[i] = float32(s) / 32768.0
	}

	// Process
	output, err := p.ProcessCapture(floatSamples)
	if err != nil {
		return nil, err
	}

	// Convert back to int16
	int16Output := make([]int16, len(output))
	for i, s := range output {
		// Clamp and convert
		if s > 1.0 {
			s = 1.0
		} else if s < -1.0 {
			s = -1.0
		}
		int16Output[i] = int16(s * 32767.0)
	}

	return int16Output, nil
}

// ProcessRender provides speaker output (far-end signal) for echo cancellation
// Must be called with speaker audio BEFORE ProcessCapture for the corresponding frame
func (p *Processor) ProcessRender(samples []float32) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.handle == nil {
		return fmt.Errorf("processor is closed")
	}

	expectedLen := p.numChannels * NumSamplesPerFrame
	if len(samples) != expectedLen {
		return fmt.Errorf("expected %d samples (%d channels x %d samples/frame), got %d",
			expectedLen, p.numChannels, NumSamplesPerFrame, len(samples))
	}

	// Make a copy to avoid modifying the input
	renderSamples := make([]float32, len(samples))
	copy(renderSamples, samples)

	return p.handle.ProcessRenderFrame(renderSamples, p.numChannels)
}

// ProcessRenderInt16 provides speaker output with int16 samples
// This is a convenience method that converts from int16 format
func (p *Processor) ProcessRenderInt16(samples []int16) error {
	// Convert int16 to float32
	floatSamples := make([]float32, len(samples))
	for i, s := range samples {
		floatSamples[i] = float32(s) / 32768.0
	}

	return p.ProcessRender(floatSamples)
}

// SetStreamDelay updates the estimated delay between render and capture
func (p *Processor) SetStreamDelay(delayMs int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.handle == nil {
		return fmt.Errorf("processor is closed")
	}

	p.handle.SetStreamDelayMs(delayMs)
	return nil
}

func (p *Processor) StreamDelay() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.handle == nil {
		return -1
	}
	return p.handle.StreamDelayMs()
}

// GetStats returns statistics from the last ProcessCapture call
func (p *Processor) GetStats() Stats {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.handle == nil {
		return Stats{}
	}

	cgoStats := p.handle.GetStats()

	return cgoStats
}

// SetOutputMuted signals that output will be muted (hint for AEC/AGC)
func (p *Processor) SetOutputMuted(muted bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.handle != nil {
		p.handle.SetOutputWillBeMuted(muted)
	}
}

// SetKeyPressed signals that a key is being pressed (hint for AEC)
func (p *Processor) SetKeyPressed(pressed bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.handle != nil {
		p.handle.SetStreamKeyPressed(pressed)
	}
}

// Close releases resources associated with the processor
func (p *Processor) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.handle != nil {
		p.handle.Destroy()
		p.handle = nil
	}
	return nil
}
