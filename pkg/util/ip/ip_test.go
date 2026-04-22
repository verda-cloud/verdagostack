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

package ip

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetLocalIP(t *testing.T) {
	ip := GetLocalIP()
	if ip == "" {
		t.Fatal("GetLocalIP returned empty string")
	}
}

func TestRemoteIP_Fallback(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	got := RemoteIP(req)
	if got != "192.168.1.1" {
		t.Fatalf("RemoteIP = %q, want 192.168.1.1", got)
	}
}

func TestRemoteIP_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(XForwardedFor, "10.0.0.1")

	got := RemoteIP(req)
	if got != "10.0.0.1" {
		t.Fatalf("RemoteIP = %q, want 10.0.0.1", got)
	}
}

func TestRemoteIP_XRealIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(XRealIP, "10.0.0.2")

	got := RemoteIP(req)
	if got != "10.0.0.2" {
		t.Fatalf("RemoteIP = %q, want 10.0.0.2", got)
	}
}

func TestRemoteIP_XClientIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(XClientIP, "10.0.0.3")

	got := RemoteIP(req)
	if got != "10.0.0.3" {
		t.Fatalf("RemoteIP = %q, want 10.0.0.3", got)
	}
}

func TestRemoteIP_IPv6Loopback(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "[::1]:12345"

	got := RemoteIP(req)
	if got != "127.0.0.1" {
		t.Fatalf("RemoteIP(::1) = %q, want 127.0.0.1", got)
	}
}
