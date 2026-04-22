// Copyright 2026 Verda Cloud Oy
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
