package sqlite

import (
	"regexp"
	"sync"

	"github.com/mxk/go-sqlite/sqlite3"
)

const (
	UUID_ID   = true
	DEBUG_SQL = true

	DEFAULT_POOL_SIZE       = 10
	DEFAULT_POLL_GC_RATE    = 10
	DEFAULT_MAX_CONNECTIONS = 100

	FILTER_GROUP_STATIC  = "static"
	FILTER_GROUP_DEFAULT = "default"

	COLLECTION_NAME_COLUMN_INFO = "collection_column_info"
)

var SQL_NAME_VALIDATOR = regexp.MustCompile("^[A-Za-z_][A-Za-z0-9_]*$")

type T_DBFilterGroup struct {
	Name         string
	FilterValues []string
	ParentGroup  string
	OrSequence   bool
}

type SQLiteCollection struct {
	Name string

	ResultColumns []string
	FilterGroups  map[string]*T_DBFilterGroup
	Order         []string

	Limit string

	TransactionId string
}

type SQLite struct {
	uri string

	// cached collection attributes array
	attributeTypes      map[string]map[string]string
	attributeTypesMutex sync.RWMutex

	// connection pool variables
	connectionPool  []*sqlite3.Conn
	connectionMutex map[*sqlite3.Conn]*sync.RWMutex
	connectionQueue map[*sqlite3.Conn]int

	// binding statement to connection
	statements map[*sqlite3.Stmt]*sqlite3.Conn

	// transaction binds to one single connection
	// and holds connection until finish
	// so queries within transaction needs "sub-mutex"
	// to synchronize usage of that connection
	transactions     map[string]*sqlite3.Conn
	transactionMutex map[string]*sync.RWMutex

	// connection allocate limits
	maxConnections int
	poolSize       int
	gcRate         int

	// to synchronize write access struct variables
	engineMutex sync.RWMutex
}
