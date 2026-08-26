package store

import (
	"database/sql"
	"errors"

	"task251-chartambiguity/internal/model"
)

// CreateFigure 写入一篇新图形稿（初始状态 importing）。
func (s *Store) CreateFigure(f *model.Figure) error {
	_, err := s.db.Exec(
		`INSERT INTO figures (id,title,status,source_fp,layer_count,created_at,updated_at)
		 VALUES (?,?,?,?,?,?,?)`,
		f.ID, f.Title, f.Status, f.SourceFP, f.LayerCount, f.CreatedAt, f.UpdatedAt,
	)
	return err
}

// GetFigure 按 ID 读取图形稿，不存在返回 model.ErrNotFound。
func (s *Store) GetFigure(id string) (*model.Figure, error) {
	row := s.db.QueryRow(
		`SELECT id,title,status,source_fp,layer_count,created_at,updated_at FROM figures WHERE id=?`, id)
	f := &model.Figure{}
	if err := row.Scan(&f.ID, &f.Title, &f.Status, &f.SourceFP, &f.LayerCount, &f.CreatedAt, &f.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return f, nil
}

// ListFigures 返回图形稿列表（按创建时间倒序，最多 limit 条；limit<=0 表示不限）。
func (s *Store) ListFigures(limit int) ([]model.Figure, error) {
	q := `SELECT id,title,status,source_fp,layer_count,created_at,updated_at FROM figures ORDER BY created_at DESC`
	if limit > 0 {
		q += " LIMIT " + itoa(limit)
	}
	rows, err := s.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Figure
	for rows.Next() {
		f := model.Figure{}
		if err := rows.Scan(&f.ID, &f.Title, &f.Status, &f.SourceFP, &f.LayerCount, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// UpdateFigureStatus 更新图形稿状态。
func (s *Store) UpdateFigureStatus(id, status string) error {
	res, err := s.db.Exec(`UPDATE figures SET status=?, updated_at=strftime('%s','now') WHERE id=?`, status, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// SetFigureSourceFP 写入导入语义指纹与图层计数，并刷新更新时间。
func (s *Store) SetFigureSourceFP(id, fp string, layerCount int) error {
	res, err := s.db.Exec(`UPDATE figures SET source_fp=?, layer_count=?, updated_at=strftime('%s','now') WHERE id=?`, fp, layerCount, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// CreateLayer 写入一个图层。
func (s *Store) CreateLayer(l *model.Layer) error {
	_, err := s.db.Exec(
		`INSERT INTO layers (id,figure_id,name,layer_type,z_order,visible) VALUES (?,?,?,?,?,?)`,
		l.ID, l.FigureID, l.Name, l.LayerType, l.ZOrder, boolToInt(l.Visible),
	)
	return err
}

// ListLayers 返回某图形稿的全部图层（按 z_order 升序）。
func (s *Store) ListLayers(figureID string) ([]model.Layer, error) {
	rows, err := s.db.Query(
		`SELECT id,figure_id,name,layer_type,z_order,visible FROM layers WHERE figure_id=? ORDER BY z_order ASC`, figureID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Layer
	for rows.Next() {
		l := model.Layer{}
		var vis int
		if err := rows.Scan(&l.ID, &l.FigureID, &l.Name, &l.LayerType, &l.ZOrder, &vis); err != nil {
			return nil, err
		}
		l.Visible = vis != 0
		out = append(out, l)
	}
	return out, rows.Err()
}

// CreateAxis 写入一条坐标轴。
func (s *Store) CreateAxis(a *model.Axis) error {
	_, err := s.db.Exec(
		`INSERT INTO axes (id,figure_id,name,variable,unit,orientation) VALUES (?,?,?,?,?,?)`,
		a.ID, a.FigureID, a.Name, a.Variable, a.Unit, a.Orientation,
	)
	return err
}

// ListAxes 返回某图形稿的全部坐标轴。
func (s *Store) ListAxes(figureID string) ([]model.Axis, error) {
	rows, err := s.db.Query(
		`SELECT id,figure_id,name,variable,unit,orientation FROM axes WHERE figure_id=? ORDER BY name ASC`, figureID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Axis
	for rows.Next() {
		a := model.Axis{}
		if err := rows.Scan(&a.ID, &a.FigureID, &a.Name, &a.Variable, &a.Unit, &a.Orientation); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}
