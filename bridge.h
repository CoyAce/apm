// apm_wrapper.h - C wrapper for WebRTC AudioProcessing module
// This provides a C interface for CGO to bind to.

#ifndef APM_WRAPPER_H
#define APM_WRAPPER_H
#include <stdint.h>
#include <stdbool.h>
#ifdef __cplusplus
extern "C" {
#endif

// Opaque handle to the audio processor
typedef void* ApmHandle;

// Sample rate must be one of: 8000, 16000, 32000, 48000 Hz
// Frame duration is fixed at 10ms
// So NUM_SAMPLES_PER_FRAME = SAMPLE_RATE_HZ * 10 / 1000

// Constants
#define APM_SAMPLE_RATE_HZ 48000
#define APM_FRAME_MS 10
#define APM_NUM_SAMPLES_PER_FRAME (APM_SAMPLE_RATE_HZ * APM_FRAME_MS / 1000)

// Noise suppression levels
typedef enum {
    NS_LEVEL_LOW = 0,
    NS_LEVEL_MODERATE = 1,
    NS_LEVEL_HIGH = 2,
    NS_LEVEL_VERY_HIGH = 3
} NsLevel;

// Gain control modes
typedef enum {
    AGC_MODE_ADAPTIVE_ANALOG = 0,
    AGC_MODE_ADAPTIVE_DIGITAL = 1,
    AGC_MODE_FIXED_DIGITAL = 2
} AgcMode;

// Voice detection likelihood
typedef enum {
    VAD_LIKELIHOOD_VERY_LOW = 0,
    VAD_LIKELIHOOD_LOW = 1,
    VAD_LIKELIHOOD_MODERATE = 2,
    VAD_LIKELIHOOD_HIGH = 3
} VadLikelihood;

// Echo cancellation configuration
typedef struct {
    bool enabled;
    bool mobile_mode;
    int stream_delay;
} ApmEchoCancellation;

// Gain control configuration
typedef struct {
    bool enabled;
    AgcMode mode;
    int target_level_dbfs;      // [0, 31]
    int compression_gain_db;    // [0, 90]
    bool enable_limiter;
} ApmGainControl;

// Noise suppression configuration
typedef struct {
    bool enabled;
    NsLevel suppression_level;
} ApmNoiseSuppression;

// Voice detection configuration
typedef struct {
    bool enabled;
    VadLikelihood detection_likelihood;
} ApmVoiceDetection;

// Full runtime configuration
typedef struct {
    ApmEchoCancellation echo_cancellation;
    ApmGainControl gain_control;
    ApmNoiseSuppression noise_suppression;
    ApmVoiceDetection voice_detection;
    bool high_pass_filter_enabled;
    int capture_channels;
    int render_channels;
} ApmConfig;

// Statistics from processing
typedef struct {
    double residual_echo_likelihood;

    double echo_return_loss;
    double echo_return_loss_enhancement;
    double divergent_filter_fraction;

    int delay_median_ms;
    int delay_std_ms;
    int delay_ms;
} ApmStats;

// Create a new audio processor instance
// Returns NULL on failure, sets error code
ApmHandle Create(ApmConfig apmConfig, int* error_code);

void Initialize(ApmHandle handle);

// ApplyConfig to audio processor
void ApplyConfig(ApmHandle handle, ApmConfig apmConfig);

// Destroy an audio processor instance
void Destroy(ApmHandle handle);

// Process a capture (microphone) frame
// samples: interleaved float samples, length = num_channels * NUM_SAMPLES_PER_FRAME
// Returns 0 on success, error code on failure
int ProcessStream(ApmHandle handle, float* samples, int num_channels);
int ProcessIntStream(ApmHandle handle, int16_t * samples, int num_channels);

// Process a render (speaker) frame for echo cancellation reference
// samples: interleaved float samples, length = num_channels * NUM_SAMPLES_PER_FRAME
// Returns 0 on success, error code on failure
int ProcessReverseStream(ApmHandle handle, float* samples, int num_channels);
int ProcessReverseIntStream(ApmHandle handle, int16_t* samples, int num_channels);

// Get statistics from the last capture frame processing
ApmStats GetStatistics(ApmHandle handle);

// Set stream delay for echo cancellation (in milliseconds)
void set_stream_delay_ms(ApmHandle handle, int delay_ms);
int stream_delay_ms(ApmHandle handle);

// Signal that output will be muted (hint for AEC/AGC)
void set_output_will_be_muted(ApmHandle handle, bool muted);

// Signal that a key is being pressed (hint for AEC)
void set_stream_key_pressed(ApmHandle handle, bool pressed);

// Check if a return code indicates success
int is_success(int code);

// Get the number of samples per frame for the fixed sample rate
int get_num_samples_per_frame(void);

// Get the sample rate in Hz
int get_sample_rate_hz(void);

#ifdef __cplusplus
}
#endif

#endif // APM_WRAPPER_H
