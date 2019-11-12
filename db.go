package tgBotVkPostSendler

import (
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
)

const (
	CreateTable = `
	CREATE TABLE %s (
		ID SERIAL PRIMARY KEY,
		DateTime INTEGER,
		IsPosted BOOLEAN
	);
	`

	DeleteTable = `DROP TABLE %s`
)

var mapReq = map[string]struct{}{
	DeleteTable: struct{}{},
	CreateTable: struct{}{},
}

type Writer struct {
	DB        *sql.DB
	TableName string

	id       string
	dateTime int
	isPosted bool
}

func (w *Writer) EditTable(mode string) (sql.Result, error) {
	if validateReq(mode) {
		return nil, errors.New("[ERR] validation err")
	}

	query := fmt.Sprintf(mode, w.TableName)
	return w.DB.Exec(query)
}

func validateReq(mode string) bool {
	if _, ok := mapReq[mode]; !ok {
		return false
	}
	return true
}

func (w *Writer) WriteToDB() error {
	query := fmt.Sprintf("INSERT INTO %s (ID, DateTime, IsPosted) VALUES ($1, $2, $3);", w.TableName)

	stmnt, err := w.DB.Prepare(query)
	if err != nil {
		return err
	}

	_, err = stmnt.Exec(w.id, w.dateTime, w.isPosted)
	if err != nil {
		return err
	}

	return nil
}

func (w *Writer) UpdateStatus() error {
	query := fmt.Sprintf("UPDATE %s SET IsPosted = $1 WHERE ID = $2;", w.TableName)

	stmnt, err := w.DB.Prepare(query)
	if err != nil {
		return err
	}

	_, err = stmnt.Exec(w.isPosted, w.id)
	if err != nil {
		return err
	}

	return nil
}

func (w *Writer) SelectRows() ([]string, error) {
	query := fmt.Sprintf("SELECT ID FROM %s WHERE IsPosted = false;", w.TableName)

	rows, err := w.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]string, 0, 1)

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}
