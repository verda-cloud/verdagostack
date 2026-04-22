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

package db

import (
	"testing"

	"gorm.io/gorm"
)

func TestTracePlugin_ImplementsPlugin(t *testing.T) {
	var _ gorm.Plugin = TracePlugin{}
}

func TestTracePlugin_Name(t *testing.T) {
	p := TracePlugin{}
	if name := p.Name(); name != "verdagostack:trace" {
		t.Errorf("expected name verdagostack:trace, got %s", name)
	}
}
