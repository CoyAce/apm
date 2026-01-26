// Package cgo provides low-level CGO bindings to the WebRTC AudioProcessing module.
package apm

/*
#cgo CXXFLAGS:  -I${SRCDIR}/google.com/webrtc
#cgo CXXFLAGS:  -I${SRCDIR}/google.com/abseil-cpp
#cgo CXXFLAGS: -std=c++17
#cgo arm,neon CXXFLAGS: -mfpu=neon -mfloat-abi=hard -DWEBRTC_HAS_NEON
#cgo arm64 CXXFLAGS: -DWEBRTC_HAS_NEON -DWEBRTC_ARCH_ARM64
#cgo arm7 CXXFLAGS: -mfpu=neon -mfloat-abi=hard -DWEBRTC_HAS_NEON -DWEBRTC_ARCH_ARM_V7
#cgo darwin CXXFLAGS: -DWEBRTC_MAC -DWEBRTC_POSIX
#cgo ios CXXFLAGS: -DWEBRTC_IOS -DWEBRTC_MAC -DWEBRTC_POSIX
#cgo linux CXXFLAGS: -DWEBRTC_LINUX -DWEBRTC_POSIX
#cgo android CXXFLAGS: -DWEBRTC_LINUX -DWEBRTC_ANDROID -DWEBRTC_POSIX
#cgo windows CXXFLAGS: -DWEBRTC_WIN
#include <bridge.h>
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"unsafe"

	_ "github.com/CoyAce/apm/google.com/webrtc"
)

// Constants exported from the C library
const (
	SampleRateHz       = C.APM_SAMPLE_RATE_HZ
	FrameMs            = C.APM_FRAME_MS
	NumSamplesPerFrame = C.APM_NUM_SAMPLES_PER_FRAME
)

// NsLevel represents noise suppression levels
type NsLevel int

const (
	NsLevelLow      NsLevel = C.NS_LEVEL_LOW
	NsLevelModerate NsLevel = C.NS_LEVEL_MODERATE
	NsLevelHigh     NsLevel = C.NS_LEVEL_HIGH
	NsLevelVeryHigh NsLevel = C.NS_LEVEL_VERY_HIGH
)

// AgcMode represents automatic gain control modes
type AgcMode int

const (
	AgcModeAdaptiveAnalog  AgcMode = C.AGC_MODE_ADAPTIVE_ANALOG
	AgcModeAdaptiveDigital AgcMode = C.AGC_MODE_ADAPTIVE_DIGITAL
	AgcModeFixedDigital    AgcMode = C.AGC_MODE_FIXED_DIGITAL
)

// VadLikelihood represents voice activity detection likelihood levels
type VadLikelihood int

const (
	VadLikelihoodVeryLow  VadLikelihood = C.VAD_LIKELIHOOD_VERY_LOW
	VadLikelihoodLow      VadLikelihood = C.VAD_LIKELIHOOD_LOW
	VadLikelihoodModerate VadLikelihood = C.VAD_LIKELIHOOD_MODERATE
	VadLikelihoodHigh     VadLikelihood = C.VAD_LIKELIHOOD_HIGH
)

// EchoCancellationConfig holds echo cancellation settings
type EchoCancellationConfig struct {
	Enabled       bool
	MobileMode    bool
	StreamDelayMs int // nil means use delay-agnostic mode
}

// GainControlConfig holds automatic gain control settings
type GainControlConfig struct {
	Enabled           bool
	Mode              AgcMode
	TargetLevelDbfs   int // [0, 31]
	CompressionGainDb int // [0, 90]
	EnableLimiter     bool
}

// NoiseSuppressionConfig holds noise suppression settings
type NoiseSuppressionConfig struct {
	Enabled          bool
	SuppressionLevel NsLevel
}

// Config holds all runtime configuration options
type Config struct {
	EchoCancellation      EchoCancellationConfig
	GainControl           GainControlConfig
	NoiseSuppression      NoiseSuppressionConfig
	HighPassFilterEnabled bool
	CaptureChannels       int
	RenderChannels        int
}

// Stats holds statistics from the audio processor
type Stats struct {
	ResidualEchoLikelihood    float64
	DivergentFilterFraction   float64
	EchoReturnLoss            float64
	EchoReturnLossEnhancement float64
	DelayMedianMs             int
	DelayStdMs                int
	DelayMs                   int
}

// Handle represents an opaque handle to the audio processor
type Handle struct {
	ptr C.ApmHandle
}

// Create creates a new audio processor with the given initialization config
func Create(config Config) (*Handle, error) {
	cConfig := parseConfig(config)

	var errorCode C.int
	ptr := C.Create(cConfig, &errorCode)
	if ptr == nil {
		return nil, fmt.Errorf("failed to create audio processor: error code %d", int(errorCode))
	}

	return &Handle{ptr: ptr}, nil
}

func (h *Handle) Initialize() {
	if h.ptr != nil {
		C.Initialize(h.ptr)
	}
}

// Destroy releases the audio processor resources
func (h *Handle) Destroy() {
	if h.ptr != nil {
		C.Destroy(h.ptr)
		h.ptr = nil
	}
}

// ApplyConfig updates the runtime configuration
func (h *Handle) ApplyConfig(config Config) {
	if h.ptr == nil {
		return
	}

	cConfig := parseConfig(config)

	C.ApplyConfig(h.ptr, cConfig)
}

func parseConfig(config Config) C.ApmConfig {
	cConfig := C.ApmConfig{
		capture_channels: C.int(config.CaptureChannels),
		render_channels:  C.int(config.RenderChannels),
		echo_cancellation: C.ApmEchoCancellation{
			enabled:      C.bool(config.EchoCancellation.Enabled),
			mobile_mode:  C.bool(config.EchoCancellation.MobileMode),
			stream_delay: C.int(config.EchoCancellation.StreamDelayMs),
		},
		gain_control: C.ApmGainControl{
			enabled:             C.bool(config.GainControl.Enabled),
			mode:                C.AgcMode(config.GainControl.Mode),
			target_level_dbfs:   C.int(config.GainControl.TargetLevelDbfs),
			compression_gain_db: C.int(config.GainControl.CompressionGainDb),
			enable_limiter:      C.bool(config.GainControl.EnableLimiter),
		},
		noise_suppression: C.ApmNoiseSuppression{
			enabled:           C.bool(config.NoiseSuppression.Enabled),
			suppression_level: C.NsLevel(config.NoiseSuppression.SuppressionLevel),
		},
		high_pass_filter_enabled: C.bool(config.HighPassFilterEnabled),
	}
	return cConfig
}

// ProcessCaptureFrame processes a capture (microphone) frame
// samples should be interleaved float32 samples with length = numChannels * NumSamplesPerFrame
func (h *Handle) ProcessCaptureFrame(samples []float32, numChannels int) error {
	if h.ptr == nil {
		return fmt.Errorf("audio processor not initialized")
	}

	expectedLen := numChannels * NumSamplesPerFrame
	if len(samples) != expectedLen {
		return fmt.Errorf("expected %d samples, got %d", expectedLen, len(samples))
	}

	result := C.ProcessStream(
		h.ptr,
		(*C.float)(unsafe.Pointer(&samples[0])),
		C.int(numChannels),
	)

	if C.is_success(result) == 0 {
		return fmt.Errorf("failed to process capture frame: error code %d", int(result))
	}

	return nil
}

// ProcessRenderFrame processes a render (speaker) frame for echo cancellation
// samples should be interleaved float32 samples with length = numChannels * NumSamplesPerFrame
func (h *Handle) ProcessRenderFrame(samples []float32, numChannels int) error {
	if h.ptr == nil {
		return fmt.Errorf("audio processor not initialized")
	}

	expectedLen := numChannels * NumSamplesPerFrame
	if len(samples) != expectedLen {
		return fmt.Errorf("expected %d samples, got %d", expectedLen, len(samples))
	}

	result := C.ProcessReverseStream(
		h.ptr,
		(*C.float)(unsafe.Pointer(&samples[0])),
		C.int(numChannels),
	)

	if C.is_success(result) == 0 {
		return fmt.Errorf("failed to process render frame: error code %d", int(result))
	}

	return nil
}

// GetStats returns statistics from the last capture frame processing
func (h *Handle) GetStats() Stats {
	var stats Stats

	if h.ptr == nil {
		return stats
	}

	cStats := C.GetStatistics(h.ptr)

	stats.ResidualEchoLikelihood = float64(cStats.residual_echo_likelihood)
	stats.DivergentFilterFraction = float64(cStats.divergent_filter_fraction)
	stats.EchoReturnLoss = float64(cStats.echo_return_loss)
	stats.EchoReturnLossEnhancement = float64(cStats.echo_return_loss_enhancement)
	stats.DelayMedianMs = int(cStats.delay_median_ms)
	stats.DelayStdMs = int(cStats.delay_std_ms)
	stats.DelayMs = int(cStats.delay_ms)

	return stats
}

// SetStreamDelayMs sets the delay between render and capture streams
func (h *Handle) SetStreamDelayMs(delayMs int) {
	if h.ptr == nil {
		return
	}
	C.set_stream_delay_ms(h.ptr, C.int(delayMs))
}

func (h *Handle) StreamDelayMs() int {
	if h.ptr == nil {
		return 0
	}
	return int(C.stream_delay_ms(h.ptr))
}

// SetOutputWillBeMuted signals that output will be muted (hint for AEC/AGC)
func (h *Handle) SetOutputWillBeMuted(muted bool) {
	if h.ptr == nil {
		return
	}
	C.set_output_will_be_muted(h.ptr, C.bool(muted))
}

// SetStreamKeyPressed signals that a key is being pressed (hint for AEC)
func (h *Handle) SetStreamKeyPressed(pressed bool) {
	if h.ptr == nil {
		return
	}
	C.set_stream_key_pressed(h.ptr, C.bool(pressed))
}

// GetNumSamplesPerFrame returns the number of samples per frame
func GetNumSamplesPerFrame() int {
	return int(C.get_num_samples_per_frame())
}

// GetSampleRateHz returns the sample rate in Hz
func GetSampleRateHz() int {
	return int(C.get_sample_rate_hz())
}
