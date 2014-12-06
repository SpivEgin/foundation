// Package errorbus is a default implementation of InterfaceErrorBus declared in
// "github.com/ottemo/foundation/env" package
package errorbus

import (
	"github.com/ottemo/foundation/env"
	"regexp"
)

// Package global constants
const (
	ConstCollectCallStack = true // flag to indicate that call stack information within error is required

	ConstConfigPatError             = "general.error"
	ConstConfigPathErrorLogLevel    = "general.error.log_level"
	ConstConfigPathErrorHideLevel   = "general.error.hide_level"
	ConstConfigPathErrorHideMessage = "general.error.hide_message"
)

// Package global variables
var (
	// ConstMsgRegexp is a regular expression used to parse error message
	ConstMsgRegexp = regexp.MustCompile(`^[\[{(]?\s*(?:(?:([^:]+)?[: ])?([0-9]+)?[-: ]([0-9a-fA-F]+)?)?\s*[\]})]?\s*[:\->]*\s*(.+)`)

	logLevel    = 5
	hideLevel   = 5
	hideMessage = "System error has occured"
)

// DefaultErrorBus InterfaceErrorBus implementer class
type DefaultErrorBus struct {
	listeners []env.FuncErrorListener
}

// OttemoError @reconcile@ InterfaceOttemoError implementer class
type OttemoError struct {
	Message string
	Module  string
	Code    string
	Level   int

	CallStack string

	handled bool
}
