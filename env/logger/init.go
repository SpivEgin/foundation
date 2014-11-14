package logger

import (
	"github.com/ottemo/foundation/env"
	"os"
)

// module entry point
func init() {
	instance := new(DefaultLogger)
	var _ env.I_Logger = instance

	env.RegisterLogger(instance)
	env.RegisterOnConfigIniStart(startup)
}

// service pre-initialization stuff
func startup() error {
	if _, err := os.Stat(baseDirectory); !os.IsExist(err) {
		err := os.MkdirAll(baseDirectory, os.ModePerm)
		if err != nil {
			return err
		}
	}

	return nil
}
