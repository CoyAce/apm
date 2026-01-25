package internal

// #cgo CXXFLAGS: -I${SRCDIR}/../../..
// #cgo CXXFLAGS: -std=c++17
import "C"

import (
	_ "github.com/CoyAce/apm/google.com/abseil-cpp/absl/strings/internal/str_format"
)
