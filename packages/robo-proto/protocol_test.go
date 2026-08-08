package roboproto

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTimestampSerializesAsRFC3339String(t *testing.T) {
	type payload struct {
		Timestamp time.Time `json:"timestamp"`
	}

	value := payload{
		Timestamp: time.Date(2026, 8, 8, 12, 0, 5, 123456789, time.UTC),
	}

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	const want = `{"timestamp":"2026-08-08T12:00:05.123456789Z"}`
	if string(data) != want {
		t.Fatalf("unexpected JSON\nwant: %s\n got: %s", want, string(data))
	}
}

func TestTimestampPreservesOffsetWhenNotUTC(t *testing.T) {
	type payload struct {
		Timestamp time.Time `json:"timestamp"`
	}

	zone := time.FixedZone("UTC+1", 60*60)
	value := payload{
		Timestamp: time.Date(2026, 8, 8, 13, 0, 5, 0, zone),
	}

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	const want = `{"timestamp":"2026-08-08T13:00:05+01:00"}`
	if string(data) != want {
		t.Fatalf("unexpected JSON\nwant: %s\n got: %s", want, string(data))
	}
}
