// internal/routes/attendance.go

package routes

import (
	"embed"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/3rr0r-505/Presenz/internal/config"
	"github.com/3rr0r-505/Presenz/internal/models"
	"github.com/3rr0r-505/Presenz/internal/security"
	"github.com/3rr0r-505/Presenz/internal/services"
)

type AttendanceHandler struct {
	Cfg     *config.Config
	Session *services.Session
	DB      *services.DB
	Entry   embed.FS
}

// ServeEntry handles GET / — serves the embedded entry.html form.
func (h *AttendanceHandler) ServeEntry(w http.ResponseWriter, r *http.Request) {
	data, err := h.Entry.ReadFile("entry.html")
	if err != nil {
		writeError(w, http.StatusNotFound, "Entry form not found")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}

// SubmitAttendance handles POST /submit.
func (h *AttendanceHandler) SubmitAttendance(w http.ResponseWriter, r *http.Request) {
	var payload models.AttendanceRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	name, err := security.ValidateName(payload.Name, h.Cfg.Security.MaxNameLength)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	roll, err := security.ValidateRoll(payload.Roll, h.Cfg.Security.MaxRollLength)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := security.ValidateSessionCode(payload.SessionCode, h.Session.GetSessionCode()); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	tableName := h.Session.GetTableName()

	if !h.Session.TryAcceptSubmission() {
		writeJSON(w, http.StatusOK, models.StatusResponse{
			Status:  "closed",
			Message: "Attendance session closed",
		})
		return
	}

	if err := h.DB.InsertAttendance(tableName, name, roll); err != nil {
		if isUniqueConstraintErr(err) {
			h.Session.DecrementCount()
			writeError(w, http.StatusConflict, "Roll number already submitted")
			return
		}
		h.Session.DecrementCount()
		// unexpected DB failure — don't leak err details to client
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if h.Session.IsFull() {
		println("+------------------------------------------------------------+")
		println("| [DEBUG] All responses submitted.                            |")
		println("+------------------------------------------------------------+")
	}

	writeJSON(w, http.StatusOK, models.StatusResponse{
		Status:  "success",
		Message: "Attendance recorded successfully",
	})
}

// isUniqueConstraintErr checks whether a DB error is a UNIQUE constraint
// violation on roll. modernc.org/sqlite doesn't expose a clean typed
// error for this, so it's a string check against the driver's message.
func isUniqueConstraintErr(err error) bool {
	return strings.Contains(err.Error(), "UNIQUE constraint")
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, detail string) {
	writeJSON(w, status, models.ErrorResponse{Detail: detail})
}
