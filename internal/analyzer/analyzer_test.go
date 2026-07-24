package analyzer

import (
	"errors"
	"testing"
	"time"

	"github.com/AbhiramiRajeev/AIRO-Analyzer/config"
	"github.com/AbhiramiRajeev/AIRO-Analyzer/internal/models"
)

type mockRedis struct {
	addDataFn             func(string, float64) error
	remOldFailuresFn      func(string, float64) error
	getFailedCountFn      func(string) (int, error)
	addSuspiciousIpFn     func(string) error
	isSuspiciousIpFn      func(string) (bool, error)
	closeFn               func() error
}

func (m *mockRedis) AddData(username string, timestamp float64) error {
	if m.addDataFn != nil {
		return m.addDataFn(username, timestamp)
	}
	return nil
}

func (m *mockRedis) RemOldFailues(username string, timestamp float64) error {
	if m.remOldFailuresFn != nil {
		return m.remOldFailuresFn(username, timestamp)
	}
	return nil
}

func (m *mockRedis) GetFailedCount(username string) (int, error) {
	if m.getFailedCountFn != nil {
		return m.getFailedCountFn(username)
	}
	return 0, nil
}

func (m *mockRedis) AddSuspiciousIp(ip string) error {
	if m.addSuspiciousIpFn != nil {
		return m.addSuspiciousIpFn(ip)
	}
	return nil
}

func (m *mockRedis) IsSuspiciousIp(ip string) (bool, error) {
	if m.isSuspiciousIpFn != nil {
		return m.isSuspiciousIpFn(ip)
	}
	return false, nil
}

func (m *mockRedis) Close() error {
	if m.closeFn != nil {
		return m.closeFn()
	}
	return nil
}

type mockDB struct {
	addIncidentFn func(models.Incident) error
	closeFn       func() error
	createTableFn func() error
}

func (m *mockDB) AddIncident(incident models.Incident) error {
	if m.addIncidentFn != nil {
		return m.addIncidentFn(incident)
	}
	return nil
}

func (m *mockDB) Close() error {
	if m.closeFn != nil {
		return m.closeFn()
	}
	return nil
}

func (m *mockDB) CreateTable() error {
	if m.createTableFn != nil {
		return m.createTableFn()
	}
	return nil
}

func TestAnalyze_FailureBelowThreshold(t *testing.T) {
	redis := &mockRedis{
		getFailedCountFn: func(username string) (int, error) {
			return 2, nil
		},
	}
	db := &mockDB{}
	cfg := &config.Config{FailureThreshold: 5}

	svc := NewAnalyzerService(cfg, redis, db)

	event := models.Event{
		UserID:    "user1",
		IpAddress: "10.0.0.1",
		EventType: "login",
		Status:    "failure",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	err := svc.Analyze(event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnalyze_FailureExceedsThreshold(t *testing.T) {
	incidentSaved := false
	redis := &mockRedis{
		getFailedCountFn: func(username string) (int, error) {
			return 5, nil
		},
	}
	db := &mockDB{
		addIncidentFn: func(incident models.Incident) error {
			incidentSaved = true
			if incident.UserID != "user1" {
				t.Errorf("expected user1, got %s", incident.UserID)
			}
			if incident.Details != "Multiple failed login attempts detected" {
				t.Errorf("unexpected details: %s", incident.Details)
			}
			return nil
		},
	}
	cfg := &config.Config{FailureThreshold: 5}

	svc := NewAnalyzerService(cfg, redis, db)

	event := models.Event{
		UserID:    "user1",
		IpAddress: "10.0.0.1",
		EventType: "login",
		Status:    "failure",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	err := svc.Analyze(event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !incidentSaved {
		t.Error("expected incident to be saved")
	}
}

func TestAnalyze_SuccessfulLogin_SuspiciousIp(t *testing.T) {
	incidentSaved := false
	redis := &mockRedis{
		isSuspiciousIpFn: func(ip string) (bool, error) {
			if ip == "10.0.0.1" {
				return true, nil
			}
			return false, nil
		},
	}
	db := &mockDB{
		addIncidentFn: func(incident models.Incident) error {
			incidentSaved = true
			if incident.Details != "Suspicious IP address detected" {
				t.Errorf("unexpected details: %s", incident.Details)
			}
			return nil
		},
	}
	cfg := &config.Config{FailureThreshold: 5}

	svc := NewAnalyzerService(cfg, redis, db)

	event := models.Event{
		UserID:    "user1",
		IpAddress: "10.0.0.1",
		EventType: "login",
		Status:    "success",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	err := svc.Analyze(event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !incidentSaved {
		t.Error("expected incident to be saved for suspicious IP")
	}
}

func TestAnalyze_SuccessfulLogin_CleanIp(t *testing.T) {
	incidentSaved := false
	redis := &mockRedis{
		isSuspiciousIpFn: func(ip string) (bool, error) {
			return false, nil
		},
	}
	db := &mockDB{
		addIncidentFn: func(incident models.Incident) error {
			incidentSaved = true
			return nil
		},
	}
	cfg := &config.Config{FailureThreshold: 5}

	svc := NewAnalyzerService(cfg, redis, db)

	event := models.Event{
		UserID:    "user1",
		IpAddress: "10.0.0.1",
		EventType: "login",
		Status:    "success",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	err := svc.Analyze(event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if incidentSaved {
		t.Error("no incident should be saved for clean IP with successful login")
	}
}

func TestAnalyze_RedisError_AddData(t *testing.T) {
	redis := &mockRedis{
		addDataFn: func(username string, timestamp float64) error {
			return errors.New("redis connection refused")
		},
	}
	db := &mockDB{}
	cfg := &config.Config{FailureThreshold: 5}

	svc := NewAnalyzerService(cfg, redis, db)

	event := models.Event{
		UserID:    "user1",
		IpAddress: "10.0.0.1",
		EventType: "login",
		Status:    "failure",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	err := svc.Analyze(event)
	if err == nil {
		t.Fatal("expected error from redis")
	}
}

func TestAnalyze_RedisError_IsSuspiciousIp(t *testing.T) {
	redis := &mockRedis{
		isSuspiciousIpFn: func(ip string) (bool, error) {
			return false, errors.New("redis connection refused")
		},
	}
	db := &mockDB{}
	cfg := &config.Config{FailureThreshold: 5}

	svc := NewAnalyzerService(cfg, redis, db)

	event := models.Event{
		UserID:    "user1",
		IpAddress: "10.0.0.1",
		EventType: "login",
		Status:    "success",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	err := svc.Analyze(event)
	if err == nil {
		t.Fatal("expected error from redis")
	}
}

func TestAnalyze_DBError_OnIncident(t *testing.T) {
	redis := &mockRedis{
		isSuspiciousIpFn: func(ip string) (bool, error) {
			return true, nil
		},
	}
	db := &mockDB{
		addIncidentFn: func(incident models.Incident) error {
			return errors.New("db connection refused")
		},
	}
	cfg := &config.Config{FailureThreshold: 5}

	svc := NewAnalyzerService(cfg, redis, db)

	event := models.Event{
		UserID:    "user1",
		IpAddress: "10.0.0.1",
		EventType: "login",
		Status:    "success",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	err := svc.Analyze(event)
	if err == nil {
		t.Fatal("expected error from db")
	}
}
