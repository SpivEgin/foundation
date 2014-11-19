package config

import (
	"sort"
	"strings"

	"github.com/ottemo/foundation/db"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
)

// enumerates registered pathes for config
func (it *DefaultConfig) ListPathes() []string {
	result := make([]string, 0)
	for key, _ := range it.configValues {
		result = append(result, key)
	}
	sort.Strings(result)

	return result
}

// registers new config value in system
func (it *DefaultConfig) RegisterItem(Item env.StructConfigItem, Validator env.FuncConfigValueValidator) error {

	if _, present := it.configValues[Item.Path]; !present {

		collection, err := db.GetCollection(ConstConfigCollectionName)
		if err != nil {
			return env.ErrorDispatch(err)
		}

		recordValues := make(map[string]interface{})

		recordValues["path"] = Item.Path
		recordValues["value"] = Item.Value
		recordValues["type"] = Item.Type
		recordValues["editor"] = Item.Editor
		recordValues["options"] = Item.Options
		recordValues["label"] = Item.Label
		recordValues["description"] = Item.Description

		_, err = collection.Save(recordValues)
		if err != nil {
			return env.ErrorDispatch(err)
		}

		it.configValues[Item.Path] = Item.Value
		it.configTypes[Item.Path] = Item.Type
	}

	if _, present := it.configValidators[Item.Path]; Validator != nil && !present {
		it.configValidators[Item.Path] = Validator
	}

	return nil
}

// removes config value from system
func (it *DefaultConfig) UnregisterItem(Path string) error {

	if _, present := it.configValues[Path]; present {

		collection, err := db.GetCollection(ConstConfigCollectionName)
		if err != nil {
			return env.ErrorDispatch(err)
		}

		err = collection.AddFilter("path", "LIKE", Path+"%")
		if err != nil {
			return env.ErrorDispatch(err)
		}

		_, err = collection.Delete()
		if err != nil {
			return env.ErrorDispatch(err)
		}

		return it.Reload()
	}

	return nil
}

// returns value for config item of nil if not present
func (it *DefaultConfig) GetValue(Path string) interface{} {
	if value, present := it.configValues[Path]; present {
		return value
	} else {
		return nil
	}
}

// updates config item with new value, returns error if not possible
func (it *DefaultConfig) SetValue(Path string, Value interface{}) error {
	if _, present := it.configValues[Path]; present {

		// updating value on GO side
		//--------------------------
		if validator, present := it.configValidators[Path]; present {

			if newVal, err := validator(Value); err != nil {
				return env.ErrorDispatch(err)
			} else {
				it.configValues[Path] = newVal
			}

		} else {
			it.configValues[Path] = Value
		}

		// updating value in DB
		//---------------------
		collection, err := db.GetCollection(ConstConfigCollectionName)
		if err != nil {
			return env.ErrorDispatch(err)
		}

		err = collection.AddFilter("path", "=", Path)
		if err != nil {
			return env.ErrorDispatch(err)
		}

		records, err := collection.Load()
		if err != nil {
			return env.ErrorDispatch(err)
		}

		if len(records) == 0 {
			return env.ErrorNew("config item '" + Path + "' is not registered")
		}

		record := records[0]

		record["value"] = it.configValues[Path]

		_, err = collection.Save(record)
		if err != nil {
			return env.ErrorDispatch(err)
		}

	} else {
		return env.ErrorNew("can not find config item '" + Path + "' ")
	}

	return nil
}

// returns information about config items with type [ConstConfigItemGroupType]
func (it *DefaultConfig) GetGroupItems() []env.StructConfigItem {

	result := make([]env.StructConfigItem, 0)

	collection, err := db.GetCollection(ConstConfigCollectionName)
	if err != nil {
		return result
	}

	err = collection.AddFilter("type", "=", env.ConstConfigItemGroupType)
	if err != nil {
		return result
	}

	records, err := collection.Load()
	if err != nil {
		return result
	}

	for _, record := range records {

		configItem := env.StructConfigItem{
			Path:  utils.InterfaceToString(record["path"]),
			Value: record["value"],

			Type: utils.InterfaceToString(record["type"]),

			Editor:  utils.InterfaceToString(record["editor"]),
			Options: record["options"],

			Label:       utils.InterfaceToString(record["label"]),
			Description: utils.InterfaceToString(record["description"]),

			Image: utils.InterfaceToString(record["image"]),
		}
		configItem.Value = db.ConvertTypeFromDbToGo(configItem.Value, configItem.Type)

		result = append(result, configItem)
	}

	return result
}

// returns information about config items with given path
// 	- use '*' to list sub-items (like "paypal.*" or "paypal*" if group item also needed)
func (it *DefaultConfig) GetItemsInfo(Path string) []env.StructConfigItem {
	result := make([]env.StructConfigItem, 0)

	collection, err := db.GetCollection(ConstConfigCollectionName)
	if err != nil {
		return result
	}

	err = collection.AddFilter("path", "LIKE", strings.Replace(Path, "*", "%", -1))
	if err != nil {
		return result
	}

	records, err := collection.Load()
	if err != nil {
		return result
	}

	for _, record := range records {

		configItem := env.StructConfigItem{
			Path:  utils.InterfaceToString(record["path"]),
			Value: record["value"],

			Type: utils.InterfaceToString(record["type"]),

			Editor:  utils.InterfaceToString(record["editor"]),
			Options: record["options"],

			Label:       utils.InterfaceToString(record["label"]),
			Description: utils.InterfaceToString(record["description"]),

			Image: utils.InterfaceToString(record["image"]),
		}
		configItem.Value = db.ConvertTypeFromDbToGo(configItem.Value, configItem.Type)

		result = append(result, configItem)
	}

	return result
}

// loads config data from DB on app startup
//   - calls env.OnConfigStart() after
func (it *DefaultConfig) Load() error {

	err := it.Reload()
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = env.OnConfigStart()
	if err != nil {
		return env.ErrorDispatch(err)
	}

	return nil
}

// updates all config values from database
func (it *DefaultConfig) Reload() error {
	it.configValues = make(map[string]interface{})
	it.configTypes = make(map[string]string)

	collection, err := db.GetCollection(ConstConfigCollectionName)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = collection.SetResultColumns("path", "type", "value")
	if err != nil {
		return env.ErrorDispatch(err)
	}

	records, err := collection.Load()
	if err != nil {
		return env.ErrorDispatch(err)
	}

	for _, record := range records {
		valuePath := utils.InterfaceToString(record["path"])
		valueType := utils.InterfaceToString(record["type"])

		it.configValues[valuePath] = db.ConvertTypeFromDbToGo(record["value"], valueType)
		it.configTypes[valuePath] = valueType
	}

	return nil
}
