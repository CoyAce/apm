// apm_wrapper.cpp - C++ implementation of the C wrapper for WebRTC AudioProcessing

#include <bridge.h>

#include <memory>
#include <vector>

#ifndef WEBRTC_POSIX
#define WEBRTC_POSIX
#endif

#ifndef WEBRTC_AUDIO_PROCESSING_ONLY_BUILD
#define WEBRTC_AUDIO_PROCESSING_ONLY_BUILD
#endif

#include <google.com/webrtc/audio_processing/include/audio_processing.h>
#include <google.com/webrtc/api/audio/builtin_audio_processing_builder.h>
#include <google.com/webrtc/api/environment/environment_factory.h>

namespace {

// Internal structure holding the processor state
    struct AudioProcessor {
        webrtc::scoped_refptr<webrtc::AudioProcessing> processor;
        webrtc::StreamConfig capture_stream_config;
        webrtc::StreamConfig render_stream_config;
        int capture_channels{};
        int render_channels{};

        // Buffers for deinterleaved audio
        std::vector<std::vector<float>> capture_buffer;
        std::vector<std::vector<float>> render_buffer;
        std::vector<float *> capture_ptrs;
        std::vector<float *> render_ptrs;
    };

// Helper to deinterleave audio from interleaved to channel-separated format
    template<typename T>
    typename std::enable_if<std::is_arithmetic<T>::value>::type
    deinterleave(const T *src, std::vector<std::vector<T>> &dst,
                 int num_channels, int num_samples) {
        for (int ch = 0; ch < num_channels; ++ch) {
            for (int i = 0; i < num_samples; ++i) {
                dst[ch][i] = src[i * num_channels + ch];
            }
        }
    }

// Helper to interleave audio from channel-separated to interleaved format
    template<typename T>
    typename std::enable_if<std::is_arithmetic<T>::value>::type
    interleave(const std::vector<std::vector<T>> &src, T *dst,
               int num_channels, int num_samples) {
        for (int ch = 0; ch < num_channels; ++ch) {
            for (int i = 0; i < num_samples; ++i) {
                dst[i * num_channels + ch] = src[ch][i];
            }
        }
    }

    webrtc::AudioProcessing::Config parseConfig(ApmConfig apmConfig) {
        webrtc::AudioProcessing::Config config;

        // High pass filter
        config.high_pass_filter.enabled = apmConfig.high_pass_filter_enabled;

        // Echo cancellation
        config.echo_canceller.enabled = apmConfig.echo_cancellation.enabled;
        config.echo_canceller.mobile_mode = apmConfig.echo_cancellation.mobile_mode;
        if (!config.high_pass_filter.enabled) {
            config.echo_canceller.enforce_high_pass_filtering = false;
        }

        // Gain control
        config.gain_controller1.enabled = apmConfig.gain_control.enabled;
        config.gain_controller1.mode =
                static_cast<webrtc::AudioProcessing::Config::GainController1::Mode>(
                        apmConfig.gain_control.mode);
        config.gain_controller1.target_level_dbfs = apmConfig.gain_control.target_level_dbfs;
        config.gain_controller1.compression_gain_db = apmConfig.gain_control.compression_gain_db;
        config.gain_controller1.enable_limiter = apmConfig.gain_control.enable_limiter != 0;

        // Noise suppression
        config.noise_suppression.enabled = apmConfig.noise_suppression.enabled;
        config.noise_suppression.level =
                static_cast<webrtc::AudioProcessing::Config::NoiseSuppression::Level>(
                        apmConfig.noise_suppression.suppression_level);
        return config;
    }

} // anonymous namespace

extern "C" {

ApmHandle Create(ApmConfig apmConfig, int *error_code) {
    *error_code = 0;
    if (apmConfig.capture_channels == 0 || apmConfig.render_channels == 0) {
        *error_code = webrtc::AudioProcessing::kBadParameterError;
        return nullptr;
    }

    auto *ap = new AudioProcessor;

    webrtc::AudioProcessing::Config config = parseConfig(apmConfig);
    ap->processor = webrtc::BuiltinAudioProcessingBuilder(config)
            .Build(webrtc::CreateEnvironment());
    if (!ap->processor) {
        *error_code = -2;
        delete ap;
        return nullptr;
    }

    ap->capture_stream_config = webrtc::StreamConfig(
            APM_SAMPLE_RATE_HZ, apmConfig.capture_channels);
    ap->render_stream_config = webrtc::StreamConfig(
            APM_SAMPLE_RATE_HZ, apmConfig.render_channels);

    webrtc::ProcessingConfig pconfig = {
            ap->capture_stream_config,
            ap->capture_stream_config,
            ap->render_stream_config,
            ap->render_stream_config,
    };

    ap->capture_channels = apmConfig.capture_channels;
    ap->render_channels = apmConfig.render_channels;

    int code = ap->processor->Initialize(pconfig);
    if (code != webrtc::AudioProcessing::kNoError) {
        *error_code = code;
        delete ap;
        return nullptr;
    }
    // set stream delay
    if (apmConfig.echo_cancellation.enabled) {
        ap->processor->set_stream_delay_ms(apmConfig.echo_cancellation.stream_delay);
    }

    // Initialize buffers
    ap->capture_buffer.resize(apmConfig.capture_channels);
    ap->capture_ptrs.resize(apmConfig.capture_channels);
    for (int i = 0; i < apmConfig.capture_channels; ++i) {
        ap->capture_buffer[i].resize(APM_NUM_SAMPLES_PER_FRAME);
        ap->capture_ptrs[i] = ap->capture_buffer[i].data();
    }

    ap->render_buffer.resize(apmConfig.render_channels);
    ap->render_ptrs.resize(apmConfig.render_channels);
    for (int i = 0; i < apmConfig.render_channels; ++i) {
        ap->render_buffer[i].resize(APM_NUM_SAMPLES_PER_FRAME);
        ap->render_ptrs[i] = ap->render_buffer[i].data();
    }
    return static_cast<ApmHandle>(ap);
}

void Destroy(ApmHandle handle) {
    if (handle) {
        auto *ap = static_cast<AudioProcessor *>(handle);
        delete ap;
    }
}

void Initialize(ApmHandle handle) {
    if (handle) {
        auto *ap = static_cast<AudioProcessor *>(handle);
        ap->processor->Initialize();
    }
}

void ApplyConfig(ApmHandle handle, ApmConfig apmConfig) {
    webrtc::AudioProcessing::Config config = parseConfig(apmConfig);

    auto *ap = static_cast<AudioProcessor *>(handle);
    ap->processor->ApplyConfig(config);
}

int ProcessStream(ApmHandle handle, float *samples, int num_channels) {
    if (!handle || !samples)
        return webrtc::AudioProcessing::kBadParameterError;

    auto *ap = static_cast<AudioProcessor *>(handle);

    if (num_channels != ap->capture_channels)
        return webrtc::AudioProcessing::kBadParameterError;

    // Deinterleave input
    deinterleave(samples, ap->capture_buffer, num_channels, APM_NUM_SAMPLES_PER_FRAME);

    // Process
    int result = ap->processor->ProcessStream(
            ap->capture_ptrs.data(),
            ap->capture_stream_config,
            ap->capture_stream_config,
            ap->capture_ptrs.data());

    if (result == webrtc::AudioProcessing::kNoError) {
        // Interleave output back
        interleave(ap->capture_buffer, samples, num_channels, APM_NUM_SAMPLES_PER_FRAME);
    }

    return result;
}

int ProcessIntStream(ApmHandle handle, int16_t *samples, int num_channels) {
    if (!handle || !samples)
        return webrtc::AudioProcessing::kBadParameterError;

    auto *ap = static_cast<AudioProcessor *>(handle);

    if (num_channels != ap->capture_channels)
        return webrtc::AudioProcessing::kBadParameterError;

    // Process
    int result = ap->processor->ProcessStream(
            samples,
            ap->capture_stream_config,
            ap->capture_stream_config,
            samples);

    return result;
}

int ProcessReverseStream(ApmHandle handle, float *samples, int num_channels) {
    if (!handle || !samples)
        return webrtc::AudioProcessing::kBadParameterError;

    auto *ap = static_cast<AudioProcessor *>(handle);

    if (num_channels != ap->render_channels)
        return webrtc::AudioProcessing::kBadParameterError;;

    // Deinterleave input
    deinterleave(samples, ap->render_buffer, num_channels, APM_NUM_SAMPLES_PER_FRAME);

    // Process reverse stream
    int result = ap->processor->ProcessReverseStream(
            ap->render_ptrs.data(),
            ap->render_stream_config,
            ap->render_stream_config,
            ap->render_ptrs.data());

    if (result == webrtc::AudioProcessing::kNoError) {
        // Interleave output back
        interleave(ap->render_buffer, samples, num_channels, APM_NUM_SAMPLES_PER_FRAME);
    }

    return result;
}

int ProcessReverseIntStream(ApmHandle handle, int16_t *samples, int num_channels) {
    if (!handle || !samples) return -1;

    auto *ap = static_cast<AudioProcessor *>(handle);

    if (num_channels != ap->render_channels)
        return webrtc::AudioProcessing::kBadParameterError;;

    // Process reverse stream
    int result = ap->processor->ProcessReverseStream(
            samples,
            ap->render_stream_config,
            ap->render_stream_config,
            samples);

    return result;
}

ApmStats GetStatistics(ApmHandle handle) {
    ApmStats stats = {};

    if (!handle) return stats;

    auto *ap = static_cast<AudioProcessor *>(handle);
    webrtc::AudioProcessingStats s = ap->processor->GetStatistics();
    // Echo detection
    stats.echo_return_loss = s.echo_return_loss.value_or(0.0);
    stats.echo_return_loss_enhancement = s.echo_return_loss_enhancement.value_or(0.0);
    stats.divergent_filter_fraction = s.divergent_filter_fraction.value_or(0.0);
    stats.residual_echo_likelihood = s.residual_echo_likelihood.value_or(0.0);

    // Delay
    stats.delay_median_ms = s.delay_median_ms.value_or(0);
    stats.delay_std_ms = s.delay_standard_deviation_ms.value_or(0);
    stats.delay_ms = s.delay_ms.value_or(0);

    return stats;
}

void set_stream_delay_ms(ApmHandle handle, int delay_ms) {
    if (!handle) return;

    auto *ap = static_cast<AudioProcessor *>(handle);
    ap->processor->set_stream_delay_ms(delay_ms);
}

int stream_delay_ms(ApmHandle handle) {
    if (!handle) return 0;
    auto *ap = static_cast<AudioProcessor *>(handle);
    return ap->processor->stream_delay_ms();
}

void set_output_will_be_muted(ApmHandle handle, bool muted) {
    if (!handle) return;

    auto *ap = static_cast<AudioProcessor *>(handle);
    ap->processor->set_output_will_be_muted(muted != 0);
}

void set_stream_key_pressed(ApmHandle handle, bool pressed) {
    if (!handle) return;

    auto *ap = static_cast<AudioProcessor *>(handle);
    ap->processor->set_stream_key_pressed(pressed != 0);
}

int is_success(int code) {
    return code == webrtc::AudioProcessing::kNoError ? 1 : 0;
}

int get_num_samples_per_frame(void) {
    return APM_NUM_SAMPLES_PER_FRAME;
}

int get_sample_rate_hz(void) {
    return APM_SAMPLE_RATE_HZ;
}

} // extern "C"
