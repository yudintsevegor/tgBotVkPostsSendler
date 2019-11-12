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
		Text STRING
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
	text     string
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

func (w *Writer) InsertToDb() error {
	query := fmt.Sprintf("INSERT INTO %s (ID, Text, IsPosted) VALUES ($1, $2, $3, $4);", w.TableName)

	stmnt, err := w.DB.Prepare(query)
	if err != nil {
		return err
	}

	_, err = stmnt.Exec(w.id, w.text, w.isPosted)
	if err != nil {
		return err
	}

	return nil
}

func (w *Writer) UpdateStatus() error {
	query := fmt.Sprintf("UPDATE %s SET IsPosted = true WHERE ID = $2;", w.TableName)

	stmnt, err := w.DB.Prepare(query)
	if err != nil {
		return err
	}

	_, err = stmnt.Exec(w.id)
	if err != nil {
		return err
	}

	return nil
}

func (w *Writer) SelectRows() (map[string]struct{}, error) {
	query := fmt.Sprintf("SELECT ID FROM %s WHERE IsPosted = true;", w.TableName)

	rows, err := w.DB.Query(query)
	if err != nil {
		return nil, err
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
	query := fmt.Sprintf("SELECT ID, Text FROM %s WHERE IsPosted = false OFFSET 100;", w.TableName)

	rows, err := w.DB.Query(query)
	if err != nil {
		return nil, err
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
