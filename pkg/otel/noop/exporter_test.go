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

package noop

import (
	"context"
	"testing"
)

func TestExporter_ExportSpans(t *testing.T) {
	e := NewExporter()
	if err := e.ExportSpans(context.Background(), nil); err != nil {
		t.Errorf("ExportSpans returned unexpected error: %v", err)
	}
}

func TestExporter_Shutdown(t *testing.T) {
	e := NewExporter()
	if err := e.Shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown returned unexpected error: %v", err)
	}
}
