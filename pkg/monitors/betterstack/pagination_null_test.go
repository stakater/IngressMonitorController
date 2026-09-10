package betterstack

import (
	"encoding/json"
	"testing"
)

// Better Stack returns pagination.next as JSON null on the last page, not "".
// Decoding null into a string must yield "" so GetAll's loop terminates; if
// this ever became a *string or an object, the loop would spin.
func TestPaginationNullDecodesToEmpty(t *testing.T) {
	var r monitorListResponse
	if err := json.Unmarshal([]byte(`{"data":[],"pagination":{"next":null}}`), &r); err != nil {
		t.Fatalf("null next must decode: %v", err)
	}
	if r.Pagination.Next != "" {
		t.Errorf("expected empty next, got %q", r.Pagination.Next)
	}
}
