package metrics

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

// newTestMetrics returns a Metrics instance backed by an isolated registry
// so tests don't pollute the global DefaultRegisterer.
func newTestMetrics(opts ...Option) *Metrics {
	reg := prometheus.NewRegistry()
	return New(append([]Option{WithRegistry(reg)}, opts...)...)
}

// --- Register / MustRegister ---

func TestRegister_Counter(t *testing.T) {
	m := newTestMetrics()
	if err := m.Register(MetricOpts{
		Name:   "req_total",
		Help:   "Total requests",
		Type:   CounterType,
		Labels: []string{"method"},
	}); err != nil {
		t.Fatal(err)
	}
}

func TestRegister_Duplicate(t *testing.T) {
	m := newTestMetrics()
	opts := MetricOpts{Name: "dup", Help: "h", Type: CounterType}
	if err := m.Register(opts); err != nil {
		t.Fatal(err)
	}
	if err := m.Register(opts); err == nil {
		t.Fatal("expected error on duplicate registration")
	}
}

func TestRegister_UnsupportedType(t *testing.T) {
	m := newTestMetrics()
	err := m.Register(MetricOpts{Name: "bad", Help: "h", Type: "unknown"})
	if err == nil {
		t.Fatal("expected error for unsupported metric type")
	}
}

func TestMustRegister_Panics(t *testing.T) {
	m := newTestMetrics()
	m.MustRegister(MetricOpts{Name: "ok", Help: "h", Type: GaugeType})

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on duplicate MustRegister")
		}
	}()
	m.MustRegister(MetricOpts{Name: "ok", Help: "h", Type: GaugeType})
}

func TestRegisterAll(t *testing.T) {
	m := newTestMetrics()
	err := m.RegisterAll([]MetricOpts{
		{Name: "a", Help: "h", Type: CounterType},
		{Name: "b", Help: "h", Type: GaugeType},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRegisterAll_StopsOnError(t *testing.T) {
	m := newTestMetrics()
	m.MustRegister(MetricOpts{Name: "a", Help: "h", Type: CounterType})
	err := m.RegisterAll([]MetricOpts{
		{Name: "a", Help: "h", Type: CounterType}, // duplicate
		{Name: "b", Help: "h", Type: GaugeType},
	})
	if err == nil {
		t.Fatal("expected error from RegisterAll")
	}
}

// --- Namespace / Subsystem ---

func TestNamespaceSubsystem(t *testing.T) {
	m := newTestMetrics(WithNamespace("ns"), WithSubsystem("sub"))
	m.MustRegister(MetricOpts{
		Name:   "my_counter",
		Help:   "h",
		Type:   CounterType,
		Labels: []string{"a"},
	})

	if err := m.UpdateMetric("my_counter", 1, []string{"x"}); err != nil {
		t.Fatal(err)
	}

	body := scrapeHandler(t, m.Handler())
	// The metric should have the fully qualified name ns_sub_my_counter
	if !strings.Contains(body, "ns_sub_my_counter") {
		t.Fatalf("expected ns_sub_my_counter in output, got:\n%s", body)
	}
}

// --- UpdateMetric ---

func TestUpdateMetric_Counter(t *testing.T) {
	m := newTestMetrics()
	m.MustRegister(MetricOpts{Name: "c", Help: "h", Type: CounterType, Labels: []string{"l"}})

	if err := m.UpdateMetric("c", 5, []string{"v"}); err != nil {
		t.Fatal(err)
	}

	body := scrapeHandler(t, m.Handler())
	if !strings.Contains(body, `c{l="v"} 5`) {
		t.Fatalf("counter not updated, got:\n%s", body)
	}
}

func TestUpdateMetric_Gauge(t *testing.T) {
	m := newTestMetrics()
	m.MustRegister(MetricOpts{Name: "g", Help: "h", Type: GaugeType, Labels: []string{"l"}})

	if err := m.UpdateMetric("g", 42, []string{"v"}); err != nil {
		t.Fatal(err)
	}

	body := scrapeHandler(t, m.Handler())
	if !strings.Contains(body, `g{l="v"} 42`) {
		t.Fatalf("gauge not updated, got:\n%s", body)
	}
}

func TestUpdateMetric_Histogram(t *testing.T) {
	m := newTestMetrics()
	m.MustRegister(MetricOpts{
		Name:    "h",
		Help:    "h",
		Type:    HistogramType,
		Labels:  []string{"l"},
		Buckets: []float64{0.1, 0.5, 1},
	})

	if err := m.UpdateMetric("h", 0.3, []string{"v"}); err != nil {
		t.Fatal(err)
	}

	body := scrapeHandler(t, m.Handler())
	if !strings.Contains(body, "h_count") {
		t.Fatalf("histogram not updated, got:\n%s", body)
	}
}

func TestUpdateMetric_Summary(t *testing.T) {
	m := newTestMetrics()
	m.MustRegister(MetricOpts{Name: "s", Help: "h", Type: SummaryType, Labels: []string{"l"}})

	if err := m.UpdateMetric("s", 1.5, []string{"v"}); err != nil {
		t.Fatal(err)
	}

	body := scrapeHandler(t, m.Handler())
	if !strings.Contains(body, "s_count") {
		t.Fatalf("summary not updated, got:\n%s", body)
	}
}

func TestUpdateMetric_NotFound(t *testing.T) {
	m := newTestMetrics()
	if err := m.UpdateMetric("missing", 1, nil); err == nil {
		t.Fatal("expected error for unknown metric")
	}
}

// --- Typed Getters ---

func TestGetCounter(t *testing.T) {
	m := newTestMetrics()
	m.MustRegister(MetricOpts{Name: "c", Help: "h", Type: CounterType, Labels: []string{"a"}})

	cv, err := m.GetCounter("c")
	if err != nil {
		t.Fatal(err)
	}
	cv.WithLabelValues("x").Inc()

	body := scrapeHandler(t, m.Handler())
	if !strings.Contains(body, `c{a="x"} 1`) {
		t.Fatalf("expected counter via typed getter, got:\n%s", body)
	}
}

func TestGetCounter_WrongType(t *testing.T) {
	m := newTestMetrics()
	m.MustRegister(MetricOpts{Name: "g", Help: "h", Type: GaugeType})
	_, err := m.GetCounter("g")
	if err == nil {
		t.Fatal("expected error when getting gauge as counter")
	}
}

func TestGetGauge(t *testing.T) {
	m := newTestMetrics()
	m.MustRegister(MetricOpts{Name: "g", Help: "h", Type: GaugeType, Labels: []string{"a"}})

	gv, err := m.GetGauge("g")
	if err != nil {
		t.Fatal(err)
	}
	gv.WithLabelValues("x").Set(99)

	body := scrapeHandler(t, m.Handler())
	if !strings.Contains(body, `g{a="x"} 99`) {
		t.Fatalf("expected gauge via typed getter, got:\n%s", body)
	}
}

func TestGetHistogram(t *testing.T) {
	m := newTestMetrics()
	m.MustRegister(MetricOpts{Name: "h", Help: "h", Type: HistogramType, Labels: []string{"a"}})

	hv, err := m.GetHistogram("h")
	if err != nil {
		t.Fatal(err)
	}
	hv.WithLabelValues("x").Observe(0.5)

	body := scrapeHandler(t, m.Handler())
	if !strings.Contains(body, "h_count") {
		t.Fatalf("expected histogram via typed getter, got:\n%s", body)
	}
}

func TestGetSummary(t *testing.T) {
	m := newTestMetrics()
	m.MustRegister(MetricOpts{Name: "s", Help: "h", Type: SummaryType, Labels: []string{"a"}})

	sv, err := m.GetSummary("s")
	if err != nil {
		t.Fatal(err)
	}
	sv.WithLabelValues("x").Observe(1.5)

	body := scrapeHandler(t, m.Handler())
	if !strings.Contains(body, "s_count") {
		t.Fatalf("expected summary via typed getter, got:\n%s", body)
	}
}

func TestGetCollector(t *testing.T) {
	m := newTestMetrics()
	m.MustRegister(MetricOpts{Name: "c", Help: "h", Type: CounterType})

	col, err := m.GetCollector("c")
	if err != nil {
		t.Fatal(err)
	}
	if col == nil {
		t.Fatal("expected non-nil collector")
	}
}

func TestGetCollector_NotFound(t *testing.T) {
	m := newTestMetrics()
	_, err := m.GetCollector("nope")
	if err == nil {
		t.Fatal("expected error for unknown metric")
	}
}

// --- Handler ---

func TestHandler_ServesMetrics(t *testing.T) {
	m := newTestMetrics()
	m.MustRegister(MetricOpts{Name: "x", Help: "h", Type: CounterType, Labels: []string{"k"}})
	_ = m.UpdateMetric("x", 7, []string{"v"})

	body := scrapeHandler(t, m.Handler())
	if !strings.Contains(body, `x{k="v"} 7`) {
		t.Fatalf("handler did not serve metrics, got:\n%s", body)
	}
}

func TestHandler_EmptyRegistry(t *testing.T) {
	m := newTestMetrics()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	m.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 from empty registry, got %d", rec.Code)
	}
}

// --- Custom Registry ---

func TestWithRegistry_Isolation(t *testing.T) {
	reg1 := prometheus.NewRegistry()
	reg2 := prometheus.NewRegistry()

	m1 := New(WithRegistry(reg1))
	m2 := New(WithRegistry(reg2))

	m1.MustRegister(MetricOpts{Name: "only_in_m1", Help: "h", Type: CounterType, Labels: []string{"a"}})
	m2.MustRegister(MetricOpts{Name: "only_in_m2", Help: "h", Type: GaugeType, Labels: []string{"a"}})

	_ = m1.UpdateMetric("only_in_m1", 1, []string{"x"})
	_ = m2.UpdateMetric("only_in_m2", 2, []string{"y"})

	body1 := scrapeHandler(t, m1.Handler())
	body2 := scrapeHandler(t, m2.Handler())

	if !strings.Contains(body1, "only_in_m1") {
		t.Fatal("m1 should contain only_in_m1")
	}
	if strings.Contains(body1, "only_in_m2") {
		t.Fatal("m1 should NOT contain only_in_m2")
	}

	if !strings.Contains(body2, "only_in_m2") {
		t.Fatal("m2 should contain only_in_m2")
	}
	if strings.Contains(body2, "only_in_m1") {
		t.Fatal("m2 should NOT contain only_in_m1")
	}
}

// --- Histogram with Default Buckets ---

func TestHistogram_DefaultBuckets(t *testing.T) {
	m := newTestMetrics()
	m.MustRegister(MetricOpts{Name: "hd", Help: "h", Type: HistogramType, Labels: []string{"a"}})

	if err := m.UpdateMetric("hd", 0.25, []string{"x"}); err != nil {
		t.Fatal(err)
	}

	body := scrapeHandler(t, m.Handler())
	if !strings.Contains(body, "hd_bucket") {
		t.Fatalf("expected default histogram buckets, got:\n%s", body)
	}
}

// --- helpers ---

func scrapeHandler(t *testing.T, h http.Handler) string {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("handler returned %d", rec.Code)
	}
	b, err := io.ReadAll(rec.Body)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
