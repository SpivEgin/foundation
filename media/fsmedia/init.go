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
		instance.imageSizes = make(map[string]string)
		instance.setupWaitCnt = 3

		env.RegisterOnConfigIniStart(instance.setupOnIniConfigStart)
		env.RegisterOnConfigStart(instance.setupConfig)
		db.RegisterOnDatabaseStart(instance.setupOnDatabaseStart)
	}
}

// setupCheckDone performs callback event if setup was done
func (it *FilesystemMediaStorage) setupCheckDone() {

	// so, we are not sure on events sequence order
	if it.setupWaitCnt--; it.setupWaitCnt == 0 {
		media.OnMediaStorageStart()
	}
}

// setupOnIniConfigStart is a initialization based on ini config service
func (it *FilesystemMediaStorage) setupOnIniConfigStart() error {

	var storageFolder = ConstMediaDefaultFolder

	if iniConfig := env.GetIniConfig(); iniConfig != nil {
		if iniValue := iniConfig.GetValue("media.fsmedia.folder", "?"+ConstMediaDefaultFolder); iniValue != "" {
			storageFolder = iniValue
		}
	}

	err := os.MkdirAll(storageFolder, os.ModePerm)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	it.storageFolder = storageFolder

	if it.storageFolder != "" && it.storageFolder[len(it.storageFolder)-1] != '/' {
		it.storageFolder += "/"
	}

	it.setupCheckDone()

	return nil
}

// setupOnDatabaseStart is a initialization based on config service
func (it *FilesystemMediaStorage) setupOnDatabaseStart() error {

	dbEngine := db.GetDBEngine()
	if dbEngine == nil {
		return env.ErrorNew("Can't get database engine")
	}

	dbCollection, err := dbEngine.GetCollection(ConstMediaDBCollection)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	dbCollection.AddColumn("model", "text", true)
	dbCollection.AddColumn("object", "text", true)
	dbCollection.AddColumn("type", "text", true)
	dbCollection.AddColumn("media", "text", false)

	it.setupCheckDone()

	return nil
}
