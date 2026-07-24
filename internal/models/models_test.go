package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestEventJSONMarshaling(t *testing.T) {
	event := Event{
		EventType: "login",
		UserID:    "user1",
		IpAddress: "10.0.0.1",
		Status:    "failure",
		Timestamp: "2026-01-15T10:30:00Z",
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal event: %v", err)
	}

	var decoded Event
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal event: %v", err)
	}

	if decoded.EventType != event.EventType {
		t.Errorf("EventType: got %s, want %s", decoded.EventType, event.EventType)
	}
	if decoded.UserID != event.UserID {
		t.Errorf("UserID: got %s, want %s", decoded.UserID, event.UserID)
	}
	if decoded.IpAddress != event.IpAddress {
		t.Errorf("IpAddress: got %s, want %s", decoded.IpAddress, event.IpAddress)
	}
	if decoded.Status != event.Status {
		t.Errorf("Status: got %s, want %s", decoded.Status, event.Status)
	}
}

func TestEventJSONFromKafka(t *testing.T) {
	raw := `{
		"event_type": "login",
		"user_id": "user1",
		"ip_address": "10.0.0.1",
		"status": "failure",
		"timestamp": "2026-01-15T10:30:00Z"
	}`

	var event Event
	if err := json.Unmarshal([]byte(raw), &event); err != nil {
		t.Fatalf("failed to unmarshal event: %v", err)
	}

	if event.EventType != "login" {
		t.Errorf("EventType: got %s, want login", event.EventType)
	}
	if event.UserID != "user1" {
		t.Errorf("UserID: got %s, want user1", event.UserID)
	}
	if event.Status != "failure" {
		t.Errorf("Status: got %s, want failure", event.Status)
	}
}

func TestIncidentJSONMarshaling(t *testing.T) {
	ts := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
	incident := Incident{
		UserID:    "user1",
		IpAddress: "10.0.0.1",
		EventType: "login",
		Timestamp: ts,
		Details:   "Multiple failed login attempts detected",
	}

	data, err := json.Marshal(incident)
	if err != nil {
		t.Fatalf("failed to marshal incident: %v", err)
	}

	var decoded Incident
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal incident: %v", err)
	}

	if decoded.UserID != incident.UserID {
		t.Errorf("UserID: got %s, want %s", decoded.UserID, incident.UserID)
	}
	if decoded.Details != incident.Details {
		t.Errorf("Details: got %s, want %s", decoded.Details, incident.Details)
	}
}

func TestIncidentJSONFieldNames(t *testing.T) {
	ts := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
	incident := Incident{
		UserID:    "user1",
		IpAddress: "10.0.0.1",
		EventType: "login",
		Timestamp: ts,
		Details:   "test",
	}

	data, err := json.Marshal(incident)
	if err != nil {
		t.Fatalf("failed to marshal incident: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal to map: %v", err)
	}

	expectedKeys := []string{"user_id", "ip_address", "event_type", "timestamp", "details"}
	for _, key := range expectedKeys {
		if _, ok := raw[key]; !ok {
			t.Errorf("missing JSON key: %s", key)
		}
	}
}
