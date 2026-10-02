package pagination

import "testing"

func TestNewResultUsesEmptyArrayForNilItems(t *testing.T) {
	result := NewResult[int](nil, Query{Page: 1, PerPage: 10}, 0)
	if result.Items == nil {
		t.Fatal("NewResult() Items is nil; want an empty slice")
	}
	if len(result.Items) != 0 {
		t.Fatalf("NewResult() Items length = %d, want 0", len(result.Items))
	}
}
