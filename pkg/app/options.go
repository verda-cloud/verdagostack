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

import "github.com/spf13/pflag"

// OptionsValidator provides methods to complete and validate options.
type OptionsValidator interface {
	Complete() error
	Validate() error
}

// NamedFlagSetOptions provides access to grouped flag sets and validation.
// Options structs implement this to organize their flags into named sections.
type NamedFlagSetOptions interface {
	Flags() NamedFlagSets
	OptionsValidator
}

// FlagSetOptions defines options that add themselves to a single flat FlagSet.
type FlagSetOptions interface {
	AddFlags(fs *pflag.FlagSet)
	OptionsValidator
}
