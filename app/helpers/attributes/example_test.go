package attributes

import (
	"github.com/ottemo/foundation/app/models"
	"github.com/ottemo/foundation/utils"
	"testing"
	"errors"
	"fmt"
)

// -----------------------
// SampleModel declaration
// -----------------------

type SampleModel struct {
	i string
	l string
	d bool
	s bool

	*ModelExternalAttributes
}

func (it *SampleModel) GetModelName() string {
	return "Example"
}

func (it *SampleModel) GetImplementationName() string {
	return "ExampleObject"
}

func (it *SampleModel) New() (models.InterfaceModel, error) {
	var err error

	newInstance := new(SampleModel)
	newInstance.ModelExternalAttributes, err = ExternalAttributes(newInstance)

	return newInstance, err
}

// --------------------------
// SampleDelegate declaration
// --------------------------

type SampleDelegate struct {
	instance interface{}
	a string
	b float64
}

func (it *SampleDelegate) GetPriority() float64 {
	return 1
}

func (it *SampleDelegate) New(instance interface{}) (models.InterfaceAttributesDelegate, error) {
	return &SampleDelegate{instance: instance}, nil
}

func (it *SampleDelegate) Get(attribute string) interface{} {
	switch attribute {
	case "a":
		return it.a
	case "b":
		return it.b
	}
	return nil
}

func (it *SampleDelegate) Set(attribute string, value interface{}) error {
	switch attribute {
	case "a":
		it.a = utils.InterfaceToString(value)
	case "b":
		it.b = utils.InterfaceToFloat64(value)
	}
	return nil
}

func (it *SampleDelegate) Load(id string) error {
	it.instance.(*SampleModel).l = id
	return nil
}

func (it *SampleDelegate) Delete() error {
	it.instance.(*SampleModel).d = true
	return nil
}

func (it *SampleDelegate) Save() error {
	it.instance.(*SampleModel).s = true
	return nil
}

func (it *SampleDelegate) SetID(newID string) error {
	it.instance.(*SampleModel).i = newID
	return nil
}

func (it *SampleDelegate) GetAttributesInfo() []models.StructAttributeInfo {
	return []models.StructAttributeInfo{
		models.StructAttributeInfo{
			Model:      "",
			Collection: "",
			Attribute:  "a",
			Type:       utils.ConstDataTypeText,
			Label:      "A",
			IsRequired: false,
			IsStatic:   false,
			Group:      "Sample",
			Editors:    "text",
		},
		models.StructAttributeInfo{
			Model:      "Example",
			Collection: "",
			Attribute:  "b",
			Type:       utils.ConstDataTypeFloat,
			Label:      "B",
			IsRequired: false,
			IsStatic:   false,
			Group:      "Sample",
			Editors:    "text",
		},
	}
}


// -----
// Tests
// -----
func TestLock(t *testing.T) {
	if err := ExampleExternalAttributes(); err != nil {
		t.Error(err)
	}
}

func ExampleExternalAttributes() error {
	// registering SampleDelegate for SampleModel on attributes "a" and "b"
	modelInstance, err := new(SampleModel).New()
	if err != nil {
		return err
	}

	modelEA, ok := modelInstance.(models.InterfaceExternalAttributes)
	if !ok {
		return errors.New("InterfaceExternalAttributes not impelemented")
	}

	delegate := new(SampleDelegate)
	modelEA.AddExternalAttributes(delegate)

	// testing result: setting "a", "b" attributes to SampleModel instances and getting them back
	var obj1, obj2 models.InterfaceObject
	if x, err := modelInstance.New(); err == nil {
		if obj1, ok = x.(models.InterfaceObject); !ok {
			return errors.New("InterfaceObject not impelemented")
		}
	} else {
		return err
	}

	if x, err := modelInstance.New(); err == nil {
		if obj2, ok = x.(models.InterfaceObject); !ok {
			return errors.New("InterfaceObject not impelemented")
		}
	} else {
		return err
	}


	if err = obj1.Set("a", "object1"); err != nil {
		return err
	}
	if err = obj2.Set("a", "object2"); err != nil {
		return err
	}
	if err = obj1.Set("b", 1.2); err != nil {
		return err
	}
	if err = obj2.Set("b", 3.3); err != nil {
		return err
	}

	if obj1.Get("a") != "object1" || obj1.Get("b") != 1.2 ||
	   obj2.Get("a") != "object2" || obj2.Get("b") != 3.3 {
		return errors.New(fmt.Sprint("incorrect get values: " +
			"obj1.a=", obj1.Get("a"), ", ",
			"obj1.b=", obj1.Get("b"), ", ",
			"obj2.a=", obj2.Get("a"), ", ",
			"obj2.b=", obj2.Get("b"),
		))
	}

	if obj1, ok := obj1.(models.InterfaceStorable); ok {
		obj1.Load("1")
		obj1.Save()
		obj1.Delete()
		obj1.SetID("10")
	} else {
		return errors.New("models.InterfaceStorable not implemented")
	}

	if obj1, ok := obj1.(*SampleModel); ok {
		if !obj1.d || !obj1.s || obj1.l != "1" || obj1.i != "10" {
			return errors.New(fmt.Sprint("incorrect get values: " +
				"obj1.l=", obj1.l, ", ",
				"obj1.s=", obj1.s, ", ",
				"obj1.d=", obj1.d, ", ",
				"obj1.i=", obj1.i,
			))
		}
	} else {
		return errors.New("(*SampleModel) conversion error")
	}

	return nil
}

