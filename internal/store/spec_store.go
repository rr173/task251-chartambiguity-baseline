package store

import (
	"database/sql"

	"task251-chartambiguity/internal/model"
)

// ---- 例外 ----

// CreateException 写入一个豁免声明。
func (s *Store) CreateException(e *model.Exception) error {
	_, err := s.db.Exec(
		`INSERT INTO exceptions (id,figure_id,kind,target_channel,target_token,target_variable,reason,created_at)
		 VALUES (?,?,?,?,?,?,?,?)`,
		e.ID, e.FigureID, e.Kind, e.TargetChannel, e.TargetToken, e.TargetVariable, e.Reason, e.CreatedAt,
	)
	return err
}

// ListExceptions 返回某图形稿的全部豁免声明。
func (s *Store) ListExceptions(figureID string) ([]model.Exception, error) {
	rows, err := s.db.Query(
		`SELECT id,figure_id,kind,target_channel,target_token,target_variable,reason,created_at
		 FROM exceptions WHERE figure_id=? ORDER BY created_at ASC`, figureID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Exception
	for rows.Next() {
		e := model.Exception{}
		if err := rows.Scan(&e.ID, &e.FigureID, &e.Kind, &e.TargetChannel, &e.TargetToken, &e.TargetVariable, &e.Reason, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// ---- 歧义 ----

// DeleteAmbiguities 删除某图形稿的全部歧义记录（重算前清理）。
func (s *Store) DeleteAmbiguities(figureID string) error {
	_, err := s.db.Exec(`DELETE FROM ambiguities WHERE figure_id=?`, figureID)
	return err
}

// InsertAmbiguities 批量写入歧义记录。
func (s *Store) InsertAmbiguities(list []model.Ambiguity) error {
	for i := range list {
		a := list[i]
		if _, err := s.db.Exec(
			`INSERT INTO ambiguities (id,figure_id,type,severity,channel,token,variables,description,resolved,exception_id,created_at)
			 VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
			a.ID, a.FigureID, a.Type, a.Severity, a.Channel, a.Token, a.Variables, a.Description,
			boolToInt(a.Resolved), a.ExceptionID, a.CreatedAt,
		); err != nil {
			return err
		}
	}
	return nil
}

// ListAmbiguities 返回某图形稿的全部歧义（含已豁免）。
func (s *Store) ListAmbiguities(figureID string) ([]model.Ambiguity, error) {
	rows, err := s.db.Query(
		`SELECT id,figure_id,type,severity,channel,token,variables,description,resolved,exception_id,created_at
		 FROM ambiguities WHERE figure_id=? ORDER BY severity,type`, figureID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Ambiguity
	for rows.Next() {
		a := model.Ambiguity{}
		var resolved, created int64
		if err := rows.Scan(&a.ID, &a.FigureID, &a.Type, &a.Severity, &a.Channel, &a.Token, &a.Variables, &a.Description, &resolved, &a.ExceptionID, &created); err != nil {
			return nil, err
		}
		a.Resolved = resolved != 0
		a.CreatedAt = created
		out = append(out, a)
	}
	return out, rows.Err()
}

// CountOpenAmbiguities 统计某图形稿仍未解决的歧义数量。
func (s *Store) CountOpenAmbiguities(figureID string) (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM ambiguities WHERE figure_id=? AND resolved=0`, figureID).Scan(&n)
	return n, err
}

// ---- 图规范 ----

// CreateSpec 写入一个图规范版本。
func (s *Store) CreateSpec(sp *model.FigureSpec) error {
	_, err := s.db.Exec(
		`INSERT INTO specs (id,figure_id,version,status,snapshot,created_at) VALUES (?,?,?,?,?,?)`,
		sp.ID, sp.FigureID, sp.Version, sp.Status, sp.Snapshot, sp.CreatedAt,
	)
	return err
}

// GetSpec 按 ID 读取图规范，不存在返回 model.ErrNotFound。
func (s *Store) GetSpec(id string) (*model.FigureSpec, error) {
	row := s.db.QueryRow(
		`SELECT id,figure_id,version,status,snapshot,created_at FROM specs WHERE id=?`, id)
	sp := &model.FigureSpec{}
	if err := row.Scan(&sp.ID, &sp.FigureID, &sp.Version, &sp.Status, &sp.Snapshot, &sp.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return sp, nil
}

// ListSpecs 返回某图形稿的全部规范版本（按版本升序）。
func (s *Store) ListSpecs(figureID string) ([]model.FigureSpec, error) {
	rows, err := s.db.Query(
		`SELECT id,figure_id,version,status,snapshot,created_at FROM specs WHERE figure_id=? ORDER BY version ASC`, figureID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.FigureSpec
	for rows.Next() {
		sp := model.FigureSpec{}
		if err := rows.Scan(&sp.ID, &sp.FigureID, &sp.Version, &sp.Status, &sp.Snapshot, &sp.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, sp)
	}
	return out, rows.Err()
}

// UpdateSpecStatus 更新规范状态。
func (s *Store) UpdateSpecStatus(id, status string) error {
	res, err := s.db.Exec(`UPDATE specs SET status=? WHERE id=?`, status, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// UpdateSpecSnapshot 更新规范在冻结前即将固化的当前语义快照。
func (s *Store) UpdateSpecSnapshot(id, snapshot string) error {
	res, err := s.db.Exec(`UPDATE specs SET snapshot=? WHERE id=?`, snapshot, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// SupersedeFrozen 将某图形稿除 keepID 外的全部 frozen 规范置为 superseded。
func (s *Store) SupersedeFrozen(figureID, keepID string) error {
	_, err := s.db.Exec(
		`UPDATE specs SET status=? WHERE figure_id=? AND status=? AND id<>?`,
		model.SpecStatusSuperseded, figureID, model.SpecStatusFrozen, keepID,
	)
	return err
}

// MaxSpecVersion 返回某图形稿当前最大版本号（无记录时返回 0）。
func (s *Store) MaxSpecVersion(figureID string) (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COALESCE(MAX(version),0) FROM specs WHERE figure_id=?`, figureID).Scan(&n)
	return n, err
}
