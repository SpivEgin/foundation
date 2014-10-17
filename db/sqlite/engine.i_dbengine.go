package sqlite

import (
	"strconv"
	"sync"

	"github.com/mxk/go-sqlite/sqlite3"
	"github.com/ottemo/foundation/db"
	"github.com/ottemo/foundation/env"
)

// returns current DB engine name
func (it *SQLite) GetName() string {
	return "Sqlite3"
}

// checks if collection(table) already exists
func (it *SQLite) HasCollection(collectionName string) bool {
	// collectionName = strings.ToLower(collectionName)

	SQL := "SELECT name FROM sqlite_master WHERE type='table' AND name='" + collectionName + "'"

	stmt, err := connectionQuery("", SQL)
	defer closeStatement("", stmt)

	if err == nil {
		return true
	} else {
		return false
	}
}

// creates cllection(table) by it's name
func (it *SQLite) CreateCollection(collectionName string) error {
	// collectionName = strings.ToLower(collectionName)

	SQL := "CREATE TABLE " + collectionName + " (_id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL)"
	if UUID_ID {
		SQL = "CREATE TABLE " + collectionName + " (_id NCHAR(24) PRIMARY KEY NOT NULL)"
	}

	if err := connectionExec("", SQL); err == nil {
		return nil
	} else {
		return env.ErrorDispatch(err)
	}
}

// returns collection(table) by name or creates new one
func (it *SQLite) GetCollection(collectionName string) (db.I_DBCollection, error) {
	if !SQL_NAME_VALIDATOR.MatchString(collectionName) {
		return nil, env.ErrorNew("not valid collection name for DB engine")
	}

	if !it.HasCollection(collectionName) {
		if err := it.CreateCollection(collectionName); err != nil {
			return nil, env.ErrorDispatch(err)
		}
	}

	collection := &SQLiteCollection{
		Name:          collectionName,
		FilterGroups:  make(map[string]*T_DBFilterGroup),
		Order:         make([]string, 0),
		ResultColumns: make([]string, 0),
	}

	return collection, nil
}

// executes SQL Query on DB Engine
//   WARNING: usage of that function is not controlled by application and can broke your DB or hang app,
//            please use it only for select statements and optimization purposes
func (it *SQLite) RawQuery(query string) (map[string]interface{}, error) {
	return it.RawQueryOnTransaction("", query)
}

// executes SQL Query on DB Engine within given transaction
func (it *SQLite) RawQueryOnTransaction(transactionId string, query string) (map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0, 10)

	row := make(sqlite3.RowMap)

	stmt, err := connectionQuery("", query)
	defer closeStatement("", stmt)

	if err == nil {
		return nil, env.ErrorDispatch(err)
	}

	for ; err == nil; err = stmt.Next() {
		if err := stmt.Scan(row); err == nil {

			if UUID_ID {
				if _, present := row["_id"]; present {
					row["_id"] = strconv.FormatInt(row["_id"].(int64), 10)
				}
			}

			result = append(result, row)
		} else {
			return result[0], nil
		}
	}

	return result[0], nil
}

// starts new transaction for DB Engine
func (it *SQLite) BeginTransaction() (string, error) {
	transactionName := generateUUID()
	connection := getConnection(transactionName)

	it.engineMutex.Lock()
	it.transactions[transactionName] = connection
	it.transactionMutex[transactionName] = new(sync.RWMutex)
	it.engineMutex.Unlock()

	connectionLock("", connection) // so we need to lock connection mutex but not transaction
	return transactionName, nil
}

// starts new transaction for DB Engine with given transaction id
//    - transaction id should be unique across concurrent threads/routines
//    - use BeginTransaction() function unless you know what you doing
func (it *SQLite) BeginNamedTransaction(transactionId string) error {
	connection := getConnection(transactionId)

	if transactionId != "" {
		if _, present := it.transactions[transactionId]; present {
			return env.ErrorNew("transaction with id '" + transactionId + "' already exists")
		}

		it.engineMutex.Lock()
		it.transactions[transactionId] = connection
		it.transactionMutex[transactionId] = new(sync.RWMutex)
		it.engineMutex.Unlock()

	} else {
		return env.ErrorNew("transaction id can't be blank")
	}

	connectionLock("", connection) // so we need to lock connection mutex but not transaction mutex
	connection.Exec("BEGIN TRANSACTION")

	return nil
}

// commits modifications made within given transaction
func (it *SQLite) CommitTransaction(transactionId string) error {
	if transactionId != "" {
		if connection, present := it.transactions[transactionId]; present {
			connection.Exec("COMMIT TRANSACTION")

			it.engineMutex.Lock()
			delete(it.transactions, transactionId)
			delete(it.transactionMutex, transactionId)
			it.engineMutex.Unlock()
		} else {
			return env.ErrorNew("can't find transaction id '" + transactionId + "'")
		}
	} else {
		return env.ErrorNew("transaction id can't be blank")
	}

	return nil
}

// rollbacks modifications made within given transaction
func (it *SQLite) RollbackTransaction(transactionId string) error {
	if transactionId != "" {
		if connection, present := it.transactions[transactionId]; present {
			connection.Exec("ROLLBACK TRANSACTION")

			it.engineMutex.Lock()
			delete(it.transactions, transactionId)
			delete(it.transactionMutex, transactionId)
			it.engineMutex.Unlock()
		} else {
			return env.ErrorNew("can't find transaction id '" + transactionId + "'")
		}
	} else {
		return env.ErrorNew("transaction id can't be blank")
	}

	return nil
}
