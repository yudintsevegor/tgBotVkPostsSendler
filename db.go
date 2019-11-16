package tgBotVkPostSendler

import (
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
)

const (
	createTable = `
	CREATE TABLE %s (
		ID SERIAL PRIMARY KEY,
		Text TEXT,
		IsPosted BOOLEAN
	);
	`

	deleteTable = `DROP TABLE %s`

	isTableExists = `
	SELECT EXISTS (
		SELECT 1
		FROM   information_schema.tables 
		WHERE  table_schema = 'public'
		AND    table_name = '%s'
		);`
)

type Writer struct {
	DB             *sql.DB
	TableName      string
	CreateNewTable bool

	// id is a ID of post, which VK API sends in response
	id string
	// text is a message, which sends in telegram channel
	text string
	// isPosted determines is message posted in telegram channel
	isPosted bool
}

func (w *Writer) CreateTable() (sql.Result, error) {
	isExistsQuery := fmt.Sprintf(isTableExists, w.TableName)
	var isExists bool
	row := w.DB.QueryRow(isExistsQuery)
	if err := row.Scan(&isExists); err != nil {
		return nil, err
	}

	if isExists && !w.CreateNewTable {
		return nil, nil
	}

	if isExists {
		query := fmt.Sprintf(deleteTable, w.TableName)
		_, err := w.DB.Exec(query)
		if err != nil {
			return nil, errors.Wrapf(err, "Query: %v", query)
		}
	}

	query := fmt.Sprintf(createTable, w.TableName)
	res, err := w.DB.Exec(query)
	if err != nil {
		return nil, errors.Wrapf(err, "Query: %v", query)
	}

	return res, err
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
	query := fmt.Sprintf("SELECT ID, Text FROM %s WHERE IsPosted = false;", w.TableName)

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
