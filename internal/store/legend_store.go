package store

import (
	"database/sql"
	"errors"

	"task251-chartambiguity/internal/model"
)

// CreateLegend 写入一个图例条目。
func (s *Store) CreateLegend(l *model.Legend) error {
	_, err := s.db.Exec(
		`INSERT INTO legends (id,figure_id,channel,label,token,covers_variable) VALUES (?,?,?,?,?,?)`,
		l.ID, l.FigureID, l.Channel, l.Label, l.Token, l.CoversVariable,
	)
	return err
}

// ListLegends 返回某图形稿的全部图例（按 channel, token 升序）。
func (s *Store) ListLegends(figureID string) ([]model.Legend, error) {
	rows, err := s.db.Query(
		`SELECT id,figure_id,channel,label,token,covers_variable FROM legends WHERE figure_id=? ORDER BY channel,token`, figureID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Legend
	for rows.Next() {
		l := model.Legend{}
		if err := rows.Scan(&l.ID, &l.FigureID, &l.Channel, &l.Label, &l.Token, &l.CoversVariable); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

// CreateVariable 写入一个变量声明。
func (s *Store) CreateVariable(v *model.Variable) error {
	_, err := s.db.Exec(
		`INSERT INTO variables (id,figure_id,name,unit,description) VALUES (?,?,?,?,?)`,
		v.ID, v.FigureID, v.Name, v.Unit, v.Description,
	)
	return err
}

// ListVariables 返回某图形稿的全部变量声明。
func (s *Store) ListVariables(figureID string) ([]model.Variable, error) {
	rows, err := s.db.Query(
		`SELECT id,figure_id,name,unit,description FROM variables WHERE figure_id=? ORDER BY name ASC`, figureID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Variable
	for rows.Next() {
		v := model.Variable{}
		if err := rows.Scan(&v.ID, &v.FigureID, &v.Name, &v.Unit, &v.Description); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// GetVariable 按名称读取某图形稿的变量声明，不存在返回 model.ErrNotFound。
func (s *Store) GetVariable(figureID, name string) (*model.Variable, error) {
	row := s.db.QueryRow(
		`SELECT id,figure_id,name,unit,description FROM variables WHERE figure_id=? AND name=?`, figureID, name)
	v := &model.Variable{}
	if err := row.Scan(&v.ID, &v.FigureID, &v.Name, &v.Unit, &v.Description); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return v, nil
}
