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

package options

import (
	"testing"
)

func TestJoin(t *testing.T) {
	tests := []struct {
		name     string
		prefixes []string
		want     string
	}{
		{"single", []string{"server"}, "server."},
		{"multiple", []string{"server", "http"}, "server.http."},
		{"empty", []string{}, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Join(tc.prefixes...)
			if got != tc.want {
				t.Errorf("Join(%v) = %q, want %q", tc.prefixes, got, tc.want)
			}
		})
	}
}

func TestValidateAddress(t *testing.T) {
	tests := []struct {
		addr    string
		wantErr bool
	}{
		{":8080", false},
		{"0.0.0.0:443", false},
		{"127.0.0.1:0", false},
		{"192.168.1.1:65535", false},
		{"bad-addr", true},
		{":99999", true},
		{"notanip:8080", true},
	}
	for _, tc := range tests {
		t.Run(tc.addr, func(t *testing.T) {
			err := ValidateAddress(tc.addr)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateAddress(%q) error = %v, wantErr %v", tc.addr, err, tc.wantErr)
			}
		})
	}
}

func TestGetLocalIP(t *testing.T) {
	ip := GetLocalIP()
	if ip == "" {
		t.Fatal("GetLocalIP returned empty string")
	}
}

func TestInsecureServingOptions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		opts    *InsecureServingOptions
		wantErr bool
	}{
		{"nil", nil, false},
		{"valid defaults", NewInsecureServingOptions(), false},
		{"bad address", &InsecureServingOptions{Addr: "bad"}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			errs := tc.opts.Validate()
			hasErr := len(errs) > 0
			if hasErr != tc.wantErr {
				t.Errorf("Validate() errors = %v, wantErr %v", errs, tc.wantErr)
			}
		})
	}
}

func TestHealthOptions_Defaults(t *testing.T) {
	opts := NewHealthOptions()
	if opts.HealthCheckPath != "/healthz" {
		t.Errorf("expected /healthz, got %q", opts.HealthCheckPath)
	}
	if opts.HealthCheckAddress != "0.0.0.0:20250" {
		t.Errorf("expected 0.0.0.0:20250, got %q", opts.HealthCheckAddress)
	}
	if opts.HTTPProfile {
		t.Error("HTTPProfile should default to false")
	}
}

func TestByteConstants(t *testing.T) {
	if KiB != 1024 {
		t.Errorf("KiB = %d, want 1024", KiB)
	}
	if MiB != 1024*1024 {
		t.Errorf("MiB = %d, want %d", MiB, 1024*1024)
	}
	if GiB != 1024*1024*1024 {
		t.Errorf("GiB = %d, want %d", GiB, 1024*1024*1024)
	}
}
