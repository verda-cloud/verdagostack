package app

import (
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// OnInitialize returns a function suitable for cobra.OnInitialize that sets up
// Viper to read configuration from a file, environment variables, and default
// search paths. This is a standalone alternative to AddConfigFlag for callers
// that manage their own Cobra commands.
//
// Parameters:
//   - cfgFile: pointer to a config file path (may be nil or empty to use search paths)
//   - envPrefix: environment variable prefix (e.g., "MYAPP")
//   - searchPaths: directories to search for the config file
//   - configName: config file name without extension (e.g., "myapp")
func OnInitialize(cfgFile *string, envPrefix string, searchPaths []string, configName string) func() {
	return func() {
		if cfgFile != nil && *cfgFile != "" {
			viper.SetConfigFile(*cfgFile)
		} else {
			for _, path := range searchPaths {
				viper.AddConfigPath(path)
			}
			viper.SetConfigType("yaml")
			viper.SetConfigName(configName)
		}

		viper.AutomaticEnv()
		viper.SetEnvPrefix(envPrefix)
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				slog.Warn("Config file not found; falling back to defaults or environment variables", "err", err)
			} else {
				slog.Error("Failed to parse config file", "err", err)
				os.Exit(1)
			}
		}

		if path := viper.ConfigFileUsed(); path != "" {
			slog.Info("Using config file", "path", path)
		} else {
			slog.Info("No config file used")
		}
	}
}
