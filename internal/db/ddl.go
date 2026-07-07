package db

// DDLInfo contiene el DDL completo de una base de datos.
type DDLInfo struct {
	DatabaseName string
	Tables       []TableDDL
}

// TableDDL guarda el nombre y el DDL de una tabla.
type TableDDL struct {
	Name       string
	CreateStmt string
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
