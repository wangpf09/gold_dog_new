package metrics

// RollingWindow provides a fixed-size FIFO buffer for any type
type RollingWindow[T any] struct {
	data  []T
	size  int
	head  int
	count int
}

// NewRollingWindow creates a new rolling window with the specified size
func NewRollingWindow[T any](size int) *RollingWindow[T] {
	if size <= 0 {
		size = 1
	}
	return &RollingWindow[T]{
		data:  make([]T, size),
		size:  size,
		head:  0,
		count: 0,
	}
}

// Push adds a new value to the window (FIFO)
func (w *RollingWindow[T]) Push(value T) {
	w.data[w.head] = value
	w.head = (w.head + 1) % w.size
	if w.count < w.size {
		w.count++
	}
}

// Values returns all values in the window (oldest to newest)
func (w *RollingWindow[T]) Values() []T {
	if w.count == 0 {
		return []T{}
	}

	result := make([]T, w.count)
	if w.count < w.size {
		// Not full yet, data is at beginning
		copy(result, w.data[:w.count])
	} else {
		// Full, need to unwrap circular buffer
		tail := w.head
		copy(result, w.data[tail:])
		copy(result[w.size-tail:], w.data[:tail])
	}
	return result
}

// Latest returns the most recently added value
func (w *RollingWindow[T]) Latest() (T, bool) {
	var zero T
	if w.count == 0 {
		return zero, false
	}

	idx := (w.head - 1 + w.size) % w.size
	return w.data[idx], true
}

// Size returns the current number of elements in the window
func (w *RollingWindow[T]) Size() int {
	return w.count
}

// Capacity returns the maximum capacity of the window
func (w *RollingWindow[T]) Capacity() int {
	return w.size
}

// IsFull returns true if the window is at capacity
func (w *RollingWindow[T]) IsFull() bool {
	return w.count == w.size
}

// Clear removes all values from the window
func (w *RollingWindow[T]) Clear() {
	w.head = 0
	w.count = 0
}
