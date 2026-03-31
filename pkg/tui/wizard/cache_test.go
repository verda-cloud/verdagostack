package wizard

import "testing"

func TestCache_GetMiss(t *testing.T) {
	c := newStepCache()
	choices, ok := c.get("gpu", map[string]any{"loc": "FIN-01"})
	if ok {
		t.Error("expected cache miss")
	}
	if choices != nil {
		t.Error("expected nil choices on miss")
	}
}

func TestCache_SetAndGet(t *testing.T) {
	c := newStepCache()
	deps := map[string]any{"loc": "FIN-01"}
	want := []Choice{{Label: "H100", Value: "h100"}}

	c.set("gpu", deps, want)

	got, ok := c.get("gpu", deps)
	if !ok {
		t.Fatal("expected cache hit")
	}
	if len(got) != 1 || got[0].Value != "h100" {
		t.Error("unexpected cached value")
	}
}

func TestCache_DifferentDeps_Miss(t *testing.T) {
	c := newStepCache()
	c.set("gpu", map[string]any{"loc": "FIN-01"}, []Choice{{Value: "h100"}})

	_, ok := c.get("gpu", map[string]any{"loc": "FIN-03"})
	if ok {
		t.Error("expected cache miss for different deps")
	}
}

func TestCache_Invalidate(t *testing.T) {
	c := newStepCache()
	c.set("gpu", map[string]any{"loc": "FIN-01"}, []Choice{{Value: "h100"}})
	c.set("image", map[string]any{"gpu": "h100"}, []Choice{{Value: "ubuntu"}})

	c.invalidate("gpu")

	_, ok := c.get("gpu", map[string]any{"loc": "FIN-01"})
	if ok {
		t.Error("expected gpu cache to be invalidated")
	}
	_, ok = c.get("image", map[string]any{"gpu": "h100"})
	if !ok {
		t.Error("expected image cache to still exist")
	}
}
