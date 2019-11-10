package tgBotVkPostSendler

import (
	"database/sql"
	"fmt"
)

type Writer struct {
	DB *sql.DB
}

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

func (w *Writer) EditTable(mode, tableName string) (sql.Result, error) {
	query := fmt.Sprintf(mode, tableName)
	return w.DB.Exec(query)
}

func (w *Writer) WriteToDB(tableName string) error {
	query := fmt.Sprintf("INSERT INTO %s (ID, DateTime, IsPosted) VALUES ($1, $2, $3);", tableName)

	stmnt, err := w.DB.Prepare(query)
	if err != nil {
		return err
	}

	_, err = stmnt.Exec(1234, 5678, false)
	if err != nil {
		return err
	}

	return nil
}
