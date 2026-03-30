package app

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const configFlagName = "config"

var cfgFile string

// AddConfigFlag registers the --config/-c flag and sets up Viper to read
// configuration from the specified file, default search paths, and environment
// variables. If watch is true, config changes are reloaded at runtime.
func AddConfigFlag(fs *pflag.FlagSet, name string, watch bool) {
	fs.AddFlag(pflag.Lookup(configFlagName))

	viper.AutomaticEnv()
	viper.SetEnvPrefix(strings.ReplaceAll(strings.ToUpper(name), "-", "_"))
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	cobra.OnInitialize(func() {
		if cfgFile != "" {
			viper.SetConfigFile(cfgFile)
		} else {
			viper.AddConfigPath(".")

			if home, err := os.UserHomeDir(); err == nil {
				if parts := strings.Split(name, "-"); len(parts) > 1 {
					viper.AddConfigPath(filepath.Join(home, "."+parts[0]))
					viper.AddConfigPath(filepath.Join("/etc", parts[0]))
				}
			}

			viper.SetConfigName(name)
		}

		if err := viper.ReadInConfig(); err != nil {
			slog.Debug("Failed to read configuration file", "file", cfgFile, "error", err)
		} else {
			slog.Debug("Loaded configuration file", "file", viper.ConfigFileUsed())
		}

		if watch {
			viper.WatchConfig()
			viper.OnConfigChange(func(e fsnotify.Event) {
				slog.Info("Config file changed", "name", e.Name)
			})
		}
	})
}

// PrintConfig logs all viper configuration keys and values.
func PrintConfig() {
	for _, key := range viper.AllKeys() {
		slog.Debug(fmt.Sprintf("CFG: %s=%v", key, viper.Get(key)))
	}
}

func init() {
	pflag.StringVarP(&cfgFile, configFlagName, "c", cfgFile, "Read configuration from specified `FILE`, "+
		"support JSON, TOML, YAML, HCL, or Java properties formats.")
}
