// These tests moved to separate directory to divide connection to sqlite package test
// which interferes with this functionality
package sqlite_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/env/errorbus"
	"github.com/ottemo/foundation/test"
	"github.com/ottemo/foundation/utils"

	"github.com/ottemo/foundation/app/models/product"
	"github.com/ottemo/foundation/app/models/seo"
)

func TestMain(m *testing.M) {
	err := test.StartAppInTestingMode()
	if err != nil {
		fmt.Println("Unable to start app in testing mode:", err)
	}

	os.Exit(m.Run())
}

func TestJoin(t *testing.T) {
	initConfig(t)

	productA := createProductFromJson(t, `{
		"sku": "sku_A",
		"name": "name_A"
	}`)

	seoA := createSeoFromJson(t, `{
		"url": "url_A",
		"rewrite": "`+productA.GetID()+`"
	}`)

	productCollection, err := product.GetProductCollectionModel()
	if err != nil {
		t.Error(err)
	}

	if err = productCollection.ListFilterAdd("_id", "=", productA.GetID()); err != nil {
		t.Error(err)
	}

	productDBCollection := productCollection.GetDBCollection()
	if err = productDBCollection.AddJoinClause("seoJoin", "url_rewrites", []string{"url"}); err != nil {
		t.Error(err)
	}

	if err = productDBCollection.AddJoinConstraintOn("seoJoin", "_id", "rewrite"); err != nil {
		t.Error(err)
	}

	productMaps, err := productDBCollection.Load()
	if err != nil {
		t.Error(err)
	}

	for _, productMap := range productMaps {
		expectedUrl, present := productMap["url_rewrites_url"]
		if !present {
			t.Error("Expected 'url_rewrites_url' field in SELECT result")
		} else if expectedUrl != "url_A" {
			t.Errorf("Expected url '%s' got '%s'", "url_A", expectedUrl)
		}
	}

	deleteSeo(t, seoA)
	deleteProduct(t, productA)
}

func createProductFromJson(t *testing.T, json string) product.InterfaceProduct {
	productData, err := utils.DecodeJSONToStringKeyMap(json)
	if err != nil {
		fmt.Println("json: " + json)
		t.Error(err)
	}

	productModel, err := product.GetProductModel()
	if err != nil || productModel == nil {
		t.Error(err)
	}

	err = productModel.FromHashMap(productData)
	if err != nil {
		t.Error(err)
	}

	err = productModel.Save()
	if err != nil {
		t.Error(err)
	}

	return productModel
}

func deleteProduct(t *testing.T, productModel product.InterfaceProduct) {
	err := productModel.Delete()
	if err != nil {
		t.Error(err)
	}
}

func createSeoFromJson(t *testing.T, json string) seo.InterfaceSEOItem {
	modelData, err := utils.DecodeJSONToStringKeyMap(json)
	if err != nil {
		fmt.Println("json: " + json)
		t.Error(err)
	}

	model, err := seo.GetSEOItemModel()
	if err != nil || model == nil {
		t.Error(err)
	}

	err = model.FromHashMap(modelData)
	if err != nil {
		t.Error(err)
	}

	err = model.Save()
	if err != nil {
		t.Error(err)
	}

	return model
}

func deleteSeo(t *testing.T, model seo.InterfaceSEOItem) {
	err := model.Delete()
	if err != nil {
		t.Error(err)
	}
}

func initConfig(t *testing.T) {
	var config = env.GetConfig()
	if err := config.SetValue(errorbus.ConstConfigPathErrorHideLevel, 0); err != nil {
		t.Error(err)
	}
}
