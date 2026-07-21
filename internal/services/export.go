// internal/services/export.go

package services

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/3rr0r-505/Presenz/internal/config"
	"github.com/3rr0r-505/Presenz/internal/models"
)

// Export writes attendance records to both CSV and JSON in the configured
// backup directory, named after the session's table name.
func Export(cfg *config.Config, tableName string, records []models.AttendanceEntry) error {
	backupDir := cfg.Export.BackupPath

	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return fmt.Errorf("creating backup dir: %w", err)
	}

	csvPath := filepath.Join(backupDir, tableName+".csv")
	if err := writeCSV(csvPath, records); err != nil {
		return fmt.Errorf("writing csv: %w", err)
	}

	jsonPath := filepath.Join(backupDir, tableName+".json")
	if err := writeJSON(jsonPath, records); err != nil {
		return fmt.Errorf("writing json: %w", err)
	}

	fmt.Println("[Presenz] Exported to", csvPath)
	fmt.Println("[Presenz] Exported to", jsonPath)

	return nil
}

func writeCSV(path string, records []models.AttendanceEntry) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write([]string{"name", "roll", "timestamp"}); err != nil {
		return err
	}
	for _, r := range records {
		row := []string{r.Name, r.Roll, r.Timestamp.String()}
		if err := w.Write(row); err != nil {
			return err
		}
	}
	return w.Error()
}

func writeJSON(path string, records []models.AttendanceEntry) error {
	type entry struct {
		Name      string `json:"name"`
		Roll      string `json:"roll"`
		Timestamp string `json:"timestamp"`
	}

	structured := make([]entry, len(records))
	for i, r := range records {
		structured[i] = entry{Name: r.Name, Roll: r.Roll, Timestamp: r.Timestamp.String()}
	}

	data, err := json.MarshalIndent(structured, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}
