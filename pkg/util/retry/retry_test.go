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

package retry

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestRetry_Success(t *testing.T) {
	attempts := 0
	err := Retry(func() (bool, error) {
		attempts++
		return attempts >= 3, nil
	}, BackoffConfig{Steps: 5, Duration: time.Millisecond, Factor: 1, Jitter: 0})

	if err != nil {
		t.Fatalf("Retry failed: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("attempts = %d, want 3", attempts)
	}
}

func TestRetry_Exhausted(t *testing.T) {
	err := Retry(func() (bool, error) {
		return false, nil
	}, BackoffConfig{Steps: 3, Duration: time.Millisecond, Factor: 1, Jitter: 0})

	if err == nil {
		t.Fatal("expected error when retries exhausted")
	}
}

func TestRetry_ErrorAborts(t *testing.T) {
	sentinel := errors.New("fatal")
	err := Retry(func() (bool, error) {
		return false, sentinel
	}, BackoffConfig{Steps: 5, Duration: time.Millisecond, Factor: 1, Jitter: 0})

	if !errors.Is(err, sentinel) {
		t.Fatalf("err = %v, want sentinel", err)
	}
}

func TestPollImmediate_ImmediateSuccess(t *testing.T) {
	err := PollImmediate(context.Background(), time.Second, time.Second, func() (bool, error) {
		return true, nil
	})
	if err != nil {
		t.Fatalf("PollImmediate failed: %v", err)
	}
}

func TestPoll_Timeout(t *testing.T) {
	ctx := context.Background()
	err := Poll(ctx, 10*time.Millisecond, 50*time.Millisecond, func() (bool, error) {
		return false, nil
	})
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestRunImmediatelyThenPeriod(t *testing.T) {
	var count atomic.Int32
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := RunImmediatelyThenPeriod(ctx, func(ctx context.Context) error {
		count.Add(1)
		return nil
	}, 10*time.Millisecond)

	if err != nil {
		t.Fatalf("RunImmediatelyThenPeriod failed: %v", err)
	}
	if count.Load() != 1 {
		t.Fatalf("expected 1 immediate call, got %d", count.Load())
	}

	time.Sleep(50 * time.Millisecond)
	cancel()

	if count.Load() < 2 {
		t.Fatalf("expected periodic calls, got %d total", count.Load())
	}
}

func TestRunImmediatelyThenPeriod_InitError(t *testing.T) {
	err := RunImmediatelyThenPeriod(context.Background(), func(ctx context.Context) error {
		return errors.New("init fail")
	}, time.Second)

	if err == nil {
		t.Fatal("expected error from initial execution")
	}
}

// --- Real-world usage examples as tests ---

func TestRetry_HTTPCall(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `{"status":"ok"}`)
	}))
	defer srv.Close()

	var body string
	err := Retry(func() (bool, error) {
		resp, err := http.Get(srv.URL)
		if err != nil {
			return false, nil // transient network error, retry
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode >= 500 {
			return false, nil // server error, retry
		}
		if resp.StatusCode == http.StatusUnauthorized {
			return false, fmt.Errorf("unauthorized, no point retrying")
		}

		b, _ := io.ReadAll(resp.Body)
		body = string(b)
		return true, nil
	}, BackoffConfig{Steps: 5, Duration: time.Millisecond, Factor: 1, Jitter: 0})

	if err != nil {
		t.Fatalf("HTTP retry failed: %v", err)
	}
	if body != `{"status":"ok"}` {
		t.Fatalf("unexpected body: %q", body)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestRetry_HTTPCall_PermanentError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	err := Retry(func() (bool, error) {
		resp, err := http.Get(srv.URL)
		if err != nil {
			return false, nil
		}
		_ = resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			return false, fmt.Errorf("401 unauthorized: abort retry")
		}
		return true, nil
	}, BackoffConfig{Steps: 5, Duration: time.Millisecond, Factor: 1, Jitter: 0})

	if err == nil {
		t.Fatal("expected error for 401")
	}
}

func TestPollImmediate_WaitForHealthy(t *testing.T) {
	var healthy atomic.Bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !healthy.Load() {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// Simulate the service becoming healthy after 30ms.
	go func() {
		time.Sleep(30 * time.Millisecond)
		healthy.Store(true)
	}()

	err := PollImmediate(context.Background(), 10*time.Millisecond, 200*time.Millisecond, func() (bool, error) {
		resp, err := http.Get(srv.URL + "/healthz")
		if err != nil {
			return false, nil
		}
		_ = resp.Body.Close()
		return resp.StatusCode == http.StatusOK, nil
	})

	if err != nil {
		t.Fatalf("expected service to become healthy: %v", err)
	}
}
