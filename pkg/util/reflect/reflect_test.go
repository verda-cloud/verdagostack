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

package reflect

import (
	"testing"
)

type sampleStruct struct {
	Name  string
	Age   int
	Email string
}

func TestGetFieldsMap_All(t *testing.T) {
	s := sampleStruct{Name: "Alice", Age: 30, Email: "alice@example.com"}
	m := GetFieldsMap(s, nil)

	if m["Name"] != "Alice" {
		t.Fatalf("Name = %v, want Alice", m["Name"])
	}
	if m["Age"] != 30 {
		t.Fatalf("Age = %v, want 30", m["Age"])
	}
	if len(m) != 3 {
		t.Fatalf("len = %d, want 3", len(m))
	}
}

func TestGetFieldsMap_Selected(t *testing.T) {
	s := sampleStruct{Name: "Bob", Age: 25, Email: "bob@example.com"}
	m := GetFieldsMap(s, []string{"Name", "Age"})

	if len(m) != 2 {
		t.Fatalf("len = %d, want 2", len(m))
	}
	if _, ok := m["Email"]; ok {
		t.Fatal("Email should not be included")
	}
}

func TestGetFieldsMap_Pointer(t *testing.T) {
	s := &sampleStruct{Name: "Charlie", Age: 40}
	m := GetFieldsMap(s, nil)

	if m["Name"] != "Charlie" {
		t.Fatalf("Name = %v, want Charlie", m["Name"])
	}
}

func TestCopyFields(t *testing.T) {
	src := &sampleStruct{Name: "New", Age: 99, Email: "new@example.com"}
	dst := &sampleStruct{Name: "Old", Age: 1, Email: "old@example.com"}

	changed, err := CopyFields(src, dst, []string{"Name", "Age"})
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected changed=true")
	}
	if dst.Name != "New" || dst.Age != 99 {
		t.Fatalf("dst = %+v, want Name=New Age=99", dst)
	}
	if dst.Email != "old@example.com" {
		t.Fatal("Email should not have changed")
	}
}

func TestCopyFields_NoChange(t *testing.T) {
	src := &sampleStruct{Name: "Same", Age: 1}
	dst := &sampleStruct{Name: "Same", Age: 1}

	changed, err := CopyFields(src, dst, []string{"Name", "Age"})
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatal("expected changed=false when fields are equal")
	}
}

func TestCopyViaYAML(t *testing.T) {
	type A struct {
		Name string `yaml:"name"`
		Age  int    `yaml:"age"`
	}
	src := A{Name: "YAML", Age: 42}
	var dst A

	if err := CopyViaYAML(&dst, &src); err != nil {
		t.Fatal(err)
	}
	if dst.Name != "YAML" || dst.Age != 42 {
		t.Fatalf("dst = %+v, want Name=YAML Age=42", dst)
	}
}

func TestCopyViaYAML_Nil(t *testing.T) {
	if err := CopyViaYAML(nil, nil); err != nil {
		t.Fatal(err)
	}
}

func TestStructName(t *testing.T) {
	if name := StructName(sampleStruct{}); name != "sampleStruct" {
		t.Fatalf("StructName = %q, want sampleStruct", name)
	}
	if name := StructName(&sampleStruct{}); name != "sampleStruct" {
		t.Fatalf("StructName(ptr) = %q, want sampleStruct", name)
	}
}
