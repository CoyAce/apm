//go:build darwin || ios

package system

// #cgo CFLAGS: -I${SRCDIR}/../..
// #cgo CFLAGS: -x objective-c
// #include "darwin/gcd_helpers.m"
import "C"
