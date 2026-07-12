package database

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"quasi.db_analysis_agent/internal/domain"
)

// MySQLReader se encarga de leer el DDL de una base de datos MySQL.
type MySQLReader struct {
	Config domain.ConnectionConfig
}

// ReadDDL extrae el DDL de todas las tablas de la base de datos configurada.
func (r *MySQLReader) ReadDDL() (*domain.DDLInfo, error) {
	db, err := sql.Open("mysql", r.Config.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open mysql connection: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping mysql: %w", err)
	}

	tables, err := r.listTables(db)
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}

	ddlInfo := &domain.DDLInfo{
		DatabaseName: r.Config.Database,
		Tables:       make([]domain.TableDDL, 0, len(tables)),
	}

	for _, tableName := range tables {
		stmt, err := r.showCreateTable(db, tableName)
		if err != nil {
			return nil, fmt.Errorf("failed to get ddl for table %s: %w", tableName, err)
		}
		ddlInfo.Tables = append(ddlInfo.Tables, domain.TableDDL{
			Name:       tableName,
			CreateStmt: stmt,
		})
	}

	return ddlInfo, nil
}

func (r *MySQLReader) listTables(db *sql.DB) ([]string, error) {
	query := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = ?
		  AND table_type = 'BASE TABLE'
	`
	rows, err := db.Query(query, r.Config.Database)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}
	return tables, rows.Err()
}

func (r *MySQLReader) showCreateTable(db *sql.DB, tableName string) (string, error) {
	query := fmt.Sprintf("SHOW CREATE TABLE `%s`", tableName)
	row := db.QueryRow(query)

	var name, createStmt string
	if err := row.Scan(&name, &createStmt); err != nil {
		return "", err
	}
	return createStmt, nil
}
