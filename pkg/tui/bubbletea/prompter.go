package bubbletea

import (
	"io"
	"os"

	"github.com/verda-cloud/verdagostack/pkg/tui"
)

// Prompter implements tui.Prompter using Bubbletea.
type Prompter struct {
	in     io.Reader
	out    io.Writer
	errOut io.Writer
}

// New creates a Bubbletea-backed Prompter.
func New(ioOpts ...func(*Prompter)) *Prompter {
	p := &Prompter{
		in:     os.Stdin,
		out:    os.Stdout,
		errOut: os.Stderr,
	}
	for _, o := range ioOpts {
		o(p)
	}
	return p
}

// WithIO configures the prompter with custom IO streams.
func WithIO(io tui.IO) func(*Prompter) {
	return func(p *Prompter) {
		if io.In != nil {
			p.in = io.In
		}
		if io.Out != nil {
			p.out = io.Out
		}
		if io.ErrOut != nil {
			p.errOut = io.ErrOut
		}
	}
}

// Compile-time interface checks.
var _ tui.Prompter = (*Prompter)(nil)
var _ tui.Status = (*Prompter)(nil)

func init() {
	tui.RegisterBuilder(func(_ ...func(*tui.IO)) tui.Prompter {
		return New()
	})
	tui.RegisterStatusBuilder(func(_ ...func(*tui.IO)) tui.Status {
		return New()
	})
}
