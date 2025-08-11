package piquery

import (
	"testing"
)

func TestNewPIStore(t *testing.T) {
	store, err := NewPIStore()
	if err != nil {
		t.Fatalf("NewPIStore failed: %v", err)
	}

	if store.MaxPosition() <= 0 {
		t.Error("Expected max position greater than 0")
	}
}

func TestQueryValidPosition(t *testing.T) {
	store, err := NewPIStore()
	if err != nil {
		t.Fatalf("NewPIStore failed: %v", err)
	}

	// 测试已知的第1位
	result, err := store.Query(1)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if result.Position != 1 {
		t.Errorf("Expected position 1, got %d", result.Position)
	}

	if result.CurrentDigit != '3' {
		t.Errorf("Expected current digit '3', got %c", result.CurrentDigit)
	}

	// 测试中间位置
	if store.MaxPosition() > 100 {
		result, err := store.Query(100)
		if err != nil {
			t.Errorf("Query for position 100 failed: %v", err)
		}
	}
}

func TestQueryInvalidPosition(t *testing.T) {
	store, err := NewPIStore()
	if err != nil {
		t.Fatalf("NewPIStore failed: %v", err)
	}

	// 测试位置0
	_, err = store.Query(0)
	if err == nil {
		t.Error("Expected error for position 0, got nil")
	}

	// 测试超出最大位置的查询
	_, err = store.Query(store.MaxPosition() + 1)
	if err == nil {
		t.Error("Expected error for position beyond max, got nil")
	}
}

func TestQueryEdgeCases(t *testing.T) {
	store, err := NewPIStore()
	if err != nil {
		t.Fatalf("NewPIStore failed: %v", err)
	}

	maxPos := store.MaxPosition()
	
	// 测试第一位（前5位不足）
	result, err := store.Query(1)
	if err != nil {
		t.Errorf("Query for position 1 failed: %v", err)
	}
	if len(result.PreviousDigits) != 0 {
		t.Errorf("Expected 0 previous digits for position 1, got %d", len(result.PreviousDigits))
	}

	// 测试最后一位（后5位不足）
	result, err = store.Query(maxPos)
	if err != nil {
		t.Errorf("Query for position %d failed: %v", maxPos, err)
	}
	if len(result.NextDigits) != 0 {
		t.Errorf("Expected 0 next digits for last position, got %d", len(result.NextDigits))
	}
}
