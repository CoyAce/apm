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
	handle *Handle
	config Config
	mu     sync.Mutex
}

// New creates a new audio processor with the given configuration
func New(config Config) (*Processor, error) {

	handle, err := Create(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create audio processor: %w", err)
	}

	p := &Processor{
		handle: handle,
		config: config,
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
func (p *Processor) ProcessCapture(samples []float32) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.handle == nil {
		return fmt.Errorf("processor is closed")
	}

	return p.handle.ProcessCaptureFrame(samples, p.config.CaptureChannels)
}

// ProcessCaptureInt16 processes microphone input with int16 samples
func (p *Processor) ProcessCaptureInt16(samples []int16) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.handle == nil {
		return fmt.Errorf("processor is closed")
	}
	// Process
	return p.handle.ProcessCaptureIntFrame(samples, p.config.CaptureChannels)
}

// ProcessRender provides speaker output (far-end signal) for echo cancellation
// Must be called with speaker audio BEFORE ProcessCapture for the corresponding frame
func (p *Processor) ProcessRender(samples []float32) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.handle == nil {
		return fmt.Errorf("processor is closed")
	}

	return p.handle.ProcessRenderFrame(samples, p.config.RenderChannels)
}

// ProcessRenderInt16 provides speaker output with int16 samples
func (p *Processor) ProcessRenderInt16(samples []int16) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.handle == nil {
		return fmt.Errorf("processor is closed")
	}
	return p.handle.ProcessRenderIntFrame(samples, p.config.RenderChannels)
}

func (p *Processor) SetStreamAnalogLevel(level int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.handle == nil {
		return
	}

	p.handle.SetStreamAnalogLevel(level)
}

func (p *Processor) RecommendedStreamAnalogLevel() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.handle == nil {
		return -1
	}
	return p.handle.RecommendedStreamAnalogLevel()
}

// SetStreamDelay updates the estimated delay between render and capture
func (p *Processor) SetStreamDelay(delayMs int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.handle == nil {
		return
	}

	p.handle.SetStreamDelayMs(delayMs)
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
