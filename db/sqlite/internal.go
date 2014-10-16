package sqlite

import (
	"sync"
	"strings"

	sqlite3 "github.com/mxk/go-sqlite/sqlite3"
	"github.com/ottemo/foundation/db"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
)

// locks given connection
func connectionLock(transactionId string, connection *sqlite3.Conn) *sqlite3.Conn {
	dbEngine.connectionMutex[connection].Lock()
	return connection
}

// locks given connection
func connectionUnlock(transactionId string, connection *sqlite3.Conn) *sqlite3.Conn {
	dbEngine.connectionMutex[connection].Unlock()
	return connection
}

// returns released DB connection to make SQL Query
func getConnection(transactionId string) *sqlite3.Conn {

	dbEngine.engineMutex.Lock()
	defer dbEngine.engineMutex.Unlock()

	if transactionId != "" {
		if transactionConnection, present := dbEngine.transactions[transactionId]; present {
			return transactionConnection
		}
	}

	bestCandidate := dbEngine.connectionPool[0]
	lowestQueue := 0

	for connection, connectionMutex := range dbEngine.connectionMutex {
		connectionQueue := connectionMutex.readerCount

		if lowestQueue == 0 || lowestQueue > connectionQueue {
			lowestQueue = connectionQueue
			bestCandidate = connection
		}

		if connectionQueue == 0 {
			break
		}
	}

	if lowestQueue != 0 && dbEngine.maxConnections < len(dbEngine.connectionPool) {
		newConnection, err := sqlite3.Open(dbEngine.uri)
		if err == nil {
			dbEngine.connectionPool = append(dbEngine.connectionPool, newConnection)
			dbEngine.connectionMutex[newConnection] = new(sync.RWMutex)

			bestCandidate = newConnection
		}
	}

	return bestCandidate
}

// executes synchronized SQL with returning last inserted id
func connectionExecWLastInsertId(transactionId string, SQL string, args ...interface{}) (int64, error) {
	connection := connectionLock(transactionId, getConnection(transactionId))
	defer connectionUnlock(transactionId, connection)

	err := dbEngine.connection.Exec(SQL, args...)
	if err != nil {
		return dbEngine.connection.LastInsertId(), err
	}

	return 0, err
}

// executes synchronized SQL with returning amount of affected rows
func connectionExecWAffected(transactionId string, SQL string, args ...interface{}) (int, error) {
	connection := connectionLock(transactionId, getConnection(transactionId))
	defer connectionUnlock(transactionId, connection)

	err := connection.Exec(SQL, args...)
	if err != nil {
		return connection.RowsAffected(), err
	}

	return 0, err
}

// executes synchronized SQL
func connectionExec(transactionId string, SQL string, args ...interface{}) error {
	connection := connectionLock(transactionId, getConnection(transactionId))
	defer connectionUnlock(transactionId, connection)

	return connection.Exec(SQL, args...)
}

// executes SQL with setting lock to connection
func connectionQuery(transactionId string, SQL string) (*sqlite3.Stmt, error) {
	connection := connectionLock(transactionId, getConnection(transactionId))
	return connection.Query(SQL)
}

// closes SQL query statement with unlocking connection
func closeStatement(transactionId string, statement *sqlite3.Stmt) {
	if statement != nil {
		statement.Close()
	}
	connectionUnlock(transactionId, dbEngine.statements[statement.Conn()])
}

// formats SQL query error for output to log
func sqlError(SQL string, err error) error {
	return env.ErrorNew("SQL \"" + SQL + "\" error: " + err.Error())
}

// returns string that represents value for SQL query
func convertValueForSQL(value interface{}) string {

	switch value.(type) {
	case bool:
		if value.(bool) {
			return "1"
		}
		return "0"

	case string:
		result := value.(string)
		result = strings.Replace(result, "'", "''", -1)
		result = strings.Replace(result, "\\", "\\\\", -1)
		result = "'" + result + "'"

		return result

	case int, int32, int64:
		return utils.InterfaceToString(value)

	case map[string]interface{}, map[string]string:
		return convertValueForSQL(utils.EncodeToJsonString(value))

	case []string, []int, []int64, []int32, []float64, []bool:
		return convertValueForSQL(utils.InterfaceToArray(value))

	case []interface{}:
		result := ""
		for _, item := range value.([]interface{}) {
			if result != "" {
				result += ", "
			}
			result += utils.InterfaceToString(item)
		}
		return convertValueForSQL(result)
	}

	return convertValueForSQL(utils.InterfaceToString(value))
}

// returns type used inside sqlite for given general name
func GetDBType(ColumnType string) (string, error) {
	ColumnType = strings.ToLower(ColumnType)
	switch {
	case strings.HasPrefix(ColumnType, "[]"):
		return "TEXT", nil
	case ColumnType == db.DB_BASETYPE_ID:
		if UUID_ID {
			return "TEXT", nil
		} else {
			return "INTEGER", nil
		}
	case ColumnType == "int" || ColumnType == "integer":
		return "INTEGER", nil
	case ColumnType == "real" || ColumnType == "float":
		return "REAL", nil
	case ColumnType == "string" || ColumnType == "text" || strings.Contains(ColumnType, "char"):
		return "TEXT", nil
	case ColumnType == "blob" || ColumnType == "struct" || ColumnType == "data":
		return "BLOB", nil
	case strings.Contains(ColumnType, "numeric") || strings.Contains(ColumnType, "decimal") || ColumnType == "money":
		return "NUMERIC", nil
	case strings.Contains(ColumnType, "date") || strings.Contains(ColumnType, "time"):
		return "NUMERIC", nil
	case ColumnType == "bool" || ColumnType == "boolean":
		return "NUMERIC", nil
	}

	return "?", env.ErrorNew("Unknown type '" + ColumnType + "'")
}
