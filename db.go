package tgBotVkPostSendler

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

const (
	CreateTable = `
	CREATE TABLE %s (
		ID SERIAL PRIMARY KEY,
		Text TEXT,
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

	offset   string
	id       string
	text     string
	isPosted bool
}

func (w *Writer) setDbOffset(duration time.Duration, count int) {
	w.offset = "20"
}

func (w *Writer) EditTable(mode string) (sql.Result, error) {
	if !validateReq(mode) {
		return nil, errors.New("[ERR] validation err")
	}

	query := fmt.Sprintf(mode, w.TableName)
	res, err := w.DB.Exec(query)
	if err != nil {
		return nil, errors.Wrapf(err, "Query: %v", query)
	}
	return res, err
}

func validateReq(mode string) bool {
	if _, ok := mapReq[mode]; !ok {
		return false
	}
	return true
}

func (w *Writer) InsertToDb() error {
	query := fmt.Sprintf("INSERT INTO %s (ID, Text, IsPosted) VALUES ($1, $2, $3);", w.TableName)

	stmnt, err := w.DB.Prepare(query)
	if err != nil {
		return errors.Wrapf(err, "Query: %v", query)
	}

	_, err = stmnt.Exec(w.id, w.text, w.isPosted)
	if err != nil {
		return errors.Wrapf(err, "Query: %v", query)
	}

	return nil
}

func (w *Writer) UpdateStatus(id string) error {
	query := fmt.Sprintf("UPDATE %s SET IsPosted = true WHERE ID = $1;", w.TableName)

	stmnt, err := w.DB.Prepare(query)
	if err != nil {
		return err
	}

	_, err = stmnt.Exec(id)
	if err != nil {
		return errors.Wrapf(err, "Query: %v", query)
	}

	return nil
}

func (w *Writer) SelectRows() (map[string]struct{}, error) {
	query := fmt.Sprintf("SELECT ID FROM %s WHERE IsPosted = true;", w.TableName)

	rows, err := w.DB.Query(query)
	if err != nil {
		return nil, errors.Wrapf(err, "Query: %v", query)
	}
	defer rows.Close()

	ids := make(map[string]struct{})

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids[id] = struct{}{}
	}

	return ids, nil
}

func (w *Writer) SelectOldRows() ([]Message, error) {
	query := fmt.Sprintf("SELECT ID, Text FROM %s WHERE IsPosted = false OFFSET %s;", w.TableName, w.offset)

	rows, err := w.DB.Query(query)
	if err != nil {
		return nil, errors.Wrapf(err, "Query: %v", query)
	}
	defer rows.Close()

	messages := make([]Message, 0, 1)

	for rows.Next() {
		mes := new(Message)
		if err := rows.Scan(mes.ID, mes.Text); err != nil {
			return nil, err
		}
		messages = append(messages, *mes)
	}

	return messages, nil
}
