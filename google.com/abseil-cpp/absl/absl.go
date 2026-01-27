package absl

// #cgo CXXFLAGS: -std=c++17
import "C"

import (
	_ "github.com/CoyAce/apm/google.com/abseil-cpp/absl/base"
	_ "github.com/CoyAce/apm/google.com/abseil-cpp/absl/container"
	_ "github.com/CoyAce/apm/google.com/abseil-cpp/absl/crc"
	_ "github.com/CoyAce/apm/google.com/abseil-cpp/absl/debugging"
	_ "github.com/CoyAce/apm/google.com/abseil-cpp/absl/flags"
	_ "github.com/CoyAce/apm/google.com/abseil-cpp/absl/functional"
	_ "github.com/CoyAce/apm/google.com/abseil-cpp/absl/hash"
	_ "github.com/CoyAce/apm/google.com/abseil-cpp/absl/numeric"
	_ "github.com/CoyAce/apm/google.com/abseil-cpp/absl/profiling"
	_ "github.com/CoyAce/apm/google.com/abseil-cpp/absl/strings"
	_ "github.com/CoyAce/apm/google.com/abseil-cpp/absl/synchronization"
	_ "github.com/CoyAce/apm/google.com/abseil-cpp/absl/time"
)
