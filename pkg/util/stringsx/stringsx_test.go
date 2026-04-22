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

package stringsx

import (
	"testing"
)

func TestDiff(t *testing.T) {
	got := Diff([]string{"a", "b", "c"}, []string{"b", "d"})
	if len(got) != 2 || got[0] != "a" || got[1] != "c" {
		t.Fatalf("Diff = %v, want [a c]", got)
	}
}

func TestDiff_Empty(t *testing.T) {
	got := Diff(nil, []string{"a"})
	if len(got) != 0 {
		t.Fatalf("Diff(nil, ...) = %v, want empty", got)
	}
}

func TestInclude(t *testing.T) {
	got := Include([]string{"a", "b", "c"}, []string{"b", "c", "d"})
	if len(got) != 2 || got[0] != "b" || got[1] != "c" {
		t.Fatalf("Include = %v, want [b c]", got)
	}
}

func TestUnique(t *testing.T) {
	got := Unique([]string{"a", "b", "a", "c", "b"})
	if len(got) != 3 {
		t.Fatalf("Unique = %v, want 3 elements", got)
	}
	seen := map[string]bool{}
	for _, s := range got {
		if seen[s] {
			t.Fatalf("Unique returned duplicate: %s", s)
		}
		seen[s] = true
	}
}

func TestContains(t *testing.T) {
	if !Contains([]string{"a", "b"}, "b") {
		t.Fatal("Contains should find b")
	}
	if Contains([]string{"a", "b"}, "c") {
		t.Fatal("Contains should not find c")
	}
}

func TestContainsEqualFold(t *testing.T) {
	if !ContainsEqualFold([]string{"Hello", "World"}, "hello") {
		t.Fatal("ContainsEqualFold should match case-insensitively")
	}
}

func TestIndex(t *testing.T) {
	if i := Index([]string{"a", "b", "c"}, "b"); i != 1 {
		t.Fatalf("Index = %d, want 1", i)
	}
	if i := Index([]string{"a"}, "z"); i != -1 {
		t.Fatalf("Index = %d, want -1", i)
	}
}

func TestFilter(t *testing.T) {
	got := Filter([]string{"a", "b", "a", "c"}, "a")
	if len(got) != 2 || got[0] != "b" || got[1] != "c" {
		t.Fatalf("Filter = %v, want [b c]", got)
	}
}

func TestAdd(t *testing.T) {
	got := Add([]string{"a", "b"}, "c")
	if len(got) != 3 {
		t.Fatalf("Add new = %v, want 3 elements", got)
	}
	got2 := Add(got, "b")
	if len(got2) != 3 {
		t.Fatalf("Add existing = %v, want 3 elements", got2)
	}
}

func TestReverse(t *testing.T) {
	if got := Reverse("hello"); got != "olleh" {
		t.Fatalf("Reverse = %q, want olleh", got)
	}
	if got := Reverse("日本語"); got != "語本日" {
		t.Fatalf("Reverse(日本語) = %q, want 語本日", got)
	}
}

func TestDecodeBase64(t *testing.T) {
	encoded := "SGVsbG8=" // "Hello"
	got, err := DecodeBase64(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "Hello" {
		t.Fatalf("DecodeBase64 = %q, want Hello", got)
	}
}

func TestDecodeBase64_Invalid(t *testing.T) {
	_, err := DecodeBase64("not-valid-base64!!!")
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}
