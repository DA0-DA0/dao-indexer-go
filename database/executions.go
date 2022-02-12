package database

import (
	"fmt"

	"github.com/disperze/wasmx/types"
)

// SaveExec saves a message execution into the database.
func (db Db) SaveExec(exec_msg types.ExecMsg) error {
	funds, funds_err := exec_msg.Funds.MarshalJSON()
	if funds_err != nil {
		return fmt.Errorf("error while marshaling: %v", funds_err.Error())
	}
	json := exec_msg.Json
	stmt := `
INSERT INTO exec_msg (sender, address, funds, json) 
VALUES ($1, $2, $3, $4)`
	_, err := db.Sql.Exec(stmt, exec_msg.Sender, exec_msg.Address, funds, json)
	return err
}
