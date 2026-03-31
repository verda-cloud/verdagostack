package app

import (
	"bytes"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// NamedFlagSets stores named flag sets in the order of calling FlagSet.
type NamedFlagSets struct {
	order    []string
	FlagSets map[string]*pflag.FlagSet
}

// FlagSet returns the FlagSet for a given name, creating it if necessary.
// Subsequent calls with the same name return the same FlagSet.
func (nfs *NamedFlagSets) FlagSet(name string) *pflag.FlagSet {
	if nfs.FlagSets == nil {
		nfs.FlagSets = map[string]*pflag.FlagSet{}
	}
	if _, ok := nfs.FlagSets[name]; !ok {
		nfs.FlagSets[name] = pflag.NewFlagSet(name, pflag.ExitOnError)
		nfs.order = append(nfs.order, name)
	}
	return nfs.FlagSets[name]
}

// SetUsageAndHelpFunc sets the usage and help functions for a cobra command
// using grouped flag sets for a cleaner display.
func SetUsageAndHelpFunc(cmd *cobra.Command, nfs NamedFlagSets, cols int) {
	if cols <= 0 {
		cols = 80
	}

	cmd.SetUsageFunc(func(cmd *cobra.Command) error {
		_, _ = fmt.Fprintf(cmd.OutOrStderr(), "Usage:\n  %s\n", cmd.UseLine())

		PrintSections(cmd.OutOrStderr(), nfs, cols)
		return nil
	})

	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n\nUsage:\n  %s\n", cmd.Long, cmd.UseLine())

		PrintSections(cmd.OutOrStdout(), nfs, cols)
	})
}

// PrintSections prints each named flag set as a titled section.
func PrintSections(w io.Writer, nfs NamedFlagSets, cols int) {
	for _, name := range nfs.order {
		fs := nfs.FlagSets[name]
		if !fs.HasFlags() {
			continue
		}

		var buf bytes.Buffer
		fs.SetOutput(&buf)
		fs.PrintDefaults()

		if buf.Len() > 0 {
			_, _ = fmt.Fprintf(w, "\n%s flags:\n\n%s", cases.Title(language.English).String(name), buf.String())
		}
	}
}

// PrintFlags logs all flags and their values via the provided print function.
func PrintFlags(fs *pflag.FlagSet, printFn func(format string, args ...any)) {
	fs.VisitAll(func(f *pflag.Flag) {
		printFn("FLAG: --%s=%q", f.Name, f.Value.String())
	})
}
