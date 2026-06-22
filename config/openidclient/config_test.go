package openidclient

import (
	"reflect"
	"testing"
)

// TestClientConnectionDetails locks the published keys to the simplified aliases
// only: it must not re-emit the raw attribute.<name> keys (Upjet adds those, and
// duplicating them conflicts), and empty values must be omitted.
func TestClientConnectionDetails(t *testing.T) {
	got, err := clientConnectionDetails(map[string]any{
		"client_secret":           "s3cret",
		"client_id":               "my-client",
		"service_account_user_id": "sa-123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[string][]byte{
		"clientSecret":         []byte("s3cret"),
		"clientID":             []byte("my-client"),
		"serviceAccountUserId": []byte("sa-123"),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("clientConnectionDetails() = %v, want %v", got, want)
	}

	// Absent or empty values are omitted, not published as empty keys.
	if got, _ := clientConnectionDetails(map[string]any{"client_secret": ""}); len(got) != 0 {
		t.Errorf("expected no keys for empty/absent values, got %v", got)
	}
}
