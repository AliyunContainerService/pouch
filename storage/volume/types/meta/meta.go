package meta

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
