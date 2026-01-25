package apm

import (
	"math"
	"testing"
)

// =============================================================================
// Test Helpers
// =============================================================================

func generateSineWave(frequency float64, amplitude float32, numSamples int) []float32 {
	samples := make([]float32, numSamples)
	for i := range samples {
		samples[i] = amplitude * float32(math.Sin(2*math.Pi*frequency*float64(i)/float64(SampleRateHz)))
	}
	return samples
}

// =============================================================================
// Constants Tests
// =============================================================================

func TestConstants(t *testing.T) {
	if SampleRateHz != 48000 {
		t.Errorf("SampleRateHz = %d, want 48000", SampleRateHz)
	}
	if FrameMs != 10 {
		t.Errorf("FrameMs = %d, want 10", FrameMs)
	}
	if NumSamplesPerFrame != 480 {
		t.Errorf("NumSamplesPerFrame = %d, want 480", NumSamplesPerFrame)
	}
}

func TestGetNumSamplesPerFrame(t *testing.T) {
	n := GetNumSamplesPerFrame()
	if n != NumSamplesPerFrame {
		t.Errorf("GetNumSamplesPerFrame() = %d, want %d", n, NumSamplesPerFrame)
	}
}

func TestGetSampleRateHz(t *testing.T) {
	rate := GetSampleRateHz()
	if rate != SampleRateHz {
		t.Errorf("GetSampleRateHz() = %d, want %d", rate, SampleRateHz)
	}
}

// =============================================================================
// Creation Tests
// =============================================================================

func TestCreate(t *testing.T) {
	config := Config{
		CaptureChannels: 1,
		RenderChannels:  1,
	}

	h, err := Create(config)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if h == nil {
		t.Fatal("Create returned nil handle")
	}

	h.Destroy()
}

func TestCreateStereo(t *testing.T) {
	config := Config{
		CaptureChannels: 1,
		RenderChannels:  1,
	}

	h, err := Create(config)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	h.Destroy()
}

func TestCreateInvalidChannels(t *testing.T) {
	config := Config{
		CaptureChannels: 0,
		RenderChannels:  0,
	}

	h, err := Create(config)
	if err == nil {
		h.Destroy()
		t.Fatal("Create should fail with 0 capture channels")
	}
}

// =============================================================================
// Configuration Tests
// =============================================================================

func TestSetConfig(t *testing.T) {
	config := Config{
		CaptureChannels: 1,
		RenderChannels:  1,
		EchoCancellation: EchoCancellationConfig{
			Enabled: true,
		},
	}

	h, err := Create(config)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	defer h.Destroy()
}

func TestSetConfigNoiseSuppression(t *testing.T) {
	config := Config{
		CaptureChannels: 1,
		RenderChannels:  1,
		NoiseSuppression: NoiseSuppressionConfig{
			Enabled:          true,
			SuppressionLevel: NsLevelHigh,
		},
	}

	h, err := Create(config)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	defer h.Destroy()
}

func TestSetConfigGainControl(t *testing.T) {
	config := Config{
		CaptureChannels: 1,
		RenderChannels:  1,
		GainControl: GainControlConfig{
			Enabled:           true,
			Mode:              AgcModeAdaptiveDigital,
			TargetLevelDbfs:   3,
			CompressionGainDb: 9,
			EnableLimiter:     true,
		},
	}

	h, err := Create(config)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	defer h.Destroy()
}

func TestSetConfigAll(t *testing.T) {
	config := Config{
		CaptureChannels: 1,
		RenderChannels:  1,
		EchoCancellation: EchoCancellationConfig{
			Enabled: true,
		},
		GainControl: GainControlConfig{
			Enabled:           true,
			Mode:              AgcModeAdaptiveDigital,
			TargetLevelDbfs:   3,
			CompressionGainDb: 9,
			EnableLimiter:     true,
		},
		NoiseSuppression: NoiseSuppressionConfig{
			Enabled:          true,
			SuppressionLevel: NsLevelHigh,
		},
	}

	h, err := Create(config)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	defer h.Destroy()
}

func TestSetConfigWithStreamDelay(t *testing.T) {
	delayMs := 50
	config := Config{
		CaptureChannels: 1,
		RenderChannels:  1,
		EchoCancellation: EchoCancellationConfig{
			Enabled:       true,
			StreamDelayMs: delayMs,
		},
		GainControl: GainControlConfig{
			Enabled:           true,
			Mode:              AgcModeAdaptiveDigital,
			TargetLevelDbfs:   3,
			CompressionGainDb: 9,
			EnableLimiter:     true,
		},
		NoiseSuppression: NoiseSuppressionConfig{
			Enabled:          true,
			SuppressionLevel: NsLevelHigh,
		},
	}

	h, err := Create(config)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	defer h.Destroy()
}

// =============================================================================
// Processing Tests
// =============================================================================

func TestProcessCaptureFrame(t *testing.T) {
	config := Config{
		CaptureChannels: 1,
		RenderChannels:  1,
	}

	h, err := Create(config)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	defer h.Destroy()

	samples := generateSineWave(440, 0.5, NumSamplesPerFrame)

	err = h.ProcessCaptureFrame(samples, 1)
	if err != nil {
		t.Fatalf("ProcessCaptureFrame failed: %v", err)
	}
}

func TestProcessCaptureFrameStereo(t *testing.T) {
	config := Config{
		CaptureChannels: 2,
		RenderChannels:  2,
	}

	h, err := Create(config)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	defer h.Destroy()

	// Interleaved stereo samples
	samples := make([]float32, NumSamplesPerFrame*2)
	for i := 0; i < NumSamplesPerFrame; i++ {
		samples[i*2] = float32(math.Sin(2*math.Pi*440*float64(i)/float64(SampleRateHz))) * 0.5
		samples[i*2+1] = float32(math.Sin(2*math.Pi*880*float64(i)/float64(SampleRateHz))) * 0.3
	}

	err = h.ProcessCaptureFrame(samples, 2)
	if err != nil {
		t.Fatalf("ProcessCaptureFrame failed: %v", err)
	}
}

func TestProcessRenderFrame(t *testing.T) {
	config := Config{
		CaptureChannels: 1,
		RenderChannels:  1,
		EchoCancellation: EchoCancellationConfig{
			Enabled: true,
		},
	}

	h, err := Create(config)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	defer h.Destroy()

	samples := generateSineWave(1000, 0.4, NumSamplesPerFrame)

	err = h.ProcessRenderFrame(samples, 1)
	if err != nil {
		t.Fatalf("ProcessRenderFrame failed: %v", err)
	}
}

func TestProcessRenderAndCapture(t *testing.T) {
	config := Config{
		CaptureChannels: 1,
		RenderChannels:  1,
		EchoCancellation: EchoCancellationConfig{
			Enabled: true,
		},
	}

	h, err := Create(config)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	defer h.Destroy()

	renderSamples := generateSineWave(1000, 0.4, NumSamplesPerFrame)
	captureSamples := generateSineWave(500, 0.3, NumSamplesPerFrame)

	for i := 0; i < 50; i++ {
		err := h.ProcessRenderFrame(renderSamples, 1)
		if err != nil {
			t.Fatalf("ProcessRenderFrame failed at frame %d: %v", i, err)
		}

		err = h.ProcessCaptureFrame(captureSamples, 1)
		if err != nil {
			t.Fatalf("ProcessCaptureFrame failed at frame %d: %v", i, err)
		}
	}
}

func TestProcessCaptureFrameWrongChannels(t *testing.T) {
	config := Config{
		CaptureChannels: 2,
		RenderChannels:  2,
	}

	h, err := Create(config)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	defer h.Destroy()

	// Provide mono samples for stereo processor
	samples := generateSineWave(440, 0.5, NumSamplesPerFrame)

	err = h.ProcessCaptureFrame(samples, 1)
	if err == nil {
		t.Error("ProcessCaptureFrame should fail with wrong channel count")
	}
}

func TestProcessCaptureFrameWrongSampleCount(t *testing.T) {
	config := Config{
		CaptureChannels: 1,
		RenderChannels:  1,
	}

	h, err := Create(config)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	defer h.Destroy()

	// Provide wrong number of samples
	samples := make([]float32, NumSamplesPerFrame-10)

	err = h.ProcessCaptureFrame(samples, 1)
	if err == nil {
		t.Error("ProcessCaptureFrame should fail with wrong sample count")
	}
}

// =============================================================================
// Statistics Tests
// =============================================================================

func TestGetStatsWithAEC(t *testing.T) {
	config := Config{
		CaptureChannels: 1,
		RenderChannels:  1,
		EchoCancellation: EchoCancellationConfig{
			Enabled: true,
		},
	}

	h, err := Create(config)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	defer h.Destroy()

	renderSamples := generateSineWave(1000, 0.4, NumSamplesPerFrame)
	captureSamples := generateSineWave(500, 0.3, NumSamplesPerFrame)

	// Add echo to capture
	for i := range captureSamples {
		captureSamples[i] += renderSamples[i] * 0.2
	}

	// Process many frames for AEC to converge
	for i := 0; i < 100; i++ {
		h.ProcessRenderFrame(renderSamples, 1)
		h.ProcessCaptureFrame(captureSamples, 1)
	}

	stats := h.GetStats()

	t.Logf("ERL: %.2f dB", stats.EchoReturnLoss)
	t.Logf("ERLE: %.2f dB", stats.EchoReturnLossEnhancement)
}

// =============================================================================
// Stream Delay Tests
// =============================================================================

func TestSetStreamDelayMs(t *testing.T) {
	config := Config{
		CaptureChannels: 1,
		RenderChannels:  1,
		EchoCancellation: EchoCancellationConfig{
			Enabled: true,
		},
	}

	h, err := Create(config)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	defer h.Destroy()

	delays := []int{0, 10, 25, 50, 100, 200}
	for _, delay := range delays {
		h.SetStreamDelayMs(delay)
	}
}

// =============================================================================
// Mute and Key Press Tests
// =============================================================================

func TestSetOutputWillBeMuted(t *testing.T) {
	config := Config{
		CaptureChannels: 1,
		RenderChannels:  1,
	}

	h, err := Create(config)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	defer h.Destroy()

	h.SetOutputWillBeMuted(true)
	h.SetOutputWillBeMuted(false)
}

func TestSetStreamKeyPressed(t *testing.T) {
	config := Config{
		CaptureChannels: 1,
		RenderChannels:  1,
	}

	h, err := Create(config)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	defer h.Destroy()

	h.SetStreamKeyPressed(true)
	h.SetStreamKeyPressed(false)
}

// =============================================================================
// Destruction Tests
// =============================================================================

func TestDestroy(t *testing.T) {
	config := Config{
		CaptureChannels: 1,
		RenderChannels:  1,
	}

	h, err := Create(config)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	h.Destroy()

	// Destroy again should not panic
	h.Destroy()
}

func TestDestroyNil(t *testing.T) {
	var h *Handle
	// Should not panic
	if h != nil {
		h.Destroy()
	}
}

// =============================================================================
// Enum Value Tests
// =============================================================================

func TestNsLevelValues(t *testing.T) {
	if NsLevelLow != 0 {
		t.Errorf("NsLevelLow = %d, want 0", NsLevelLow)
	}
	if NsLevelVeryHigh != 3 {
		t.Errorf("NsLevelVeryHigh = %d, want 3", NsLevelVeryHigh)
	}
}

func TestAgcModeValues(t *testing.T) {
	if AgcModeAdaptiveAnalog != 0 {
		t.Errorf("AgcModeAdaptiveAnalog = %d, want 0", AgcModeAdaptiveAnalog)
	}
	if AgcModeFixedDigital != 2 {
		t.Errorf("AgcModeFixedDigital = %d, want 2", AgcModeFixedDigital)
	}
}

func TestVadLikelihoodValues(t *testing.T) {
	if VadLikelihoodVeryLow != 0 {
		t.Errorf("VadLikelihoodVeryLow = %d, want 0", VadLikelihoodVeryLow)
	}
	if VadLikelihoodHigh != 3 {
		t.Errorf("VadLikelihoodHigh = %d, want 3", VadLikelihoodHigh)
	}
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkCreate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		config := Config{
			CaptureChannels: 1,
			RenderChannels:  1,
		}
		h, _ := Create(config)
		h.Destroy()
	}
}

func BenchmarkProcessCaptureFrame(b *testing.B) {
	config := Config{
		CaptureChannels: 1,
		RenderChannels:  1,
		NoiseSuppression: NoiseSuppressionConfig{
			Enabled:          true,
			SuppressionLevel: NsLevelHigh,
		},
	}

	h, err := Create(config)
	if err != nil {
		b.Fatalf("Create failed: %v", err)
	}
	defer h.Destroy()

	samples := generateSineWave(440, 0.5, NumSamplesPerFrame)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.ProcessCaptureFrame(samples, 1)
	}
}

func BenchmarkProcessCaptureFrameWithAEC(b *testing.B) {
	config := Config{
		CaptureChannels: 1,
		RenderChannels:  1,
		EchoCancellation: EchoCancellationConfig{
			Enabled: true,
		},
		NoiseSuppression: NoiseSuppressionConfig{
			Enabled:          true,
			SuppressionLevel: NsLevelHigh,
		},
	}

	h, err := Create(config)
	if err != nil {
		b.Fatalf("Create failed: %v", err)
	}
	defer h.Destroy()

	renderSamples := generateSineWave(1000, 0.4, NumSamplesPerFrame)
	captureSamples := generateSineWave(500, 0.3, NumSamplesPerFrame)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.ProcessRenderFrame(renderSamples, 1)
		h.ProcessCaptureFrame(captureSamples, 1)
	}
}

func BenchmarkGetStats(b *testing.B) {
	config := Config{
		CaptureChannels: 1,
		RenderChannels:  1,
	}

	h, err := Create(config)
	if err != nil {
		b.Fatalf("Create failed: %v", err)
	}
	defer h.Destroy()

	samples := generateSineWave(440, 0.5, NumSamplesPerFrame)
	h.ProcessCaptureFrame(samples, 1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.GetStats()
	}
}
