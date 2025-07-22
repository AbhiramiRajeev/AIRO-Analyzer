package models

import "time"


type Incident struct {
	UserID          string `json:"user_id"`
	IpAddress       string `json:"ip_address"`
	EventType       string `json:"event_type"`
	Timestamp       time.Time `json:"timestamp"`
	Details		    string `json:"details"`
}


type Event struct {
	EventType string   `json:"event_type" binding:"required"`
	UserID    string   `json:"user_id" binding:"required"`
	IpAddress string   `json:"ip_address" binding:"required"`
	Status	  string   `json:"status" binding:"required"`
	Timestamp string   `json:"timestamp" binding:"required"`
} 

