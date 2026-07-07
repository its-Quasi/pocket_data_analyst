package db

import (
	"fmt"
)

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
