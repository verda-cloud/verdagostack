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

package pagination

import "testing"

func TestGetPageOffset(t *testing.T) {
	tests := []struct {
		page, size, want int64
	}{
		{1, 10, 0},
		{2, 10, 10},
		{3, 20, 40},
		{1, 100, 0},
		{5, 25, 100},
	}
	for _, tt := range tests {
		got := GetPageOffset(tt.page, tt.size)
		if got != tt.want {
			t.Errorf("GetPageOffset(%d, %d) = %d, want %d", tt.page, tt.size, got, tt.want)
		}
	}
}
