package neoread

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type translationPreferences struct {
	Enabled      *bool `json:"translationEnabled"`
	AutoMaxPages *int  `json:"autoTranslateMaxPages"`
}

func decodeTranslationPreferences(raw []byte) translationPreferences {
	var preferences translationPreferences
	if json.Unmarshal(raw, &preferences) != nil {
		return translationPreferences{}
	}
	return preferences
}

func (s *Server) failTranslationDisabled(w http.ResponseWriter) {
	fail(w, http.StatusConflict, "Translation features are disabled")
}

func (s *Store) translationPreferences() translationPreferences {
	var raw string
	_ = s.DB.QueryRow("SELECT value FROM settings WHERE key='neoread'").Scan(&raw)
	return decodeTranslationPreferences([]byte(raw))
}

func (p translationPreferences) enabled() bool { return p.Enabled == nil || *p.Enabled }

// An absent limit leaves legacy queued work eligible. New imports still default to manual.
func (p translationPreferences) automatic() bool { return p.AutoMaxPages == nil || *p.AutoMaxPages > 0 }

func automaticTranslation(p Processing) bool {
	return !p.ManualTranslationRequested && !p.Retranslating && len(p.TranslationBlocks) == 0
}

func (p translationPreferences) applyToView(job *Processing) {
	if job.Phase != "translating" || job.Status == "complete" || (p.enabled() && (p.automatic() || !automaticTranslation(*job))) {
		return
	}
	job.Enabled = false
	job.ManualTranslation = p.enabled()
	job.Status, job.Detail = "paused", "Translation paused in settings"
	job.Translating = &ProcessingStage{Status: "paused", Detail: job.Detail}
}

// The caller holds processingMu, so no new worker can start after preferences
// are saved and before cancellation has been requested.
func (s *Server) pauseDisabledTranslation(preferences translationPreferences) {
	count := 0
	for _, task := range s.activeProcessing {
		if task.translating && (!preferences.enabled() || (!preferences.automatic() && task.automaticTranslation)) {
			task.cancel()
			count++
		}
	}
	if count > 0 {
		s.diagnosticEvent(slog.LevelInfo, "translation.preferences.pause", "count", count)
	}
}
