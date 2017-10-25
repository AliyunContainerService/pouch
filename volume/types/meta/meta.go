package meta

import (
	"fmt"
	"reflect"

	"github.com/alibaba/pouch/pkg/serializer"
)

var errNotList = fmt.Errorf("object does not implement the List interfaces")
var errNotObject = fmt.Errorf("object does not implement the object interfaces")

// List represents list interface.
type List interface {
	GetResourceVersion() int64
	SetResourceVersion(version int64)
}

// ListMetaAccessor represents list meta accessor interface.
type ListMetaAccessor interface {
	GetListMeta() List
}

// Object represents meta's object interface.
type Object interface {
	GetUID() string
	SetUID(uid string)
	GetName() string
	SetName(name string)
	GetResourceVersion() int64
	SetResourceVersion(version int64)
}

// ObjectMetaAccessor represents object meta accessor interface.
type ObjectMetaAccessor interface {
	GetObjectMeta() Object
}

// GetListMeta returns ListMeta instance.
func (meta *ListMeta) GetListMeta() List { return meta }

// GetResourceVersion returns ListMeta's resource version.
func (meta *ListMeta) GetResourceVersion() int64 { return meta.ResourceVersion }

// SetResourceVersion is used to set ListMeta's resource version.
func (meta *ListMeta) SetResourceVersion(version int64) { meta.ResourceVersion = version }

// GetUID returns meta's uid.
func (meta *ObjectMeta) GetUID() string { return meta.UID }

// GetName returns meta's name.
func (meta *ObjectMeta) GetName() string { return meta.Name }

// GetResourceVersion returns meta's resource version.
func (meta *ObjectMeta) GetResourceVersion() int64 { return meta.ResourceVersion }

// SetUID is used to set meta's uid.
func (meta *ObjectMeta) SetUID(uid string) { meta.UID = uid }

// SetName is used to set meta's name.
func (meta *ObjectMeta) SetName(name string) { meta.Name = name }

// SetResourceVersion is used to set meta's resource version.
func (meta *ObjectMeta) SetResourceVersion(version int64) { meta.ResourceVersion = version }

// ListAccessor is used to check obj types, is list type or not.
func ListAccessor(obj interface{}) (List, error) {
	switch t := obj.(type) {
	case List:
		return t, nil
	case ListMetaAccessor:
		if m := t.GetListMeta(); m != nil {
			return m, nil
		}
		return nil, errNotList
	case Object:
		return t, nil
	case ObjectMetaAccessor:
		if m := t.GetObjectMeta(); m != nil {
			return m, nil
		}
		return nil, errNotList
	default:
		return nil, errNotList
	}
}

// Accessor is used to check obj types.
func Accessor(obj interface{}) (Object, error) {
	switch t := obj.(type) {
	case Object:
		return t, nil
	case ObjectMetaAccessor:
		if m := t.GetObjectMeta(); m != nil {
			return m, nil
		}
		return nil, errNotObject
	case List, ListMetaAccessor:
		return nil, errNotObject
	default:
		return nil, errNotObject
	}
}

// ListMetaFor is used to change obj to ListMeta and return it.
func ListMetaFor(obj serializer.Object) (*ListMeta, error) {
	v, err := EnforcePtr(obj)
	if err != nil {
		return nil, err
	}
	var meta *ListMeta
	err = FieldPtr(v, "ListMeta", &meta)
	return meta, err
}

// FieldPtr is used to get type that base on filedName from v, return dest type.
func FieldPtr(v reflect.Value, fieldName string, dest interface{}) error {
	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return fmt.Errorf("couldn't find %v field in %#v", fieldName, v.Interface())
	}
	v, err := EnforcePtr(dest)
	if err != nil {
		return err
	}
	if field.Kind() != reflect.Ptr && field.CanAddr() {
		field = field.Addr()
	}
	if field.Type().AssignableTo(v.Type()) {
		v.Set(field)
		return nil
	}
	if field.Type().ConvertibleTo(v.Type()) {
		v.Set(field.Convert(v.Type()))
		return nil
	}
	return fmt.Errorf("couldn't assign/convert %v to %v", field.Type(), v.Type())
}
