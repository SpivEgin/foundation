package sqlite

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	sqlite3 "github.com/mxk/go-sqlite/sqlite3"
	"github.com/ottemo/foundation/db"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
)

// locks given connection or transaction mutex
// (we will wait on this function if connection/transaction mutex is busy)
func connectionLock(transactionId string, connection *sqlite3.Conn) *sqlite3.Conn {

	// making queue calculations
	if connection != nil {
		dbEngine.engineMutex.Lock()

		if value, present := dbEngine.connectionQueue[connection]; present {
			dbEngine.connectionQueue[connection] = value + 1
		} else {
			dbEngine.connectionQueue[connection] = 1
		}

		dbEngine.engineMutex.Unlock()
	}

	// locking transaction or connection mutex
	if transactionId != "" {
		if mutex, present := dbEngine.transactionMutex[transactionId]; present {
			mutex.Lock()
		}
	} else {
		if mutex, present := dbEngine.connectionMutex[connection]; present {
			mutex.Lock()
		}
	}

	return connection
}

// unlocks given connection or transaction mutex
func connectionUnlock(transactionId string, connection *sqlite3.Conn) *sqlite3.Conn {

	// making queue calculations
	if connection != nil {
		dbEngine.engineMutex.Lock()

		if value, present := dbEngine.connectionQueue[connection]; present {
			dbEngine.connectionQueue[connection] = value - 1
		} else {
			dbEngine.connectionQueue[connection] = 0
		}

		dbEngine.engineMutex.Unlock()
	}

	// unlocking transaction or connection mutex
	if transactionId != "" {
		if mutex, present := dbEngine.transactionMutex[transactionId]; present {
			mutex.Unlock()
		}
	} else {
		if mutex, present := dbEngine.connectionMutex[connection]; present {
			mutex.Unlock()
		}
	}

	// connection pool cleanup
	randomNumber, err := rand.Int(rand.Reader, big.NewInt(int64(dbEngine.gcRate)))
	if err == nil && randomNumber.Cmp(big.NewInt(1)) == 0 {
		gcConnectionPool()
	}

	return connection
}

// garbage collect unused connections above pool limit
func gcConnectionPool() {
	dbEngine.engineMutex.Lock()
	defer dbEngine.engineMutex.Unlock()

	newPool := make([]*sqlite3.Conn, 0, dbEngine.poolSize)
	wasModified := false
	for idx, connection := range dbEngine.connectionPool {
		if idx > dbEngine.poolSize {
			if queue, present := dbEngine.connectionQueue[connection]; present && queue == 0 {
				wasModified = true
				delete(dbEngine.connectionMutex, connection)
				delete(dbEngine.connectionQueue, connection)
			}
		}
		newPool = append(newPool, connection)
	}

	if wasModified {
		dbEngine.connectionPool = newPool
	}
}

// returns released DB connection to make SQL Query
func getConnection(transactionId string) *sqlite3.Conn {

	// structure variables exclusive access
	dbEngine.engineMutex.Lock()
	defer dbEngine.engineMutex.Unlock()

	// transactions have own static connection
	if transactionId != "" {
		if transactionConnection, present := dbEngine.transactions[transactionId]; present {
			return transactionConnection
		}
	}

	// supposedly we have at least one connection
	bestCandidate := dbEngine.connectionPool[0]
	lowestQueue := 0

	// looking for free connection or at least where queue lower
	for _, connection := range dbEngine.connectionPool {

		connectionQueue := 0

		// we have struct variable to calculate current queue for connection
		if queue, present := dbEngine.connectionQueue[connection]; present {
			connectionQueue = queue
		}

		// if first iteration or better queue value - updating candidate info
		if lowestQueue == 0 || lowestQueue > connectionQueue {
			lowestQueue = connectionQueue
			bestCandidate = connection
		}

		// we found best option
		if connectionQueue == 0 {
			break
		}
	}

	// no free connections - probably we can open new one connection to DB
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

	err := connection.Exec(SQL, args...)
	if err != nil {
		return connection.LastInsertId(), err
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
	// obtaining connection for statement
	connection := connectionLock(transactionId, getConnection(transactionId))

	// assigning statement to connection
	stmt, err := connection.Query(SQL)
	if err == nil && stmt != nil {
		dbEngine.statements[stmt] = connection
	} else {
		connectionUnlock(transactionId, connection)
	}

	return stmt, err
}

// closes SQL query statement with unlocking connection
func closeStatement(transactionId string, statement *sqlite3.Stmt) {
	if statement != nil {
		statement.Close()
	}

	// unlocking connection
	if connection, present := dbEngine.statements[statement]; present {
		connectionUnlock(transactionId, connection)
	} else {
		connectionUnlock(transactionId, nil)
	}
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

// generates 24 byte UUID string based on 8 byte timestamp and 16 byte random numbers
func generateUUID() string {
	timeStamp := strconv.FormatInt(time.Now().Unix(), 16)

	randomBytes := make([]byte, 8)
	rand.Reader.Read(randomBytes)

	randomHex := make([]byte, 16)
	hex.Encode(randomHex, randomBytes)

	return timeStamp + string(randomHex)
}
