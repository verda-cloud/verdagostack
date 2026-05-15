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

package bubbletea

// Common key names used in bubbletea Update handlers (still used by
// editor/pager/progress/spinner; prompt models now use KeyBinding).
const (
	keyCtrlC = "ctrl+c"
	keyEsc   = "esc"
)

// Common hint-label strings shared across prompt models. Centralizing
// them keeps the goconst linter quiet and ensures consistent wording.
const (
	hintEscBack    = "esc back"
	hintCtrlCExit  = "ctrl+c exit"
	hintEnterEntry = "enter submit"
)
