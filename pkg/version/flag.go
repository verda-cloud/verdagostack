package version

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/pflag"
)

type versionValue int

const (
	versionNotSet versionValue = iota
	versionEnabled
	versionRaw
)

const versionFlagName = "version"

var versionFlag = versionNotSet

// Implements pflag.Value so --version accepts "true", "false", "raw".
func (v *versionValue) Set(s string) error {
	if s == "raw" {
		*v = versionRaw
		return nil
	}
	b, err := strconv.ParseBool(s)
	if err != nil {
		return fmt.Errorf("invalid value %q for --%s: must be true, false, or raw", s, versionFlagName)
	}
	if b {
		*v = versionEnabled
	} else {
		*v = versionNotSet
	}
	return nil
}

func (v *versionValue) String() string {
	switch *v {
	case versionRaw:
		return "raw"
	case versionEnabled:
		return "true"
	default:
		return "false"
	}
}

func (v *versionValue) Type() string { return "version" }

func init() {
	pflag.CommandLine.Var(&versionFlag, versionFlagName, `Print version information and quit.
Accepts "true", "false", or "raw" for full details.`)
	pflag.CommandLine.Lookup(versionFlagName).NoOptDefVal = "true"
}

// AddFlags adds the --version flag to the given FlagSet.
func AddFlags(fs *pflag.FlagSet) {
	if f := pflag.CommandLine.Lookup(versionFlagName); f != nil {
		fs.AddFlag(f)
	}
}

// PrintAndExitIfRequested prints version info and exits if --version was set.
func PrintAndExitIfRequested() {
	switch versionFlag {
	case versionRaw:
		fmt.Println(Get().Text())
		os.Exit(0)
	case versionEnabled:
		fmt.Println(Get().String())
		os.Exit(0)
	}
}
