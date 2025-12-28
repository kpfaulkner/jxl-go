package util

import (
	"testing"
)

func TestNewDeque(t *testing.T) {
	dq := NewDeque[int]()
	if dq == nil {
		t.Error("NewDeque should not return nil")
	}
	if !dq.IsEmpty() {
		t.Error("New deque should be empty")
	}
}

func TestDequeAddFirst(t *testing.T) {
	dq := NewDeque[int]()
	dq.AddFirst(1)
	dq.AddFirst(2)
	dq.AddFirst(3)

	// Should be [3, 2, 1]
	if dq.IsEmpty() {
		t.Error("Deque should not be empty after AddFirst")
	}

	val := dq.RemoveFirst()
	if val == nil || *val != 3 {
		t.Errorf("RemoveFirst should return 3, got %v", val)
	}

	val = dq.RemoveFirst()
	if val == nil || *val != 2 {
		t.Errorf("RemoveFirst should return 2, got %v", val)
	}

	val = dq.RemoveFirst()
	if val == nil || *val != 1 {
		t.Errorf("RemoveFirst should return 1, got %v", val)
	}

	if !dq.IsEmpty() {
		t.Error("Deque should be empty after removing all elements")
	}
}

func TestDequeAddLast(t *testing.T) {
	dq := NewDeque[int]()
	dq.AddLast(1)
	dq.AddLast(2)
	dq.AddLast(3)

	// Should be [1, 2, 3]
	val := dq.RemoveFirst()
	if val == nil || *val != 1 {
		t.Errorf("RemoveFirst should return 1, got %v", val)
	}

	val = dq.RemoveFirst()
	if val == nil || *val != 2 {
		t.Errorf("RemoveFirst should return 2, got %v", val)
	}

	val = dq.RemoveFirst()
	if val == nil || *val != 3 {
		t.Errorf("RemoveFirst should return 3, got %v", val)
	}
}

func TestDequeRemoveLast(t *testing.T) {
	dq := NewDeque[int]()
	dq.AddLast(1)
	dq.AddLast(2)
	dq.AddLast(3)

	// Should be [1, 2, 3], remove from end
	val := dq.RemoveLast()
	if val == nil || *val != 3 {
		t.Errorf("RemoveLast should return 3, got %v", val)
	}

	val = dq.RemoveLast()
	if val == nil || *val != 2 {
		t.Errorf("RemoveLast should return 2, got %v", val)
	}

	val = dq.RemoveLast()
	if val == nil || *val != 1 {
		t.Errorf("RemoveLast should return 1, got %v", val)
	}
}

func TestDequeRemoveFirstEmpty(t *testing.T) {
	dq := NewDeque[int]()
	val := dq.RemoveFirst()
	if val != nil {
		t.Error("RemoveFirst on empty deque should return nil")
	}
}

func TestDequeRemoveLastEmpty(t *testing.T) {
	dq := NewDeque[int]()
	val := dq.RemoveLast()
	if val != nil {
		t.Error("RemoveLast on empty deque should return nil")
	}
}

func TestDequeIsEmpty(t *testing.T) {
	dq := NewDeque[string]()
	if !dq.IsEmpty() {
		t.Error("New deque should be empty")
	}

	dq.AddFirst("test")
	if dq.IsEmpty() {
		t.Error("Deque with elements should not be empty")
	}

	dq.RemoveFirst()
	if !dq.IsEmpty() {
		t.Error("Deque after removing all elements should be empty")
	}
}

func TestDequeMixedOperations(t *testing.T) {
	dq := NewDeque[int]()

	// Add from both ends
	dq.AddFirst(2)  // [2]
	dq.AddFirst(1)  // [1, 2]
	dq.AddLast(3)   // [1, 2, 3]
	dq.AddLast(4)   // [1, 2, 3, 4]
	dq.AddFirst(0)  // [0, 1, 2, 3, 4]

	// Remove and verify
	val := dq.RemoveFirst()
	if *val != 0 {
		t.Errorf("Expected 0, got %d", *val)
	}

	val = dq.RemoveLast()
	if *val != 4 {
		t.Errorf("Expected 4, got %d", *val)
	}

	val = dq.RemoveFirst()
	if *val != 1 {
		t.Errorf("Expected 1, got %d", *val)
	}

	val = dq.RemoveLast()
	if *val != 3 {
		t.Errorf("Expected 3, got %d", *val)
	}

	val = dq.RemoveFirst()
	if *val != 2 {
		t.Errorf("Expected 2, got %d", *val)
	}

	if !dq.IsEmpty() {
		t.Error("Deque should be empty")
	}
}

func TestDequeWithStrings(t *testing.T) {
	dq := NewDeque[string]()
	dq.AddLast("hello")
	dq.AddLast("world")

	val := dq.RemoveFirst()
	if *val != "hello" {
		t.Errorf("Expected 'hello', got '%s'", *val)
	}

	val = dq.RemoveFirst()
	if *val != "world" {
		t.Errorf("Expected 'world', got '%s'", *val)
	}
}

func TestDequeWithStructs(t *testing.T) {
	type Item struct {
		ID   int
		Name string
	}

	dq := NewDeque[Item]()
	dq.AddLast(Item{ID: 1, Name: "first"})
	dq.AddLast(Item{ID: 2, Name: "second"})

	val := dq.RemoveFirst()
	if val.ID != 1 || val.Name != "first" {
		t.Errorf("Expected {1, first}, got %v", val)
	}
}

// Benchmarks
func BenchmarkDequeAddFirst(b *testing.B) {
	dq := NewDeque[int]()
	for i := 0; i < b.N; i++ {
		dq.AddFirst(i)
	}
}

func BenchmarkDequeAddLast(b *testing.B) {
	dq := NewDeque[int]()
	for i := 0; i < b.N; i++ {
		dq.AddLast(i)
	}
}

func BenchmarkDequeAddRemove(b *testing.B) {
	dq := NewDeque[int]()
	for i := 0; i < b.N; i++ {
		dq.AddLast(i)
		dq.RemoveFirst()
	}
}
