package tui

import "sync"

var (
	mu            sync.Mutex
	builder       func(ioOpts ...func(*IO)) Prompter
	statusBuilder func(ioOpts ...func(*IO)) Status
)

// RegisterBuilder sets the factory used by Default(). Called by backend packages in init().
// This enables swapping backends without import-path changes in calling code.
func RegisterBuilder(fn func(ioOpts ...func(*IO)) Prompter) {
	mu.Lock()
	defer mu.Unlock()
	builder = fn
}

// Default returns a Prompter created by the registered builder.
// Panics if no builder has been registered — import a backend package
// (e.g., _ "github.com/verda-cloud/verdagostack/pkg/tui/bubbletea") to register one.
func Default() Prompter {
	mu.Lock()
	defer mu.Unlock()
	if builder == nil {
		panic("tui: no backend registered — import a backend package")
	}
	return builder()
}

// RegisterStatusBuilder sets the factory used by DefaultStatus().
func RegisterStatusBuilder(fn func(ioOpts ...func(*IO)) Status) {
	mu.Lock()
	defer mu.Unlock()
	statusBuilder = fn
}

// DefaultStatus returns a Status created by the registered builder.
// Panics if no builder has been registered.
func DefaultStatus() Status {
	mu.Lock()
	defer mu.Unlock()
	if statusBuilder == nil {
		panic("tui: no status backend registered — import a backend package")
	}
	return statusBuilder()
}
