package ctrd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_containerLock_Trylock(t *testing.T) {
	l := &containerLock{
		ids: make(map[string]struct{}),
	}

	assert.Equal(t, len(l.ids), 0)

	// lock a new element
	ok := l.Trylock("element1")
	assert.Equal(t, ok, true)
	assert.Equal(t, len(l.ids), 1)
	assert.Equal(t, l.ids["element1"], struct{}{})

	// lock an existent element
	ok = l.Trylock("element1")
	assert.Equal(t, ok, false)
	assert.Equal(t, len(l.ids), 1)
	assert.Equal(t, l.ids["element1"], struct{}{})

	// lock another new element
	ok = l.Trylock("element2")
	assert.Equal(t, ok, true)
	assert.Equal(t, len(l.ids), 2)
	assert.Equal(t, l.ids["element1"], struct{}{})
}

func Test_containerLock_Unlock(t *testing.T) {
	l := &containerLock{
		ids: make(map[string]struct{}),
	}

	// unlock a non-existent element
	l.Unlock("non-existent")
	assert.Equal(t, len(l.ids), 0)

	// lock a new element
	ok := l.Trylock("element1")
	assert.Equal(t, ok, true)
	assert.Equal(t, len(l.ids), 1)
	assert.Equal(t, l.ids["element1"], struct{}{})

	// unlock an existent element
	l.Unlock("element1")
	assert.Equal(t, len(l.ids), 0)
}
