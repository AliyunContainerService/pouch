package mgr

import (
	"sync"

	"github.com/alibaba/pouch/pkg/errtypes"
	"github.com/alibaba/pouch/pkg/reference"

	digest "github.com/opencontainers/go-digest"
	pkgerrors "github.com/pkg/errors"
	"github.com/tchap/go-patricia/patricia"
)

// imageStore stores the relationship between references.
//
// Primary reference is the reference used for pulling the image. For example,
//
// 	pouch pull busybox:1.26
//
// the busybox:1.26 is the primary reference. However, the user can use the
// the following references to consume the busybox:1.26:
//
//	busybox:1.26
//	busybox@sha256:be3c11fdba7cfe299214e46edc642e09514dbb9bbefcd0d3836c05a1e0cd0642
//	c30178c5239f
//
// The Name[:Tag][@Digest] are the references generated from busybox:1.26. Therefore,
// We can call it searchable references, which allows users use different reference
// to use the busybox:1.26.
//
// Searchable reference can be named by user. For example, the localhost:5000/busybox:latest
// is the searchable reference for the busybox:1.26. Then the localhost:5000/busybox:latest
// is like alias. You can remove the alias and the image is still here.
//
// However, if we remove the primary reference, bot the image and all searchable
// references will be removed.
//
// NOTE: In the Containerd, the Name to be pulled is used as primary key. We can
// only use the name to search the image in Containerd. Based on this, we should
// make primary and searchable references here.
//
// Besides this, the user can use busybox[:whatever]@digest to pull image. The
// registry doesn't care the tag information if you provide digest. The tag maybe
// wrong. In order to avoid this kind of data, we will trim tag information from
// Name:Tag@Digest reference.
type imageStore struct {
	sync.Mutex

	// idSet builds the imageID set in trie
	idSet *patricia.Trie

	// primaryRefIndexByRef stores primary reference, index by searchable reference
	primaryRefIndexByRef referenceMap

	// idIndexByPrimaryRef stores image ID, index by primary reference
	idIndexByPrimaryRef map[string]digest.Digest

	// refsIndexByPrimaryRef stores searchable references, index by primary reference
	refsIndexByPrimaryRef map[string]referenceMap

	// primaryRefsIndexByID stores primay references, index by image ID
	primaryRefsIndexByID map[digest.Digest]referenceMap
}

// referenceMap represents reference string to corresponding reference.Named
type referenceMap map[string]reference.Named

func newImageStore() (*imageStore, error) {
	return &imageStore{
		idSet: patricia.NewTrie(),

		primaryRefIndexByRef: make(referenceMap),
		idIndexByPrimaryRef:  make(map[string]digest.Digest),

		primaryRefsIndexByID:  make(map[digest.Digest]referenceMap),
		refsIndexByPrimaryRef: make(map[string]referenceMap),
	}, nil
}

// GetReferences returns the list of searchable references by the given image ID.
func (store *imageStore) GetReferences(id digest.Digest) []reference.Named {
	store.Lock()
	defer store.Unlock()

	res := make([]reference.Named, 0)
	for pRefStr := range store.primaryRefsIndexByID[id] {
		for _, ref := range store.refsIndexByPrimaryRef[pRefStr] {
			res = append(res, ref)
		}
	}
	return res
}

// GetPrimaryReferences returns the list of primary references by the given imageID.
func (store *imageStore) GetPrimaryReferences(id digest.Digest) []reference.Named {
	store.Lock()
	defer store.Unlock()

	res := make([]reference.Named, 0, len(store.primaryRefsIndexByID[id]))
	for _, pRef := range store.primaryRefsIndexByID[id] {
		res = append(res, pRef)
	}
	return res
}

// GetPrimaryReference returns the primary reference by the given searchable reference.
func (store *imageStore) GetPrimaryReference(ref reference.Named) (reference.Named, error) {
	trimRef := reference.TrimTagForDigest(ref)

	store.Lock()
	defer store.Unlock()

	if p, ok := store.primaryRefIndexByRef[trimRef.String()]; ok {
		return p, nil
	}
	return nil, errtypes.ErrNotfound
}

// Search returns the imageID, reference.Named by the given reference.Named.
//
// NOTE: the reference.Named in the result maybe different from the given reference.Named
// because it can be added latest tag if missing.
func (store *imageStore) Search(ref reference.Named) (digest.Digest, reference.Named, error) {
	store.Lock()
	defer store.Unlock()

	trimRef := reference.TrimTagForDigest(ref)

	// if the reference is searchable tagged or digest reference
	if p, ok := store.primaryRefIndexByRef[trimRef.String()]; ok {
		return store.idIndexByPrimaryRef[p.String()], ref, nil
	}

	// try to add latest if the reference is only name without tag or digest
	if reference.IsNamedOnly(ref) {
		latestRef := reference.WithDefaultTagIfMissing(ref)
		if p, ok := store.primaryRefIndexByRef[latestRef.String()]; ok {
			return store.idIndexByPrimaryRef[p.String()], latestRef, nil
		}
	}

	// if the reference is short ID or ID
	//
	// NOTE: by default, use the sha256 as the digest algorithm
	id, err := store.searchIDs(digest.Canonical.String(), ref.String())
	if err != nil {
		return "", nil, err
	}
	return id, ref, nil
}

func (store *imageStore) searchIDs(algo string, prefixID string) (digest.Digest, error) {
	var ids []digest.Digest

	fn := func(_ patricia.Prefix, item patricia.Item) error {
		if got, ok := item.(digest.Digest); ok {
			ids = append(ids, got)
		}

		if len(ids) > 1 {
			return pkgerrors.Wrap(errtypes.ErrTooMany, "image: "+prefixID)
		}
		return nil
	}

	if err := store.idSet.VisitSubtree(patricia.Prefix(algo+":"+prefixID), fn); err != nil {
		return "", err
	}

	if len(ids) == 0 {
		return "", pkgerrors.Wrap(errtypes.ErrNotfound, "image: "+prefixID)
	}
	return ids[0], nil
}

// AddReference adds new reference to the imageID.
func (store *imageStore) AddReference(id digest.Digest, primaryRef reference.Named, ref reference.Named) error {
	if reference.IsNamedOnly(ref) ||
		reference.IsNamedOnly(primaryRef) {

		return pkgerrors.Wrap(errtypes.ErrInvalidType, "invalid reference: missing a tag or digest")
	}

	var (
		trimRef        = reference.TrimTagForDigest(ref)
		trimPrimaryRef = reference.TrimTagForDigest(primaryRef)
	)

	store.Lock()
	defer store.Unlock()

	// remove the relationship if the ref has been used by other
	if oldP, ok := store.primaryRefIndexByRef[trimRef.String()]; ok {
		// if the trimRef is primary reference
		if oldP.String() == trimRef.String() {
			if oldP.String() != trimPrimaryRef.String() {

				return pkgerrors.Wrap(errtypes.ErrInvalidType, "invalid reference: cannot replace primary reference")
			}
		}

		delete(store.primaryRefIndexByRef, trimRef.String())
		delete(store.refsIndexByPrimaryRef[oldP.String()], trimRef.String())
	}

	// remove the relationship if the id of primaryRef doesn't equal to original one
	if oldID, ok := store.idIndexByPrimaryRef[trimPrimaryRef.String()]; ok {
		if oldID.String() != id.String() {
			delete(store.primaryRefsIndexByID[oldID], trimPrimaryRef.String())
		}
	}

	store.idSet.Set(patricia.Prefix(id.String()), id)
	store.idIndexByPrimaryRef[trimPrimaryRef.String()] = id
	store.primaryRefIndexByRef[trimRef.String()] = trimPrimaryRef

	// store mapping between primary reference and searchable reference
	if store.refsIndexByPrimaryRef[trimPrimaryRef.String()] == nil {
		store.refsIndexByPrimaryRef[trimPrimaryRef.String()] = make(referenceMap)
	}
	store.refsIndexByPrimaryRef[trimPrimaryRef.String()][trimRef.String()] = trimRef

	if store.primaryRefsIndexByID[id] == nil {
		store.primaryRefsIndexByID[id] = make(map[string]reference.Named)
	}
	store.primaryRefsIndexByID[id][trimPrimaryRef.String()] = trimPrimaryRef
	return nil
}

// RemoveReference removes the reference from imageID.
func (store *imageStore) RemoveReference(id digest.Digest, ref reference.Named) error {
	store.Lock()
	defer store.Unlock()

	if store.idSet.Get(patricia.Prefix(id.String())) == nil {
		return nil
	}

	trimRefStr := reference.TrimTagForDigest(ref).String()

	if p, ok := store.primaryRefIndexByRef[trimRefStr]; ok {
		delete(store.primaryRefIndexByRef, trimRefStr)
		delete(store.refsIndexByPrimaryRef[p.String()], trimRefStr)

		// delete other references if the trimRefStr is primary reference
		if trimRefStr == p.String() {
			for ref := range store.refsIndexByPrimaryRef[p.String()] {
				delete(store.primaryRefIndexByRef, ref)
			}

			delete(store.primaryRefsIndexByID[id], p.String())
			delete(store.refsIndexByPrimaryRef, p.String())

			// if the reference is the final one, we should remove the image
			if len(store.primaryRefsIndexByID[id]) == 0 {
				store.idSet.Delete(patricia.Prefix(id.String()))
			}
		}
	}
	return nil
}

// RemoveAllReferences removes all the reference by the given imageID.
func (store *imageStore) RemoveAllReferences(id digest.Digest) error {
	store.Lock()
	defer store.Unlock()

	for pRefStr := range store.primaryRefsIndexByID[id] {
		for ref := range store.refsIndexByPrimaryRef[pRefStr] {
			delete(store.primaryRefIndexByRef, ref)
		}

		delete(store.refsIndexByPrimaryRef, pRefStr)
		delete(store.idIndexByPrimaryRef, pRefStr)
	}

	delete(store.primaryRefsIndexByID, id)
	store.idSet.Delete(patricia.Prefix(id.String()))
	return nil
}
