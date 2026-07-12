package domain

import "fmt"

// DBType representa el tipo de base de datos soportada.
type DBType string

const (
	MySQL DBType = "mysql"
)

// ConnectionConfig guarda las credenciales y parámetros de conexión.
type ConnectionConfig struct {
	Type     DBType
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

// DSN construye el Data Source Name para el driver de MySQL.
func (c ConnectionConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.Database,
	)
}

// TableDDL guarda el nombre y el DDL de una tabla.
type TableDDL struct {
	Name       string
	CreateStmt string
}

// DDLInfo contiene el DDL completo de una base de datos.
type DDLInfo struct {
	DatabaseName string
	Tables       []TableDDL
}

// ToContextString convierte toda la info de DDL en un string para el contexto del LLM.
func (d *DDLInfo) ToContextString() string {
	var out string
	out += "Database: " + d.DatabaseName + "\n\n"
	for _, t := range d.Tables {
		out += "--- Table: " + t.Name + " ---\n"
		out += t.CreateStmt + "\n\n"
	}
	return out
}
