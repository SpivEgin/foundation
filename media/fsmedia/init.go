package fsmedia

import (
	"os"

	"github.com/ottemo/foundation/db"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/media"
)

// init makes package self-initialization routine
func init() {
	instance := new(FilesystemMediaStorage)
	if err := media.RegisterMediaStorage(instance); err == nil {
		instance.setupWaitCnt = 2

		env.RegisterOnConfigIniStart(instance.setupOnIniConfig)
		db.RegisterOnDatabaseStart(instance.setupOnDatabase)
	}
}

// setupCheckDone performs callback event if setup was done
func (it *FilesystemMediaStorage) setupCheckDone() {

	// so, we are not sure on events sequence order
	if it.setupWaitCnt--; it.setupWaitCnt == 0 {
		media.OnMediaStorageStart()
	}
}

// setupOnIniConfig is a initialization based on ini config service
func (it *FilesystemMediaStorage) setupOnIniConfig() error {

	var storageFolder = ConstMediaDefaultFolder

	if iniConfig := env.GetIniConfig(); iniConfig != nil {
		if iniValue := iniConfig.GetValue("media.fsmedia.folder", "?"); iniValue != "" {
			storageFolder = iniValue
		}
	}

	err := os.MkdirAll(storageFolder, os.ModePerm)
	if err != nil {
		return err
	}

	it.storageFolder = storageFolder

	if it.storageFolder != "" && it.storageFolder[len(it.storageFolder)-1] != '/' {
		it.storageFolder += "/"
	}

	it.setupCheckDone()

	return nil
}

// setupOnDatabase is a initialization based on config service
func (it *FilesystemMediaStorage) setupOnDatabase() error {

	dbEngine := db.GetDBEngine()
	if dbEngine == nil {
		return env.ErrorNew("Can't get database engine")
	}

	collection, err := dbEngine.GetCollection(ConstMediaDBCollection)
	if err != nil {
		return err
	}

	collection.AddColumn("model", "text", true)
	collection.AddColumn("object", "text", true)
	collection.AddColumn("type", "text", true)
	collection.AddColumn("media", "text", false)

	it.setupCheckDone()

	return nil
}
