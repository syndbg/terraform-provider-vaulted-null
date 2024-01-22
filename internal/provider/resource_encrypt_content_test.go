package provider

import (
	"testing"
)

func TestGenerateHashedTimestamp(t *testing.T) {
	timestamp := int64(1705911795)
	expectedHash := "344c10d9f5f38cc958ed5e1422d5cdb56624402dca6acd809b836a5c4c6a46ce"

	actualHash := generateHashedTimestamp(timestamp)

	if actualHash != expectedHash {
		t.Errorf("generateHashedTimestamp(%v) = %v; want %v", timestamp, actualHash, expectedHash)
	}
}
