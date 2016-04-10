package utils

import (
	"sync"
	"reflect"
	"errors"
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


type pathItem struct {
	parent *pathItem
	mutex  *syncMutex
	locked bool
	key    reflect.Value
	value  reflect.Value
}

func (it *pathItem) Lock() {
	it.locked = true
	it.mutex.Lock()
}

func (it *pathItem) Unlock() {
	for x := it.parent; x != nil; x = x.parent {
		if x.locked {
			x.mutex.Unlock()
			x.locked = false
		}
	}
}

func (it *pathItem) Update(newSubject reflect.Value) {
	stack := make([]*pathItem, 0, 100)
	locked := make([]*syncMutex, 0, 100)

	// collecting pathItem stack and locking items
	for x := it.parent; x != nil; x = x.parent {
		x.mutex.Lock()
		stack = append(stack, x)
		locked = append(locked, x.mutex)
	}

	// updating pathItem references
	for i := len(stack)-1; i > 0; i-- {
		oldValue := stack[i-1].value

		switch stack[i].value.Kind() {
		case reflect.Map:
			stack[i-1].value = stack[i].value.MapIndex(stack[i].key)

		case reflect.Slice, reflect.Array:
			idx := int(stack[i].key.Int())
			stack[i-1].value = stack[i].value.Index(idx)
		}

		if oldValue != stack[i-1].value {
			mutex, err := SyncMutex(stack[i - 1].value)
			if err != nil {
				panic(err)
			}
			mutex.Lock()

			locked[i-1].Unlock()
			locked[i-1] = mutex

			stack[i-1].locked = true
			stack[i-1].mutex = mutex
			stack[i-1].parent = stack[i]
		}
	}

	// updating the item with new key value
	switch it.value.Kind() {
	case reflect.Map:
		it.value.SetMapIndex(it.key, newSubject)

	case reflect.Slice, reflect.Array:
		idx := int(it.key.Int())
		it.value.Index(idx).Set(newSubject)
	}

	// un-locking locked items
	for _, x := range locked {
		x.Unlock()
	}
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

	if value.CanAddr() {
		value = value.Addr()
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

	// debug.PrintStack()
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

	pathItem, err := getPathItem(subject, path, true, nil)
	if err != nil {
		return err
	}

	rSubject := pathItem.value
	if !rSubject.IsValid() {
		return errors.New("invalid subject")
	}

	kind := rSubject.Kind()
	if kind != reflect.Ptr && pathItem.parent != nil {
		pathItem = pathItem.parent
		rSubject = pathItem.value
	}

	rSubject = reflect.Indirect(rSubject)
	rKey := pathItem.key

	//mutex := pathItem.mutex
	//if mutex == nil {
		mutex, err := SyncMutex(rSubject)
		if err != nil {
			return err
		}
	//}

	// new value validation
	rValue := reflect.ValueOf(value)
	rValueType := rValue.Type()

	// allowing to have setter function instead of just value
	funcValue := func(oldValue reflect.Value, valueType reflect.Type) reflect.Value {
		if !oldValue.IsValid() {
			oldValue = reflect.New(valueType).Elem()
		}
		if rValue.Kind() == reflect.Func {
			if rValueType.NumOut() == 1 && rValueType.NumIn() == 1 {
				// oldValueType := oldValue.Type()
				// !rValueType.In(0).AssignableTo(oldValueType) &&
				//!rValueType.Out(0).AssignableTo(oldValueType) {
				return rValue.Call([]reflect.Value{oldValue})[0]
			}
		}
		return rValue
	}

	mutex.Lock()
	switch rSubject.Kind() {
	case reflect.Map:
		oldValue := rSubject.MapIndex(rKey)
		rSubject.SetMapIndex(rKey, funcValue(oldValue, oldValue.Type()))

	case reflect.Slice, reflect.Array:
		idx := int(rKey.Int())
		oldValue := rSubject.Index(idx)
		oldValue.Set(funcValue(oldValue, oldValue.Type()))

	default:
		rSubject.Set(funcValue(rSubject, rSubject.Type()))
	}
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

// initBlankValue makes a new blankValue for a given type
func initBlankValue(valueType reflect.Type) (reflect.Value, error) {
	switch valueType.Kind() {
	case reflect.Map:
		return reflect.MakeMap(valueType), nil
	case reflect.Slice, reflect.Array:
		value := reflect.New(valueType).Elem()
		value.Set(reflect.MakeSlice(valueType, 0, 10))
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

	// preparing result item
	pathItem := &pathItem {
		parent: parent,
		value:  rSubject,
	}

	// checking for end of path, if so we are done
	if len(path) == 0 {
		return pathItem, nil
	}

	// taking mutex for subject
	mutex, err := SyncMutex(rSubject)
	if err != nil {
		return nil, err
	}
	pathItem.mutex = mutex

	// taking first item from path as key item
	rKey := reflect.ValueOf(path[0])
	if !rKey.IsValid() {
		return nil, errors.New("invalid path")
	}
	rKeyType := rKey.Type()
	pathItem.key = rKey

	newPath := path[1:]

	mutex.Lock()
	// getting path item from subject based on it's type
	switch rSubjectKind {

	case reflect.Map:
		// comparing path kay type to subject key type
		if rKeyType != rSubjectType.Key() {
			return nil, errors.New("invalid path item type " +
			rKeyType.String() + " != " + rSubjectType.Key().String())
		}

		// accessing to subject key item
		rSubjectItem := rSubject.MapIndex(rKey)

		// checking if item is not defined, and we should make new value
		if !rSubjectItem.IsValid() && initBlank {
			if rSubjectItem, err = initBlankValue(rSubjectType.Elem()); err != nil {
				mutex.Unlock()
				return nil, err
			}
			rSubject.SetMapIndex(rKey, rSubjectItem)
		}

		mutex.Unlock()
		return getPathItem(rSubjectItem, newPath, initBlank, pathItem)

	case reflect.Slice, reflect.Array:
		// the key should be integer index
		if rKey.Kind() != reflect.Int {
			return nil, errors.New("invalid path item type: " +
			rKey.Kind().String() + " != Int")
		}
		idx := int(rKey.Int())

		// access items - time critical as locked
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

			rSubject = rSubjectItem
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

			// initializing new blank item
			newItem, err := initBlankValue(rSubjectType.Elem())
			if err != nil {
				mutex.Unlock()
				return nil, err
			}

			// checking if capacity allows to increase length
			length := rSubject.Len()
			if rSubject.Cap() < length {
				rSubject.SetLen(length + 1)
				rSubject.Index(length).Set(newItem)
			} else {
				// new slice creation needed
				newSubject := reflect.New(rSubjectType).Elem()
				newSubject.Set(reflect.Append(rSubject, newItem))
				rSubject.Set(newSubject)

				if parent != nil {
					parent.Update(newSubject)
					pathItem.parent = parent
				}
			}

			pathItem.value = rSubject

			pathItem.key = reflect.ValueOf(length)
			rSubject = rSubject.Index(length)
		}

		mutex.Unlock()
		return getPathItem(rSubject, newPath, initBlank, pathItem)

	default:
		mutex.Unlock()
		return nil, errors.New("invalid subject, path can not be applied to: " + rSubjectType.String())
	}

	return pathItem, nil
}