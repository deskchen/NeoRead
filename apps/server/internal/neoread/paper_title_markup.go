package neoread

import (
	"log/slog"

	"neoread.local/server/internal/scholar"
)

func (s *Store) migratePaperTitleMarkup() error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	type titleChange struct{ id, title string }
	var changes []titleChange
	rows, err := tx.Query(`
		SELECT id, title FROM documents
		WHERE library = 'papers'
		  AND (instr(title, '<') > 0 OR instr(title, '&') > 0)
		  AND CASE WHEN json_valid(metadata) THEN json_extract(metadata, '$.sources.title') END IN ('lookup', 'file')
	`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id, title string
		if err := rows.Scan(&id, &title); err != nil {
			rows.Close()
			return err
		}
		if cleaned := scholar.StripTitleMarkup(title); cleaned != "" && cleaned != title {
			changes = append(changes, titleChange{id, cleaned})
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	for _, change := range changes {
		if _, err := tx.Exec("UPDATE documents SET title=? WHERE id=?", change.title, change.id); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	if len(changes) > 0 {
		slog.Info("paper.title_markup.migrated", "count", len(changes))
	}
	return nil
}
