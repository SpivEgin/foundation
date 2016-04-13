package attributes
import (
	"github.com/ottemo/foundation/app/models"
	"github.com/ottemo/foundation/env"
)

// GenericProduct type implements:
// 	- InterfaceExternalAttributes
// 	- InterfaceObject
// 	- InterfaceStorable

// Init initializes per instance helper before usage
// {instance} is a reference to object which using helper
func ExternalAttributes(instance interface{}) (*ModelExternalAttributes, error) {
	newInstance := &ModelExternalAttributes{instance: instance}

	// getting model name from instance
	modelName := ""
	instanceAsModel, ok := instance.(models.InterfaceModel)
	if !ok || instanceAsModel == nil {
		return nil, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "fe42f2db-2d4b-444a-9891-dc4632ad6dff", "Invalid instance")
	}
	modelName = instanceAsModel.GetModelName()

	if modelName == "" {
		return nil, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "fe42f2db-2d4b-444a-9891-dc4632ad6dff", "Invalid instance")
	}
	newInstance.model = modelName

	// making copy of model attributes and delegates to new instance
	modelExternalAttributesMutex.Lock()
	modelExternalDelegatesMutex.Lock()
	defer modelExternalAttributesMutex.Unlock()
	defer modelExternalDelegatesMutex.Unlock()

	//if externalAttributes, present := modelExternalAttributes[modelName]; !present {
	//	newInstance.attributes = make(map[string]models.StructAttributeInfo)
	//	for attribute, info := range externalAttributes {
	//		newInstance.attributes[attribute] = info
	//	}
	//}

	if delegates, present := modelExternalDelegates[modelName]; !present {
		newInstance.delegates = make(map[string]interface{})
		for attribute, delegate := range delegates {
			if delegate, ok := delegate.(interface{ New(instance interface{}) (interface{}, error) }); ok {
				if delegateInstance, err := delegate.New(instance); err == nil {
					newInstance.delegates[attribute] = delegateInstance
				} else {
					return nil, env.ErrorDispatch(err)
				}
			} else {
				return nil, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "79523f6c-6d1a-4af6-88e8-2f154e21afe7", "Delegate have no New(interface{}) (interface{}, error) method")
			}
		}
	}

	return newInstance, nil
}

// ----------------------------------------------------------------------------------------------
// InterfaceExternalAttributes implementation (package "github.com/ottemo/foundation/app/models")
// ----------------------------------------------------------------------------------------------


// GetCurrentInstance returns current instance delegate attached to
func (it *ModelExternalAttributes) GetInstance() interface{} {
	return it.instance
}

// AddExternalAttribute registers new delegate for a given attribute
func (it *ModelExternalAttributes) AddExternalAttribute(newAttribute models.StructAttributeInfo, delegate interface{}) error {
	modelName := it.model
	attributeName := newAttribute.Attribute

	modelExternalAttributesMutex.Lock()
	modelExternalDelegatesMutex.Lock()
	defer modelExternalAttributesMutex.Unlock()
	defer modelExternalDelegatesMutex.Unlock()

	attributesInfo, present := modelExternalAttributes[modelName]
	if !present {
		attributesInfo = make(map[string]models.StructAttributeInfo)
		modelExternalAttributes[modelName] = attributesInfo
		modelExternalDelegates[modelName] = make(map[string]interface{})
	}

	_, present = attributesInfo[attributeName]
	if !present {
		modelExternalAttributes[modelName][attributeName] = newAttribute
		modelExternalDelegates[modelName][attributeName] = delegate
	} else {
		return env.ErrorNew(ConstErrorModule, ConstErrorLevel, "c2175996-b5f1-40dc-9ce2-9df133c3a2c4", "Attribute already exist")
	}

	// updating instance
	if newInstance, err := ExternalAttributes(it.instance); err == nil {
		it = newInstance
	}

	return nil
}

// RemoveExternalAttribute registers new delegate for a given attribute
func (it *ModelExternalAttributes) RemoveExternalAttribute(attributeName string) error {
	modelName := it.model

	modelExternalAttributesMutex.Lock()
	modelExternalDelegatesMutex.Lock()
	defer modelExternalAttributesMutex.Unlock()
	defer modelExternalDelegatesMutex.Unlock()

	attributesInfo, present := modelExternalAttributes[modelName]
	if !present {
		modelExternalAttributes[modelName] = make(map[string]models.StructAttributeInfo)
	}

	_, present = attributesInfo[attributeName]
	if present {
		delete(attributesInfo, attributeName)
	} else {
		return env.ErrorNew(ConstErrorModule, ConstErrorLevel, "c2175996-b5f1-40dc-9ce2-9df133c3a2c4", "Attribute not exist")
	}

	delegates, present := modelExternalDelegates[modelName]
	if !present {
		modelExternalDelegates[modelName] = make(map[string]interface{})
	}

	_, present = delegates[attributeName]
	if present {
		delete(delegates, attributeName)
	} else {
		return env.ErrorNew(ConstErrorModule, ConstErrorLevel, "c2175996-b5f1-40dc-9ce2-9df133c3a2c4", "Attribute not exist")
	}

	// updating instance
	if newInstance, err := ExternalAttributes(it.instance); err == nil {
		it = newInstance
	}

	return nil
}

// ListExternalAttributes registers new delegate for a given attribute
func (it *ModelExternalAttributes) ListExternalAttributes() []string {
	var result []string
	modelName := it.model

	modelExternalAttributesMutex.Lock()
	defer modelExternalAttributesMutex.Unlock()

	attributesInfo, present := modelExternalAttributes[modelName]
	if !present {
		modelExternalAttributes[modelName] = make(map[string]models.StructAttributeInfo)
	}

	for name := range attributesInfo {
		result = append(result, name)
	}

	return result
}

// ----------------------------------------------------------------------------------
// InterfaceModel implementation (package "github.com/ottemo/foundation/app/models")
// ----------------------------------------------------------------------------------


// GetModelName stub func for Model interface - returns model name
func (it *ModelExternalAttributes) GetModelName() string {
	return it.model
}

// GetImplementationName stub func for Model interface - doing callback to model instance function if possible
func (it *ModelExternalAttributes) GetImplementationName() string {
	if instanceAsModel, ok := it.instance.(models.InterfaceModel); ok {
		return instanceAsModel.GetImplementationName()
	}
	return ""
}

// New makes per-instance initialization routines
func (it *ModelExternalAttributes) New() (models.InterfaceModel, error) {
	modelExternalDelegatesMutex.Lock()
	defer modelExternalDelegatesMutex.Unlock()

	if delegates, present := modelExternalDelegates[it.model]; present {
		for attribute, delegate := range delegates {
			it.delegates[attribute] = delegate
			if delegate, ok := delegate.(interface{ New(instance interface{}) (interface{}, error) }); ok {
				instancedDelegate, err := delegate.New(it.instance)
				if err != nil {
					return it, err
				}
				it.delegates[attribute] = instancedDelegate
			}
		}
	}
	return it, nil
}


// ----------------------------------------------------------------------------------
// InterfaceObject implementation (package "github.com/ottemo/foundation/app/models")
// ----------------------------------------------------------------------------------


// Get returns object attribute value or nil
func (it *ModelExternalAttributes) Get(attribute string) interface{} {
	if delegate, present := it.delegates[attribute]; present {
		if delegate, ok := delegate.(interface{ Get(string) interface{} }); ok {
			return delegate.Get(attribute)
		}
	}

	return nil
}

// Set sets attribute value to object or returns error
func (it *ModelExternalAttributes) Set(attribute string, value interface{}) error {
	if delegate, present := it.delegates[attribute]; present {
		if delegate, ok := delegate.(interface{ Set(string, interface{}) error }); ok {
			return delegate.Set(attribute, value)
		}
	}

	return nil
}

// GetAttributesInfo represents object as map[string]interface{}
func (it *ModelExternalAttributes) GetAttributesInfo() []models.StructAttributeInfo {
	var result []models.StructAttributeInfo

	modelExternalAttributesMutex.Lock()
	defer modelExternalAttributesMutex.Unlock()

	if attributesInfo, present := modelExternalAttributes[it.model]; present {
		for _, info := range attributesInfo {
			result = append(result, info)
		}
	}

	return result
}

// FromHashMap represents object as map[string]interface{}
func (it *ModelExternalAttributes) FromHashMap(input map[string]interface{}) error {
	for attribute, delegate := range it.delegates {
		if value, present := input[attribute]; present {
			if delegate, ok := delegate.(interface{ Set(string, interface{}) error }); ok {
				err := delegate.Set(attribute, value)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// ToHashMap fills object attributes from map[string]interface{}
func (it *ModelExternalAttributes) ToHashMap() map[string]interface{} {
	result := make(map[string]interface{})
	for attribute, delegate := range it.delegates {
		if delegate, ok := delegate.(interface{ Get(string) interface{} }); ok {
			result[attribute] = delegate.Get(attribute)
		}
	}
	return result
}


// ------------------------------------------------------------------------------------
// InterfaceStorable implementation (package "github.com/ottemo/foundation/app/models")
// ------------------------------------------------------------------------------------


// GetID delegates call back to instance (stub method)
func (it *ModelExternalAttributes) GetID() string {
	if instance, ok := it.instance.(interface{ GetID() string }); ok {
		return instance.GetID()
	}
	return ""
}

// SetID callbacks all external attribute delegates
func (it *ModelExternalAttributes) SetID(id string) error {
	for _, delegate := range it.delegates {
		if delegate, ok := delegate.(interface{ SetID(newID string) error }); ok {
			if err := delegate.SetID(id); err != nil {
				return err
			}
		}
	}
	return nil
}

// Load callbacks all external attribute delegates
func (it *ModelExternalAttributes) Load(id string) error {
	for _, delegate := range it.delegates {
		if delegate, ok := delegate.(interface{ Load(loadID string) error }); ok {
			if err := delegate.Load(id); err != nil {
				return err
			}
		}
	}
	return nil
}

// Delete callbacks all external attribute delegates
func (it *ModelExternalAttributes) Delete() error {
	for _, delegate := range it.delegates {
		if delegate, ok := delegate.(interface{ Delete() error }); ok {
			if err := delegate.Delete(); err != nil {
				return err
			}
		}
	}
	return nil
}

// Save callbacks all external attribute delegates
func (it *ModelExternalAttributes) Save() error {
	for _, delegate := range it.delegates {
		if delegate, ok := delegate.(interface{ Save() error }); ok {
			if err := delegate.Save(); err != nil {
				return err
			}
		}
	}
	return nil
}

