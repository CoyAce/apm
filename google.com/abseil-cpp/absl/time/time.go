package synchronization

// #cgo CXXFLAGS: -I${SRCDIR}/../..
// #cgo CXXFLAGS: -std=c++17
// #cgo darwin LDFLAGS: -framework CoreFoundation
import "C"
import (
	_ "github.com/CoyAce/apm/google.com/abseil-cpp/absl/time/internal"
)
