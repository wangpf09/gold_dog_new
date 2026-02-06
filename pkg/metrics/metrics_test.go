package metrics

import "testing"

func TestRollingWindow(t *testing.T) {
	w := NewRollingWindow[int](3)

	w.Push(1)
	if w.Size() != 1 {
		t.Errorf("Expected size 1, got %d", w.Size())
	}

	w.Push(2)
	w.Push(3)
	if !w.IsFull() {
		t.Error("Expected window to be full")
	}

	// Add one more, should wrap around
	w.Push(4)
	values := w.Values()
	expected := []int{2, 3, 4}
	if len(values) != len(expected) {
		t.Errorf("Expected %d values, got %d", len(expected), len(values))
	}

	for i, v := range expected {
		if values[i] != v {
			t.Errorf("Expected values[%d] = %d, got %d", i, v, values[i])
		}
	}
}
