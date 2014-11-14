// Package contain interfaces to access environment services.
package env

const (
	CONFIG_ITEM_GROUP_TYPE = "group"

	// predefined logging prefixes
	LOG_PREFIX_ERROR   = "ERROR"
	LOG_PREFIX_WARNING = "WARNING"
	LOG_PREFIX_DEBUG   = "DEBUG"
	LOG_PREFIX_INFO    = "INFO"
)

// interface to system event processor
type I_EventBus interface {
	RegisterListener(event string, listener F_EventListener)
	New(event string, eventData map[string]interface{})
}

// interface to system error processor
type I_ErrorBus interface {
	GetErrorLevel(error) int
	GetErrorCode(error) string
	GetErrorMessage(error) string

	RegisterListener(F_ErrorListener)

	Dispatch(error) error
	New(string) error
}

// interface to system logging service
type I_Logger interface {
	Log(storage string, prefix string, message string)

	LogError(err error)
	LogMessage(message string)

	LogToStorage(storage string, message string)
	LogWithPrefix(prefix string, message string)
}

// interface to startup configuration predefined values service
type I_IniConfig interface {
	SetWorkingSection(sectionName string) error
	SetValue(valueName string, value string) error

	GetSectionValue(sectionName string, valueName string, defaultValue string) string
	GetValue(valueName string, defaultValue string) string

	ListSections() []string
	ListItems() []string
	ListSectionItems(sectionName string) []string
}

// interface to configuration values managing service
type I_Config interface {
	RegisterItem(Item T_ConfigItem, Validator F_ConfigValueValidator) error
	UnregisterItem(Path string) error

	ListPathes() []string
	GetValue(Path string) interface{}
	SetValue(Path string, Value interface{}) error

	GetGroupItems() []T_ConfigItem
	GetItemsInfo(Path string) []T_ConfigItem

	Load() error
	Reload() error
}

// interface to errors generated by I_ErrorBus service
type I_OttemoError interface {
	ErrorFull() string
	ErrorLevel() int
	ErrorCode() string
	ErrorStack() string

	error
}

// configuration value validator callback function interface
type F_ConfigValueValidator func(interface{}) (interface{}, error)

// event listener callback function interface
type F_EventListener func(string, map[string]interface{}) bool

// error listener callback function interface
type F_ErrorListener func(error) bool

// structure to hold information about configuration value
type T_ConfigItem struct {
	Path  string
	Value interface{}

	Type string

	Editor  string
	Options interface{}

	Label       string
	Description string

	Image string
}
