package mgr

import (
	"strings"
	"testing"

	"github.com/alibaba/pouch/pkg/errtypes"
	"github.com/alibaba/pouch/pkg/reference"

	digest "github.com/opencontainers/go-digest"
	pkgerrors "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestGetAllReferences(t *testing.T) {
	store, err := newImageStore()
	if err != nil {
		t.Fatalf("unexpected error during creating store: %v", err)
	}

	var (
		id = digest.Digest("sha256:dc5f67a48da730d67bf4bfb8824ea8a51be26711de090d6d5a1ffff2723168a1")

		primaryRefStr = "busybox:latest"
		otherRefStrs  = map[string]struct{}{
			"ref:1": {},
			"ref:2": {},
			"ref:3": {},
		}

		primaryRefNamed reference.Named
	)

	// prepare reference
	primaryRefNamed, err = reference.Parse(primaryRefStr)
	if err != nil {
		t.Fatalf("unexpected error during parsing reference %s: %v", primaryRefStr, err)
	}

	for refStr := range otherRefStrs {
		refNamed, err := reference.Parse(refStr)
		if err != nil {
			t.Fatalf("unexpected error during parsing reference %s: %v", refStr, err)
		}

		if err := store.AddReference(id, primaryRefNamed, refNamed); err != nil {
			t.Fatalf("unexpected error during add reference %v: %v", refNamed, err)
		}
	}

	got := store.GetReferences(id)
	assert.Equal(t, len(got), len(otherRefStrs), "expected to return three references, but got %v", len(got))

	for _, refNamed := range got {
		_, ok := otherRefStrs[refNamed.String()]
		assert.Equal(t, ok, true, "expected to have %s", refNamed.String())
	}
}

func TestSearch(t *testing.T) {
	store, err := newImageStore()
	if err != nil {
		t.Fatalf("unexpected error during creating store: %v", err)
	}

	var (
		id      = digest.Digest("sha256:dc5f67a48da730d67bf4bfb8824ea8a51be26711de090d6d5a1ffff2723168a1")
		otherID = digest.Digest("sha256:dc5f67a48da730d67bf4bfb8824ea8a51be26711de090d6d5a1ffff2723168a3")

		dig = digest.Digest("sha256:dc5f67a48da730d67bf4bfb8824ea8a51be26711de090d6d5a1ffff2723168a2")

		primaryRefStr = "busybox:latest"
		otherRefStrs  = map[string]struct{}{
			"busybox:latest":                        {},
			"busybox:whatever" + "@" + dig.String(): {},
			"localhost:5000/busybox:1.25":           {},
		}

		otherIDRefStr = "alpine:latest"
	)

	// prepare reference
	{
		primaryRefNamed, err := reference.Parse(primaryRefStr)
		if err != nil {
			t.Fatalf("unexpected error during parsing reference %s: %v", primaryRefStr, err)
		}

		for refStr := range otherRefStrs {
			refNamed, err := reference.Parse(refStr)
			if err != nil {
				t.Fatalf("unexpected error during parsing reference %s: %v", refStr, err)
			}

			if err := store.AddReference(id, primaryRefNamed, refNamed); err != nil {
				t.Fatalf("unexpected error during add reference %v: %v", refNamed, err)
			}
		}

		primaryRefNamed, err = reference.Parse(otherIDRefStr)
		if err != nil {
			t.Fatalf("unexpected error during parsing reference %s: %v", otherIDRefStr, err)
		}

		if err := store.AddReference(otherID, primaryRefNamed, primaryRefNamed); err != nil {
			t.Fatalf("unexpected error during add reference %v: %v", primaryRefNamed, err)
		}
	}

	got := store.GetReferences(id)
	assert.Equal(t, len(got), len(otherRefStrs), "expected to return %v references, but got %v", len(otherRefStrs), len(got))

	// search
	{
		// should return id if the reference is id without algorithm header
		{
			namedStr := id.Hex()

			namedRef, err := reference.Parse(namedStr)
			if err != nil {
				t.Fatalf("unexpected error during parse reference %v: %v", namedStr, err)
			}

			gotID, gotRef, err := store.Search(namedRef)
			assert.Equal(t, err, nil)
			assert.Equal(t, gotID.String(), id.String())
			assert.Equal(t, gotRef.String(), namedRef.String())
		}

		// should return id if the reference is digest id
		{
			namedStr := id.String()

			namedRef, err := reference.Parse(namedStr)
			if err != nil {
				t.Fatalf("unexpected error during parse reference %v: %v", namedStr, err)
			}

			gotID, gotRef, err := store.Search(namedRef)
			assert.Equal(t, err, nil)
			assert.Equal(t, gotID.String(), id.String())
			assert.Equal(t, gotRef.String(), namedRef.String())
		}

		// should return busybox:latest if the reference is busybox
		{
			namedStr := "busybox"

			namedRef, err := reference.Parse(namedStr)
			if err != nil {
				t.Fatalf("unexpected error during parse reference %v: %v", namedStr, err)
			}

			gotID, gotRef, err := store.Search(namedRef)
			assert.Equal(t, err, nil)
			assert.Equal(t, gotID.String(), id.String())
			assert.Equal(t, gotRef.String(), namedRef.String()+":latest")
		}

		// should return busybox:latest@digest if the reference is busybox:latest@digest
		{
			namedStr := "busybox:latest@" + dig.String()

			namedRef, err := reference.Parse(namedStr)
			if err != nil {
				t.Fatalf("unexpected error during parse reference %v: %v", namedStr, err)
			}

			gotID, gotRef, err := store.Search(namedRef)
			assert.Equal(t, err, nil)
			assert.Equal(t, gotID.String(), id.String())
			assert.Equal(t, gotRef.String(), namedRef.String())
		}

		// should return busybox@digest if the reference is busybox@digest
		{
			namedStr := "busybox@" + dig.String()

			namedRef, err := reference.Parse(namedStr)
			if err != nil {
				t.Fatalf("unexpected error during parse reference %v: %v", namedStr, err)
			}

			gotID, gotRef, err := store.Search(namedRef)
			assert.Equal(t, err, nil)
			assert.Equal(t, gotID.String(), id.String())
			assert.Equal(t, gotRef.String(), namedRef.String())
		}

		// should return 404 if the reference is busybox@id
		{
			namedStr := "busybox@" + id.String()

			namedRef, err := reference.Parse(namedStr)
			if err != nil {
				t.Fatalf("unexpected error during parse reference %v: %v", namedStr, err)
			}

			_, _, err = store.Search(namedRef)
			assert.Equal(t, errtypes.IsNotfound(err), true)
			assert.Equal(t, err.Error(), "image: "+namedRef.String()+": not found")
		}

		// should return 404 if the reference is 123x
		{
			namedStr := "123x"

			namedRef, err := reference.Parse(namedStr)
			if err != nil {
				t.Fatalf("unexpected error during parse reference %v: %v", namedStr, err)
			}

			_, _, err = store.Search(namedRef)
			assert.Equal(t, errtypes.IsNotfound(err), true)
			assert.Equal(t, err.Error(), "image: "+namedRef.String()+": not found")
		}

		// should return ErrTooMany if the reference is commonPart
		{
			namedStr := id.String()[:20]

			namedRef, err := reference.Parse(namedStr)
			if err != nil {
				t.Fatalf("unexpected error during parse reference %v: %v", namedStr, err)
			}

			_, _, err = store.Search(namedRef)
			assert.Equal(t, pkgerrors.Cause(err), errtypes.ErrTooMany)
		}
	}

	// get primary reference
	{
		// should return busybox:latest if the reference is localhost:5000/busybox:1.25
		{
			namedStr := "localhost:5000/busybox:1.25"

			namedRef, err := reference.Parse(namedStr)
			if err != nil {
				t.Fatalf("unexpected error during parse reference %v: %v", namedStr, err)
			}

			pRef, err := store.GetPrimaryReference(namedRef)
			assert.Equal(t, err, nil)
			assert.Equal(t, pRef.String(), "busybox:latest")
		}

		// should return busybox:latest if the reference is busybox:1.27989@digest
		{
			namedStr := "busybox:1.27989@" + dig.String()

			namedRef, err := reference.Parse(namedStr)
			if err != nil {
				t.Fatalf("unexpected error during parse reference %v: %v", namedStr, err)
			}

			pRef, err := store.GetPrimaryReference(namedRef)
			assert.Equal(t, err, nil)
			assert.Equal(t, pRef.String(), "busybox:latest")
		}

		// should return 404 if the reference is whatever
		{
			namedStr := "whatever"

			namedRef, err := reference.Parse(namedStr)
			if err != nil {
				t.Fatalf("unexpected error during parse reference %v: %v", namedStr, err)
			}

			_, err = store.GetPrimaryReference(namedRef)
			assert.Equal(t, errtypes.IsNotfound(err), true)
		}
	}

	{
		// remove busybox:latest@digest
		{
			namedStr := "busybox:latest@" + dig.String()
			namedRef, err := reference.Parse(namedStr)
			if err != nil {
				t.Fatalf("unexpected error during parse reference %v: %v", namedStr, err)
			}

			assert.Equal(t, store.RemoveReference(id, namedRef), nil)

			_, _, err = store.Search(namedRef)
			assert.Equal(t, errtypes.IsNotfound(err), true)
			assert.Equal(t, len(store.GetReferences(id)), 2)
		}

		// remove the primary reference
		{
			namedRef, err := reference.Parse(primaryRefStr)
			if err != nil {
				t.Fatalf("unexpected error during parse reference %v: %v", primaryRefStr, err)
			}

			assert.Equal(t, store.RemoveReference(id, namedRef), nil)
			assert.Equal(t, len(store.GetReferences(id)), 0)
		}
	}
}

func TestAdd(t *testing.T) {
	store, err := newImageStore()
	if err != nil {
		t.Fatalf("unexpected error during creating store: %v", err)
	}

	var (
		primaryRef1 = "busybox:latest"
		id1         = digest.Digest("sha256:dc5f67a48da730d67bf4bfb8824ea8a51be26711de090d6d5a1ffff2723168a1")

		primaryRef2 = "busybox:1.25@sha256:dc5f67a48da730d67bf4bfb8824ea8a51be26711de090d6d5a1ffff272316812"
		id2         = digest.Digest("sha256:dc5f67a48da730d67bf4bfb8824ea8a51be26711de090d6d5a1ffff2723168a2")

		primaryRefNamed1, _ = reference.Parse(primaryRef1)
		primaryRefNamed2, _ = reference.Parse(primaryRef2)
	)

	// should pass if add primary reference
	{
		assert.Equal(t, store.AddReference(id1, primaryRefNamed1, primaryRefNamed1), nil)

		gotID, gotRef, err := store.Search(primaryRefNamed1)
		assert.Equal(t, err, nil)
		assert.Equal(t, gotID.String(), id1.String())
		assert.Equal(t, gotRef.String(), primaryRef1)
	}

	// should pass if add same reference
	{
		assert.Equal(t, store.AddReference(id1, primaryRefNamed1, primaryRefNamed1), nil)

		gotID, gotRef, err := store.Search(primaryRefNamed1)
		assert.Equal(t, err, nil)
		assert.Equal(t, gotID.String(), id1.String())
		assert.Equal(t, gotRef.String(), primaryRef1)
		assert.Equal(t, len(store.GetReferences(id1)), 1)
	}

	// should fail if add reference used as other primary reference
	{
		err := store.AddReference(id2, primaryRefNamed2, primaryRefNamed1)
		assert.Equal(t, strings.Contains(err.Error(), "cannot replace primary reference"), true)

		assert.Equal(t, len(store.GetReferences(id1)), 1)
		assert.Equal(t, len(store.GetReferences(id2)), 0)
	}

	// should fail if add reference which only contains the name
	{
		namedStr := "oopsfail1"

		namedRef, err := reference.Parse(namedStr)
		if err != nil {
			t.Fatalf("unexpected error during parse reference %v: %v", namedStr, err)
		}

		if err := store.AddReference(id2, primaryRefNamed2, namedRef); err == nil {
			t.Fatalf("expected fail but got nothing")
		}

		_, _, err = store.Search(namedRef)
		assert.Equal(t, errtypes.IsNotfound(err), true)
	}
}
