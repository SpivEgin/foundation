package sqlite

import (
	"github.com/mxk/go-sqlite/sqlite3"
	"github.com/ottemo/foundation/db"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
	"sync"
)

var (
	dbEngine *SQLite
)

func init() {
	dbEngine = new(SQLite)
	dbEngine.attributeTypes = make(map[string]map[string]string)

	var _ db.I_DBEngine = dbEngine
	var _ db.I_DBCollection = new(SQLiteCollection)

	env.RegisterOnConfigIniStart(dbEngine.Startup)
	db.RegisterDBEngine(dbEngine)
}

func (it *SQLite) Startup() error {

	// reading ini values
	it.uri = env.IniGetValue("db.sqlite3.uri", "ottemo.db")
	it.poolSize = utils.InterfaceToInt(env.IniGetValue("db.sqlite3.poolSize", ""))
	it.maxConnections = utils.InterfaceToInt(env.IniGetValue("db.sqlite3.maxConnectinos", ""))
	it.gcRate = utils.InterfaceToInt(env.IniGetValue("db.sqlite3.gcRate", ""))

	// initializing db engine struct variables
	it.connectionPool = make([]*sqlite3.Conn, 0, it.poolSize)
	it.connectionMutex = make(map[*sqlite3.Conn]*sync.RWMutex)
	it.connectionQueue = make(map[*sqlite3.Conn]int)
	it.statements = make(map[*sqlite3.Stmt]*sqlite3.Conn)
	it.transactions = make(map[string]*sqlite3.Conn)
	it.transactionMutex = make(map[string]*sync.RWMutex)

	if it.poolSize <= 0 {
		it.poolSize = DEFAULT_POOL_SIZE
	}

	if it.maxConnections <= 0 {
		it.maxConnections = DEFAULT_MAX_CONNECTIONS
	}

	if it.maxConnections < it.poolSize {
		it.maxConnections = it.poolSize
	}

	if it.gcRate <= 0 {
		it.gcRate = DEFAULT_POLL_GC_RATE
	}

	// opening connections
	for i := 0; i < it.poolSize; i++ {
		if newConnection, err := sqlite3.Open(it.uri); err == nil {
			it.connectionPool = append(it.connectionPool, newConnection)
			it.connectionMutex[newConnection] = new(sync.RWMutex)
		} else {
			return env.ErrorDispatch(err)
		}
	}

	// making column info table
	SQL := "CREATE TABLE IF NOT EXISTS " + COLLECTION_NAME_COLUMN_INFO + ` (
		_id        INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
		collection VARCHAR(255),
		column     VARCHAR(255),
		type       VARCHAR(255),
		indexed    NUMERIC)`

	err := connectionExec("", SQL)
	if err != nil {
		return sqlError(SQL, err)
	}

	db.OnDatabaseStart()

	return nil
}
