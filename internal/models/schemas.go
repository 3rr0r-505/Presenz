// internal/models/schemas.go

package models

import "time"

type AttendanceRequest struct {
	Name        string `json:"name"`
	Roll        string `json:"roll"`
	SessionCode string `json:"session_code"`
}

type AttendanceEntry struct {
	Name      string
	Roll      string
	Timestamp time.Time
}

type StatusResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Detail string `json:"detail"`
}
