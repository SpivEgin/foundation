package impex

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/app/models"
	"github.com/ottemo/foundation/env"
)

// setups package related API endpoint routines
func setupAPI() error {

	var err error

	err = api.GetRestService().RegisterAPI("impex", "GET", "models", restImpexListModels)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("impex", "GET", "export/:model", restImpexExportModel)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("impex", "POST", "import/:model", restImpexImportModel)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("impex", "POST", "import", restImpexImport)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("impex", "POST", "test/import", restImpexTestImport)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("impex", "POST", "test/mapping", restImpexTestCsvToMap)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	return nil
}

// WEB REST API used to list available models for Impex system
func restImpexListModels(context api.InterfaceApplicationContext) (interface{}, error) {
	var result []string

	for modelName, modelInstance := range models.GetDeclaredModels() {

		if _, present := impexModels[modelName]; present {
			continue
		}

		_, isObject := modelInstance.(models.InterfaceObject)
		_, isStorable := modelInstance.(models.InterfaceStorable)

		if isObject && isStorable {
			result = append(result, modelName)
		}
	}

	for modelName := range impexModels {
		result = append(result, modelName)
	}

	return result, nil
}

// WEB REST API used export specific model data from system
func restImpexExportModel(context api.InterfaceApplicationContext) (interface{}, error) {

	modelName := context.GetRequestArgument("model")

	var records []map[string]interface{}

	if model, present := impexModels[modelName]; present {
		exportIterator := func(item map[string]interface{}) bool {
			records = append(records, item)
			return true
		}
		err := model.Export(exportIterator)
		if err != nil {
			return nil, err
		}
	} else {

		model, err := models.GetModel(modelName)
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		listable, isListable := model.(models.InterfaceListable)
		object, isObject := model.(models.InterfaceObject)

		if isListable && isObject {
			collection := listable.GetCollection()
			if collection == nil {
				return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "12f978f5-3a90-438b-a3f4-475b34a97884", "can't obtain model collection")
			}

			var attributes []string
			for _, attribute := range object.GetAttributesInfo() {
				attributes = append(attributes, attribute.Attribute)
				collection.ListAddExtraAttribute(attribute.Attribute)
			}

			list, err := collection.List()
			if err != nil {
				return nil, err
			}

			for _, item := range list {
				records = append(records, item.Extra)
			}
		}
	}

	// preparing csv writer
	csvWriter := csv.NewWriter(context.GetResponseWriter())
	csvWriter.Comma = ','

	exportFilename := strings.ToLower(modelName) + "_export_" + time.Now().Format(time.RFC3339) + ".csv"

	context.SetResponseContentType("text/csv")
	context.SetResponseSetting("Content-disposition", "attachment;filename="+exportFilename)

	err := MapToCSV(records, csvWriter)
	if err != nil {
		env.ErrorDispatch(err)
	}

	return nil, nil
}

// WEB REST API used import data to system
func restImpexImportModel(context api.InterfaceApplicationContext) (interface{}, error) {

	context.SetResponseContentType("text/plain")

	var commandLine string
	exchangeDict := make(map[string]interface{})

	modelName := context.GetRequestArgument("model")

	if _, present := impexModels[modelName]; present {
		commandLine = "IMPORT " + modelName
	} else {
		commandLine = "UPDATE " + modelName
	}

	filesProcessed := 0
	additionalMessage := ""

	for _, attachedFile := range context.GetRequestFiles() {
		csvReader := csv.NewReader(attachedFile)
		csvReader.Comma = ','

		err := ImportCSVData(commandLine, exchangeDict, csvReader, context.GetResponseWriter(), false)
		if err != nil && additionalMessage == "" {
			env.ErrorDispatch(err)
			additionalMessage += "with errors"
		}

		filesProcessed++
	}

	return []byte(fmt.Sprintf("%d file(s) processed %s", filesProcessed, additionalMessage)), nil
}

// WEB REST API used to test csv file before import
func restImpexTestImport(context api.InterfaceApplicationContext) (interface{}, error) {

	context.SetResponseContentType("text/plain")

	filesProcessed := 0
	additionalMessage := ""
	for _, attachedFile := range context.GetRequestFiles() {
		csvReader := csv.NewReader(attachedFile)
		csvReader.Comma = ','

		err := ImportCSVScript(csvReader, context.GetResponseWriter(), true)
		if err != nil && additionalMessage == "" {
			env.ErrorDispatch(err)
			additionalMessage += "with errors"
		}
		filesProcessed++
	}

	return []byte(fmt.Sprintf("%d file(s) processed %s", filesProcessed, additionalMessage)), nil
}

// WEB REST API used to process csv file script in impex format
func restImpexImport(context api.InterfaceApplicationContext) (interface{}, error) {

	filesProcessed := 0
	additionalMessage := ""
	for _, attachedFile := range context.GetRequestFiles() {
		csvReader := csv.NewReader(attachedFile)
		csvReader.Comma = ','

		err := ImportCSVScript(csvReader, context.GetResponseWriter(), false)
		if err != nil && additionalMessage == "" {
			env.ErrorDispatch(err)
			additionalMessage += "with errors"
		}

		filesProcessed++
	}

	return []byte(fmt.Sprintf("%d file(s) processed %s", filesProcessed, additionalMessage)), nil
}

// WEB REST API to test conversion from csv to json / map[string]interface{}
func restImpexTestCsvToMap(context api.InterfaceApplicationContext) (interface{}, error) {

	var attachedFile io.Reader

	for fileName, requestFile := range context.GetRequestFiles() {
		if strings.HasSuffix(fileName, ".csv") {
			attachedFile = requestFile
		}
	}

	// file was not found
	if attachedFile == nil {
		return nil, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "6f5a582f-46f1-4b46-834d-e6b39da1ca68", "no csv file was attached")
	}

	// preparing csv reader
	var result []map[string]interface{}
	processor := func(item map[string]interface{}) bool {
		result = append(result, item)
		return true
	}

	reader := csv.NewReader(attachedFile)
	reader.Comma = ','

	// making csv processing
	exchangeDict := make(map[string]interface{})
	err := CSVToMap(reader, processor, exchangeDict)

	return result, err
}
