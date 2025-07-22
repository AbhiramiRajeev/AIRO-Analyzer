package analyzer

import (
	"time"

	"github.com/AbhiramiRajeev/AIRO-Analyzer/config"
	db "github.com/AbhiramiRajeev/AIRO-Analyzer/internal/db"
	"github.com/AbhiramiRajeev/AIRO-Analyzer/internal/models"
	redis "github.com/AbhiramiRajeev/AIRO-Analyzer/internal/redis"
)

type AnalyzerService struct {
	RedisClient redis.RedisClient
	Respository db.Repository
	config      config.Config
}

func NewAnalyzerService(cfg *config.Config, redisClient redis.RedisClient, repository db.Repository) *AnalyzerService {
	return &AnalyzerService{
		RedisClient: redisClient,
		Respository: repository,
		config:      *cfg,
	}
}

func (a *AnalyzerService) Analyze(event models.Event) error {
	Ip := event.IpAddress
	timestamp := float64(time.Now().Unix())
	username := event.UserID
	if event.Status == "failure" {
		err := a.RedisClient.AddData(username, timestamp)
		if err != nil {
			return err
		}
		err = a.RedisClient.RemOldFailues(username, timestamp)
		if err != nil {
			return err
		}
		count, err := a.RedisClient.GetFailedCount(username)
		if err != nil {
			return err
		}
		if count >= a.config.FailureThreshold {
			incident := models.Incident{
				UserID:    username,
				IpAddress: Ip,
				EventType: event.EventType,
				Timestamp: time.Now(),
				Details:   "Multiple failed login attempts detected",
			}
			err = a.Respository.AddIncident(incident)
			if err != nil {
				return err
			}
			return nil // Incident detected
		}

	}

	IsSuspiciousIp, err := a.RedisClient.IsSuspiciousIp(Ip)
	if err != nil {
		return err
	}
	if IsSuspiciousIp {
		incident := models.Incident{
			UserID:    username,
			IpAddress: Ip,
			EventType: event.EventType,
			Timestamp: time.Now(),
			Details:   "Suspicious IP address detected",
		}
		err = a.Respository.AddIncident(incident)
		if err != nil {
			return err
		}
		return nil // Incident detected
	}
	return nil

}
