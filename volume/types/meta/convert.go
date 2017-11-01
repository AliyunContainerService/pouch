package meta

import (
	"fmt"
	"reflect"

	"github.com/alibaba/pouch/pkg/serializer"
)

// EnforcePtr ensures that obj is a pointer of some sort. Returns a reflect.Value
// of the dereferenced pointer, ensuring that it is settable/addressable.
// Returns an error if this is not possible.
func EnforcePtr(obj interface{}) (reflect.Value, error) {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr {
		if v.Kind() == reflect.Invalid {
			return reflect.Value{}, fmt.Errorf("expected pointer, but got invalid kind")
		}
		return reflect.Value{}, fmt.Errorf("expected pointer, but got %v type", v.Type())
	}
	if v.IsNil() {
		return reflect.Value{}, fmt.Errorf("expected pointer, but got nil")
	}
	return v.Elem(), nil
}

// GetItemsPtr returns a pointer to the list object's Items member.
// If 'list' doesn't have an Items member, it's not really a list type
// and an error will be returned.
// This function will either return a pointer to a slice, or an error, but not both.
func GetItemsPtr(list serializer.Object) (interface{}, error) {
	v, err := EnforcePtr(list)
	if err != nil {
		return nil, err
	}
	items := v.FieldByName("Items")
	if !items.IsValid() {
		return nil, fmt.Errorf("no Items field in %#v", list)
	}
	switch items.Kind() {
	case reflect.Interface, reflect.Ptr:
		target := reflect.TypeOf(items.Interface()).Elem()
		if target.Kind() != reflect.Slice {
			return nil, fmt.Errorf("items: Expected slice, got %s", target.Kind())
		}
		return items.Interface(), nil
	case reflect.Slice:
		return items.Addr().Interface(), nil
	default:
		return nil, fmt.Errorf("items: Expected slice, got %s", items.Kind())
	}
}

// ExtractList returns obj's Items element as an array of serializer.Objects.
// Returns an error if obj is not a List type (does not have an Items member).
func ExtractList(obj serializer.Object) ([]interface{}, error) {
	itemsPtr, err := GetItemsPtr(obj)
	if err != nil {
		return nil, err
	}
	items, err := EnforcePtr(itemsPtr)
	if err != nil {
		return nil, err
	}
	list := make([]interface{}, items.Len())
	for i := range list {
		raw := items.Index(i)
		switch raw.Interface().(type) {
		//case serializer.Object:
		//	list[i] = item
		default:
			var found bool
			if list[i], found = raw.Addr().Interface().(serializer.Object); !found {
				return nil, fmt.Errorf("%v: item[%v]: Expected object, got %#v(%s)", obj, i, raw.Interface(), raw.Kind())
			}
		}
	}
	return list, nil
}
