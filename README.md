# Go WebRTC Audio Processing Module (APM)

Go bindings for [WebRTC's AudioProcessing module](https://webrtc.googlesource.com/src/+/refs/heads/main/modules/audio_processing/), providing production-quality audio processing for voice applications. Built on the extensively tested and optimized WebRTC audio engine, this library is ideal for real-time voice communication scenarios.

## ✨ Key Features

### 🎯 Audio Processing Capabilities
- **Echo Cancellation (AEC)** - Effectively removes speaker audio that leaks into the microphone
- **Noise Suppression (NS)** - Significantly reduces background noise while preserving speech clarity
- **Automatic Gain Control (AGC)** - Intelligently normalizes audio levels automatically
- **High-Pass Filter (HPF)** - Removes DC offset and low-frequency noise

### 🚀 Performance Advantages
- ✅ Powered by industry-leading WebRTC audio processing engine
- ✅ Optimized specifically for real-time voice applications
- ✅ Low-latency, high-performance processing

## 📦 Installation

### System Requirements

- **Operating System**: Linux, macOS, Windows
- **Go Version**: 1.19 or higher
- **Dependencies**: Self-contained WebRTC library (based on commit: `d8dd1409d661`), no additional system dependencies required. 

### Installation Steps

1. **Install the Go package**
```bash
go get github.com/CoyAce/apm
```

2. **Import in your project**
```go
import "github.com/CoyAce/apm"
```

3. **Example**
```go
config := apm.Config{
		CaptureChannels:       1,
		RenderChannels:        1,
		HighPassFilterEnabled: true,
		EchoCancellation:      apm.EchoCancellationConfig{Enabled: true, MobileMode: mobile, StreamDelayMs: 54},
		NoiseSuppression:      apm.NoiseSuppressionConfig{Enabled: true, SuppressionLevel: apm.NsLevelModerate},
		GainControl: apm.GainControlConfig{
			Enabled:                      true,
			InputVolumeControllerEnabled: true,
			HeadroomDB:                   15,
			MaxGainDb:                    50,
		},
	}
processor, _ := apm.New(config)
processor.SetStreamAnalogLevel(180)

// add far end
_ = ae.processor.ProcessRenderInt16(farEnd)

// process near end
ae.processor.SetStreamDelay(config.EchoCancellation.StreamDelayMs)
_ := ae.processor.ProcessCapture(output)
```