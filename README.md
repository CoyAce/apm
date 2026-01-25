# Go WebRTC Audio Processing Module (APM)

Go bindings for [WebRTC's AudioProcessing module](https://webrtc.googlesource.com/src/+/refs/heads/main/modules/audio_processing/), providing production-quality audio processing for voice applications. Built on the extensively tested and optimized WebRTC audio engine, this library is ideal for real-time voice communication scenarios.

## ✨ Key Features

### 🎯 Audio Processing Capabilities
- **Echo Cancellation (AEC)** - Effectively removes speaker audio that leaks into the microphone
- **Noise Suppression (NS)** - Significantly reduces background noise while preserving speech clarity
- **Automatic Gain Control (AGC)** - Intelligently normalizes audio levels automatically
- **Voice Activity Detection (VAD)** - Accurately detects speech presence
- **High-Pass Filter (HPF)** - Removes DC offset and low-frequency noise

### 🚀 Performance Advantages
- ✅ Powered by industry-leading WebRTC audio processing engine
- ✅ Optimized specifically for real-time voice applications
- ✅ Low-latency, high-performance processing
- ✅ Minimal memory footprint, suitable for embedded environments

## 📦 Installation

### System Requirements

- **Operating System**: Linux, macOS, Windows
- **Go Version**: 1.19 or higher
- **Dependencies**: Self-contained WebRTC library, no additional system dependencies required

### Installation Steps

1. **Install the Go package**
```bash
go get github.com/CoyAce/apm
```

2. **Import in your project**
```go
import "github.com/CoyAce/apm"
```