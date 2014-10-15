package sqlite

import (
	"sync"
	"github.com/mxk/go-sqlite/sqlite3"
	"github.com/ottemo/foundation/db"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
)

var (
	dbEngine *SQLite
)

func init() {
	dbEngine = new(SQLite)
	dbEngine.attributeTypes = make(map[string]map[string]string)

	var _ db.I_DBEngine = dbEngine

	env.RegisterOnConfigIniStart(dbEngine.Startup)
	db.RegisterDBEngine(dbEngine)
}

func (it *SQLite) Startup() error {

	// opening connection
	it.uri := env.IniGetValue("db.sqlite3.uri", "ottemo.db")

	it.poolSize       := utils.InterfaceToInt( env.IniGetValue("db.sqlite3.poolSize", "1") )
	it.maxConnections := utils.InterfaceToInt( env.IniGetValue("db.sqlite3.maxConnectinos", "1") )

	it.connectionPool  = make([]*sqlite3.Conn, 0, it.poolSize)
	it.connectionMutex = make([]sync.RWMutex,  0, it.poolSize)

	for i:=0; i<it.poolSize; i++ {
		if newConnection, err := sqlite3.Open(it.uri); err == nil {
			it.connectionPool = append(it.connectionPool, newConnection)
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

	err := it.connection.Exec(SQL)
	if err != nil {
		return sqlError(SQL, err)
	}

	db.OnDatabaseStart()

	return nil
}
