package grunway

import (
	"database/sql"

	"github.com/amattn/deeperror"
)

func PrepareStatements(DB *sql.DB, queries map[string]**sql.Stmt) error {
	for query, stmtPtrAddress := range queries {
		preparedStmt, err := DB.Prepare(query)
		*stmtPtrAddress = preparedStmt
		if err != nil {
			return deeperror.New(322916003, "failure to prepare query:"+query, err)
		}
	}

	return nil
}
