//go:build linux && race

package windata

// raceEnabled is true when the binary is built with the race detector,
// whose shadow memory a limit of address space would break.
const raceEnabled = true
