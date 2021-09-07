package store

import "time"

type Schedule struct {
	ID            int64     `json:"id"`
	TransactionID int64     `json:"transaction_id"`
	StartingTime  time.Time `json:"starting_time"`
	Frequency     string    `json:"frequency"`             // annually or monthly
	IsCanceled    bool      `json:"is_canceled"`           // default false
	EndingTime    time.Time `json:"ending_time,omitempty"` // optional
	CreatedAt     time.Time `json:"created_at"`
	Version       int       `json:"version"`
	// Reminder      string    `json:"reminder"`
}
