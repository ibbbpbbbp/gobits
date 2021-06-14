package gobits

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSliceByteAccessor(t *testing.T) {
	s := []byte{}
	ba := NewSliceByteAccessor(s)
	assert.NotNil(t, ba)
}

func TestSliceByteAccessor_At(t *testing.T) {
	s := []byte{1, 2, 3, 4, 5}
	ba := NewSliceByteAccessor(s)
	assert.NotNil(t, ba)

	t.Run("in_rage", func(t *testing.T) {
		for i, expect := range s {
			actual, ok := ba.At(int64(i))
			assert.True(t, ok)
			assert.Equal(t, expect, actual)
		}
	})
	t.Run("out_of_range_at_minus1", func(t *testing.T) {
		v, ok := ba.At(-1)
		assert.False(t, ok)
		assert.Equal(t, byte(0), v)
	})
	t.Run("out_of_range_at5", func(t *testing.T) {
		v, ok := ba.At(int64(len(s)))
		assert.False(t, ok)
		assert.Equal(t, byte(0), v)
	})
}

func TestSliceByteAccessor_Slice(t *testing.T) {
	s := []byte{1, 2, 3, 4, 5}
	ba := NewSliceByteAccessor(s)
	assert.NotNil(t, ba)

	t.Run("all_match", func(t *testing.T) {
		s1 := ba.Slice(0, int64(len(s)))
		assert.Equal(t, s, s1)
	})
	t.Run("partial_match", func(t *testing.T) {
		s1 := ba.Slice(1, 3)
		assert.Equal(t, s[1:4], s1)
	})
	t.Run("out_of_range", func(t *testing.T) {
		s1 := ba.Slice(-1, 1)
		assert.Equal(t, []byte{}, s1)

		for i := 0; i < len(s); i++ {
			s2 := ba.Slice(int64(i), int64(len(s))+1)
			assert.Equal(t, s[i:], s2)
		}

		s3 := ba.Slice(int64(len(s)), 1)
		assert.Equal(t, []byte{}, s3)
	})
	t.Run("invalid_length", func(t *testing.T) {
		s1 := ba.Slice(0, 0)
		assert.Equal(t, []byte{}, s1)

		s2 := ba.Slice(0, -1)
		assert.Equal(t, []byte{}, s2)
	})
}

func TestSliceByteAccessor_Put(t *testing.T) {
	t.Run("in_range", func(t *testing.T) {
		s := []byte{1, 2, 3, 4, 5}
		ba := NewSliceByteAccessor(s)
		assert.NotNil(t, ba)

		w := []byte{10, 20, 30}
		assert.True(t, ba.Put(w, 1))
		assert.Equal(t, []byte{1, 10, 20, 30, 5}, ba.Slice(0, int64(len(s))))

		assert.True(t, ba.Put(w, 2))
		assert.Equal(t, []byte{1, 10, 10, 20, 30}, ba.Slice(0, int64(len(s))))
	})
	t.Run("out_of_range", func(t *testing.T) {
		s := []byte{1, 2, 3, 4, 5}
		ba := NewSliceByteAccessor(s)
		assert.NotNil(t, ba)

		w := []byte{10, 20, 30}
		assert.True(t, ba.Put(w, 3))
		assert.Equal(t, []byte{1, 2, 3, 10, 20}, ba.Slice(0, int64(len(s))))

		assert.False(t, ba.Put(w, -1))
		assert.Equal(t, []byte{1, 2, 3, 10, 20}, ba.Slice(0, int64(len(s))))
	})
	t.Run("invalid_length", func(t *testing.T) {
		s := []byte{1, 2, 3, 4, 5}
		ba := NewSliceByteAccessor(s)
		assert.NotNil(t, ba)

		w := []byte{}
		assert.True(t, ba.Put(w, 3))
		assert.Equal(t, []byte{1, 2, 3, 4, 5}, ba.Slice(0, int64(len(s))))

		assert.False(t, ba.Put(nil, 3))
		assert.Equal(t, []byte{1, 2, 3, 4, 5}, ba.Slice(0, int64(len(s))))
	})
}

func TestSliceByteAccessor_Length(t *testing.T) {
	s := []byte{1, 2, 3, 4, 5, 6}
	ba := NewSliceByteAccessor(s)
	assert.NotNil(t, ba)

	assert.Equal(t, int64(6), ba.Length())
}
