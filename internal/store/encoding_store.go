package store

import (
	"task251-chartambiguity/internal/model"
)

// CreateEncoding 写入一个视觉编码声明。
func (s *Store) CreateEncoding(e *model.VisualEncoding) error {
	_, err := s.db.Exec(
		`INSERT INTO encodings (id,figure_id,layer_id,variable,channel,token,status) VALUES (?,?,?,?,?,?,?)`,
		e.ID, e.FigureID, e.LayerID, e.Variable, e.Channel, e.Token, e.Status,
	)
	return err
}

// ListEncodings 返回某图形稿的全部视觉编码（按 channel, token 升序）。
func (s *Store) ListEncodings(figureID string) ([]model.VisualEncoding, error) {
	rows, err := s.db.Query(
		`SELECT id,figure_id,layer_id,variable,channel,token,status FROM encodings WHERE figure_id=? ORDER BY channel,token`, figureID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.VisualEncoding
	for rows.Next() {
		e := model.VisualEncoding{}
		if err := rows.Scan(&e.ID, &e.FigureID, &e.LayerID, &e.Variable, &e.Channel, &e.Token, &e.Status); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// SetEncodingStatus 更新单个编码的状态。
func (s *Store) SetEncodingStatus(id, status string) error {
	_, err := s.db.Exec(`UPDATE encodings SET status=? WHERE id=?`, status, id)
	return err
}

// CreateMapping 写入一个变量-通道映射决议。
func (s *Store) CreateMapping(m *model.VariableMapping) error {
	_, err := s.db.Exec(
		`INSERT INTO mappings (id,figure_id,variable,channel,decision,note) VALUES (?,?,?,?,?,?)`,
		m.ID, m.FigureID, m.Variable, m.Channel, m.Decision, m.Note,
	)
	return err
}

// ListMappings 返回某图形稿的全部变量-通道映射。
func (s *Store) ListMappings(figureID string) ([]model.VariableMapping, error) {
	rows, err := s.db.Query(
		`SELECT id,figure_id,variable,channel,decision,note FROM mappings WHERE figure_id=? ORDER BY variable,channel`, figureID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.VariableMapping
	for rows.Next() {
		m := model.VariableMapping{}
		if err := rows.Scan(&m.ID, &m.FigureID, &m.Variable, &m.Channel, &m.Decision, &m.Note); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// DeleteMappings 删除某图形稿的全部映射（重算前清理）。
func (s *Store) DeleteMappings(figureID string) error {
	_, err := s.db.Exec(`DELETE FROM mappings WHERE figure_id=?`, figureID)
	return err
}
