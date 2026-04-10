package tui

import "errors"

// ErrInterrupted is returned when the user presses Ctrl+C (hard cancel).
// Distinct from context.Canceled which indicates Esc (soft cancel / go back).
var ErrInterrupted = errors.New("interrupted")
