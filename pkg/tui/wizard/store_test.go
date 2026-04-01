package wizard

import "testing"

func TestStore_Collected(t *testing.T) {
	s := NewStore()
	s.SetCollected("region", "FIN-01")
	s.SetCollected("gpu", "h100")

	col := s.Collected()
	if col["region"] != "FIN-01" {
		t.Errorf("expected FIN-01, got %v", col["region"])
	}
	if col["gpu"] != "h100" {
		t.Errorf("expected h100, got %v", col["gpu"])
	}
}

func TestStore_GetSet(t *testing.T) {
	s := NewStore()
	s.Set("cost", 3.50)

	v, ok := s.Get("cost")
	if !ok {
		t.Fatal("expected cost to be set")
	}
	if v != 3.50 {
		t.Errorf("expected 3.50, got %v", v)
	}

	_, ok = s.Get("missing")
	if ok {
		t.Error("expected missing key to return false")
	}
}

func TestStore_CollectedReturnsSnapshot(t *testing.T) {
	s := NewStore()
	s.SetCollected("a", "1")

	snap := s.Collected()
	s.SetCollected("a", "2")

	if snap["a"] != "1" {
		t.Error("Collected() should return a snapshot, not a live reference")
	}
}

func TestStore_Clear(t *testing.T) {
	s := NewStore()
	s.SetCollected("a", "1")
	s.Set("cost", 5.0)

	s.ClearCollected("a")

	col := s.Collected()
	if _, ok := col["a"]; ok {
		t.Error("expected 'a' to be cleared")
	}

	v, ok := s.Get("cost")
	if !ok || v != 5.0 {
		t.Error("store data should be unaffected by ClearCollected")
	}
}
