package utils

import (
	"sync"
	"reflect"
	"errors"
	"runtime/debug"
)

var (
	locks = make(map[uintptr]*syncMutex)
	locksMutex sync.Mutex
)

type syncMutex struct {
	index uintptr
	refs  int
	lock  bool
	mutex sync.Mutex
}

func (it *syncMutex) Lock() {
	it.mutex.Lock()
	it.lock = true
}

func (it *syncMutex) Unlock() {
	locksMutex.Lock()
	defer locksMutex.Unlock()

	it.lock = false
	it.refs--
	if it.refs == 0 {
		delete(locks, it.index)
	}

	it.mutex.Unlock()
}

func (it *syncMutex) GetIndex() uintptr {
	return it.index
}

func (it *syncMutex) IsLocked() bool {
	return it.lock
}

func (it *syncMutex) Refs() int {
	return it.refs
}

func GetPointer(subject interface{}) (uintptr, error) {
	if subject == nil {
		return 0, errors.New("can't get pointer to nil")
	}

	var value reflect.Value

	if rValue, ok := subject.(reflect.Value); ok {
		value = rValue
	} else {
		value = reflect.ValueOf(subject)
	}

	switch value.Kind() {
	case reflect.Chan,
		reflect.Map,
		reflect.Ptr,
		reflect.UnsafePointer,
		reflect.Func,
		reflect.Slice,
		reflect.Array:

		return value.Pointer(), nil
	}

	debug.PrintStack()
	return 0, errors.New("can't get pointer to " + value.Type().String())
}

func SyncMutex(subject interface{}) (*syncMutex, error) {
	locksMutex.Lock()
	defer locksMutex.Unlock()

	index, err := GetPointer(subject)
	if err != nil {
		return nil, err
	}
	if index == 0 {
		return nil, errors.New("mutex to zero pointer")
	}
	mutex, present := locks[index]
	if !present {
		mutex = new(syncMutex)
		mutex.index = index
		locks[index] = mutex
	}
	mutex.refs++
	return mutex, nil
}

func SyncSet(subject interface{}, value interface{}, path ...interface{}) error {

	result, err := getPathItem(subject, path, true, nil)
	if err != nil {
		return err
	}

	rSubject := result.value
	if !rSubject.IsValid() {
		return errors.New("invalid subject")
	}
	mutex := result.mutex

	// new value validation
	rValue := reflect.ValueOf(value)
	rValueType := rValue.Type()

	if !rValueType.AssignableTo(rSubject.Type()) {
		return errors.New(rValueType.String() + " is unassignable to " +  rSubject.Type().String())
	}


	// allowing to have setter function instead of just value
	funcValue := func(oldValue reflect.Value, valueType reflect.Type) reflect.Value {
		if !oldValue.IsValid() {
			oldValue = reflect.New(valueType).Elem()
		}
		if rValue.Kind() == reflect.Func {
			if rValueType.NumOut()==1 && rValueType.NumIn()==1 {
				// oldValueType := oldValue.Type()
				// !rValueType.In(0).AssignableTo(oldValueType) &&
				//!rValueType.Out(0).AssignableTo(oldValueType) {
				return rValue.Call([]reflect.Value{oldValue})[0]
			}
		}
		return rValue
	}

	mutex.Lock()
	rSubject.Set( funcValue(rSubject, rSubject.Type()) )
	mutex.Unlock()

	return nil
}

func SyncGet(subject interface{}, initBlank bool, path ...interface{}) (interface{}, error) {
	result, err := getPathItem(subject, path, initBlank, nil)
	if err != nil {
		return nil, err
	}
	return result.value.Interface(), nil
}

type pathItem struct {
	parent *pathItem
	mutex  *syncMutex
	key    reflect.Value
	value  reflect.Value
}

// initBlankValue makes a new blankValue for a given type
func initBlankValue(valueType reflect.Type) (reflect.Value, error) {
	switch valueType.Kind() {
	case reflect.Map:
		return reflect.MakeMap(valueType), nil
	case reflect.Slice, reflect.Array:
		value := reflect.New(valueType).Elem()
		value.Set(reflect.MakeSlice(valueType, 0, 0))
		return value, nil
	case reflect.Chan, reflect.Func:
		break
	default:
		return reflect.New(valueType).Elem(), nil
	}
	return reflect.ValueOf(nil), errors.New("unsuported blank value type " + valueType.String())
}

func getPathItem(subject interface{}, path []interface{}, initBlank bool, parent *pathItem) (*pathItem, error) {

	// do nothing for nil objects
	if subject == nil {
		return nil, errors.New("nil subject")
	}

	// checking for reflect.Value in subject
	rSubject, ok := subject.(reflect.Value)
	if !ok {
		rSubject = reflect.ValueOf(subject)
	}

	// handling pointers
	rSubject = reflect.Indirect(rSubject)

	// checking subject for a zero value
	if !rSubject.IsValid() {
		return nil, errors.New("invalid subject")
	}

	// taking subject type and kind
	rSubjectKind := rSubject.Kind()
	rSubjectType := rSubject.Type()

	// taking mutex for subject
	mutex, err := SyncMutex(rSubject)
	if err != nil {
		return nil, err
	}

	// preparing result item
	result := &pathItem {
		parent: parent,
		value:  rSubject,
		mutex:  mutex,
	}

	// checking for end of path, if so we are done
	if len(path) == 0 {
		return result, nil
	}
	newPath := path[1:]

	// taking first item from path as key item
	rKey := reflect.ValueOf(path[0])
	if !rKey.IsValid() {
		return nil, errors.New("invalid path")
	}
	rKeyType := rKey.Type()
	result.key = rKey

	// getting path item from subject based on it's type
	switch rSubjectKind {

	case reflect.Map:
		// comparing path kay type to subject key type
		if rKeyType != rSubjectType.Key() {
			return result, errors.New("invalid path item type " +
			rKeyType.String() + " != " + rSubjectType.Key().String())
		}

		// accessing to subject key item
		mutex.Lock()
		rSubjectItem := rSubject.MapIndex(rKey)

		// checking if item is not defined, and we should make new value
		if !rSubjectItem.IsValid() && initBlank {
			if rSubjectItem, err = initBlankValue(rSubjectType.Elem()); err != nil {
				defer mutex.Unlock()
				return nil, err
			}
			rSubject.SetMapIndex(rKey, rSubjectItem)
		}
		mutex.Unlock()

		return getPathItem(rSubjectItem, newPath, initBlank, result)

	case reflect.Slice, reflect.Array:
		// the key should be integer index
		if rKey.Kind() != reflect.Int {
			return result, errors.New("invalid path item type: " +
			rKey.Kind().String() + " != Int")
		}
		idx := int(rKey.Int())


		// access items - time critical as locked
		mutex.Lock()
		if rSubject.Len() <= idx {
			mutex.Unlock()
			return nil, errors.New("index " + rKey.String() + " is out of bound")
		}

		// (idx = -1) is used to create new item, otherwise it is existing item
		if idx >= 0 {
			rSubjectItem := rSubject.Index(idx)

			// checking if existing item is nil but we should make it
			if !rSubjectItem.IsValid() && initBlank {
				rSubjectValue, err := initBlankValue(rSubjectType.Elem())
				if err != nil {
					mutex.Unlock()
					return nil, err
				}
				rSubjectItem.Set(rSubjectValue)
			}

		} else {
			// checking if new item creation was specified, and we can create it
			if (!initBlank) {
				mutex.Unlock()
				return nil, errors.New("invalid index -1 as initBlank = false")
			}

			if !rSubject.CanAddr() {
				mutex.Unlock()
				return nil, errors.New("not addresable subject")
			}

			// initializing new item
			newItemValue, err := initBlankValue(rSubjectType.Elem())
			if err != nil {
				mutex.Unlock()
				return nil, err
			}

			// checking if capacity allows to increase length
			length := rSubject.Len()
			if rSubject.Cap() < length {
				rSubject.SetLen(length + 1)
				rSubject.Index(length).Set(newItemValue)
			} else {
				// if parent != nil {
				// 	parent.mutex.Lock()
				// }

				// new slice creation needed
				value := reflect.New(rSubjectType).Elem()
				value.Set(reflect.Append(rSubject, newItemValue))
				rSubject.Set(value)

				if parent != nil {
					switch parent.value.Kind() {
					case reflect.Map:
						parent.value.SetMapIndex(parent.key, rSubject)
					case reflect.Slice, reflect.Array:
						idx := int(parent.key.Int())
						if idx < 0 {
							idx = parent.value.Len()-1
						}
						parent.value.Set(rSubject)
					}
					// parent.mutex.Unlock()
				}
			}
		}
		mutex.Unlock()

		return getPathItem(rSubject, newPath, initBlank, result)

	default:
		return nil, errors.New("invalid subject, path can not be applied to: " + rSubjectType.String())
	}

	return result, nil
}