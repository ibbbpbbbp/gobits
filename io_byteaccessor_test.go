package gobits

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewIOByteAccessor(t *testing.T) {
	rwseeker, teardown := setupTestDataFile(t)
	defer teardown()
	ba := NewIOByteAccessor(rwseeker)
	assert.NotNil(t, ba)
}

func TestIOByteAccessor_At(t *testing.T) {
	rwseeker, teardown := setupTestDataFile(t)
	defer teardown()
	ba := NewIOByteAccessor(rwseeker)
	assert.NotNil(t, ba)

	t.Run("out_of_range_at_minus1", func(t *testing.T) {
		at_minus1, ok := ba.At(-1)
		assert.False(t, ok)
		assert.Equal(t, rawAt(rwseeker, -1), at_minus1)
		assert.Zero(t, at_minus1)
	})
	t.Run("in_range_at0", func(t *testing.T) {
		at0, ok := ba.At(0)
		assert.True(t, ok)
		assert.Equal(t, rawAt(rwseeker, 0), at0)
		assert.Equal(t, rawAt(rwseeker, 0), ba.buffer[0])
		assert.Equal(t, rawAt(rwseeker, 4096-1), ba.buffer[len(ba.buffer)-1])
	})
	t.Run("in_range_at4096", func(t *testing.T) {
		at4096, ok := ba.At(4096)
		assert.True(t, ok)
		assert.Equal(t, rawAt(rwseeker, 4096), at4096)
		assert.Equal(t, rawAt(rwseeker, 4096-(4096/2)), ba.buffer[0])
		assert.Equal(t, rawAt(rwseeker, 4096-(4096/2)+4096-1), ba.buffer[len(ba.buffer)-1])
	})
	t.Run("in_range_at7247", func(t *testing.T) {
		at7247, ok := ba.At(7247)
		assert.True(t, ok)
		assert.Equal(t, rawAt(rwseeker, 7247), at7247)
		assert.Equal(t, rawAt(rwseeker, 7247-(4096/2)), ba.buffer[0])
		assert.Equal(t, rawAt(rwseeker, 7247), ba.buffer[len(ba.buffer)-1])
	})
	t.Run("out_of_range_at7248", func(t *testing.T) {
		at7248, ok := ba.At(7248)
		assert.False(t, ok)
		assert.Equal(t, rawAt(rwseeker, 7248), at7248)
		assert.Zero(t, at7248)
	})
}

func TestIOByteAccessor_Slice(t *testing.T) {
	rwseeker, teardown := setupTestDataFile(t)
	defer teardown()
	ba := NewIOByteAccessor(rwseeker)
	assert.NotNil(t, ba)

	t.Run("in_range", func(t *testing.T) {
		s1 := ba.Slice(4096, 1024)
		assert.Equal(t, rawSlice(rwseeker, 4096, 1024), s1)

		s2 := ba.Slice(7247, 1)
		assert.Equal(t, rawSlice(rwseeker, 7247, 1), s2)
	})
	t.Run("out_of_range", func(t *testing.T) {
		s1 := ba.Slice(-1, 1024)
		assert.Equal(t, []byte{}, s1)

		s2 := ba.Slice(7247, 1024)
		assert.Equal(t, rawSlice(rwseeker, 7247, 1024), s2)

		s3 := ba.Slice(7248, 0)
		assert.Equal(t, []byte{}, s3)

		s4 := ba.Slice(7248, 1)
		assert.Equal(t, []byte{}, s4)
	})
	t.Run("invalid_length", func(t *testing.T) {
		s1 := ba.Slice(0, 0)
		assert.Equal(t, []byte{}, s1)

		s2 := ba.Slice(0, -1)
		assert.Equal(t, []byte{}, s2)
	})
}

func TestIOByteAccessor_Put(t *testing.T) {
	t.Run("in_range1", func(t *testing.T) {
		rwseeker, teardown := setupTestDataFile(t)
		defer teardown()
		ba := NewIOByteAccessor(rwseeker)
		assert.NotNil(t, ba)

		before := rawSlice(rwseeker, 0, 1024)
		after := rawSlice(rwseeker, 4096, 1024)
		assert.True(t, ba.Put(after, 0))
		assert.Equal(t, after, rawSlice(rwseeker, 0, 1024))

		at0, _ := ba.At(0)
		assert.Equal(t, after[0], at0)
		at1023, _ := ba.At(1023)
		assert.Equal(t, after[1023], at1023)

		assert.True(t, ba.Put(before, 0))
		assert.Equal(t, before, rawSlice(rwseeker, 0, 1024))

		at0, _ = ba.At(0)
		assert.Equal(t, before[0], at0)
		at1023, _ = ba.At(1023)
		assert.Equal(t, before[1023], at1023)

	})
	t.Run("in_range2", func(t *testing.T) {
		rwseeker, teardown := setupTestDataFile(t)
		defer teardown()
		ba := NewIOByteAccessor(rwseeker)
		assert.NotNil(t, ba)

		before := rawSlice(rwseeker, 4096, 1024)
		after := rawSlice(rwseeker, 0, 1024)
		assert.True(t, ba.Put(after, 4096))
		assert.Equal(t, after, rawSlice(rwseeker, 4096, 1024))

		at4096, _ := ba.At(4096)
		assert.Equal(t, after[0], at4096)
		at5119, _ := ba.At(4096 + 1023)
		assert.Equal(t, after[1023], at5119)

		assert.True(t, ba.Put(before, 4096))
		assert.Equal(t, before, rawSlice(rwseeker, 4096, 1024))

		at4096, _ = ba.At(4096)
		assert.Equal(t, before[0], at4096)
		at5119, _ = ba.At(4096 + 1023)
		assert.Equal(t, before[1023], at5119)

	})
	t.Run("in_range3", func(t *testing.T) {
		rwseeker, teardown := setupTestDataFile(t)
		defer teardown()
		ba := NewIOByteAccessor(rwseeker)
		assert.NotNil(t, ba)

		before := rawSlice(rwseeker, 1024, 2048)
		after := rawSlice(rwseeker, 0, 2048)
		assert.True(t, ba.Put(after, 1024))
		assert.Equal(t, after, rawSlice(rwseeker, 1024, 2048))

		at := make([]byte, 10)
		for i := 0; i < len(at); i++ {
			at[i], _ = ba.At(4096 + int64(i))
		}
		assert.Equal(t, rawSlice(rwseeker, 4096, 10), at)

		assert.True(t, ba.Put(before, 1024))
		assert.Equal(t, before, rawSlice(rwseeker, 1024, 2048))

		at = make([]byte, 100)
		for i := 0; i < len(at); i++ {
			at[i], _ = ba.At(2048 + int64(i))
		}
		assert.Equal(t, rawSlice(rwseeker, 2048, int64(len(at))), at)
		at = make([]byte, 100)
		for i := 0; i < len(at); i++ {
			at[i], _ = ba.At(3000 + int64(i))
		}
		assert.Equal(t, rawSlice(rwseeker, 3000, int64(len(at))), at)
	})
	t.Run("out_of_range", func(t *testing.T) {
		rwseeker, teardown := setupTestDataFile(t)
		defer teardown()
		ba := NewIOByteAccessor(rwseeker)
		assert.NotNil(t, ba)

		w := []byte{1, 2, 3, 4, 5}
		assert.False(t, ba.Put(w, -1))

		backupTestDataFile(rwseeker)
		assert.True(t, ba.Put(w, 8000))
		for i, expect := range w {
			at, ok := ba.At(8000 + int64(i))
			assert.True(t, ok)
			assert.Equal(t, expect, at)
		}
		recoverDataFile()
	})
	t.Run("invalid_length", func(t *testing.T) {
		rwseeker, teardown := setupTestDataFile(t)
		defer teardown()
		ba := NewIOByteAccessor(rwseeker)
		assert.NotNil(t, ba)

		s1 := ba.Slice(0, 5)
		w1 := []byte{}
		assert.True(t, ba.Put(w1, 0))
		s1_ := ba.Slice(0, 5)
		assert.Equal(t, s1, s1_)

		s2 := ba.Slice(0, 5)
		assert.False(t, ba.Put(nil, 0))
		s2_ := ba.Slice(0, 5)
		assert.Equal(t, s2, s2_)
	})
}

func TestIOByteAccessor_Length(t *testing.T) {
	rwseeker, teardown := setupTestDataFile(t)
	defer teardown()
	ba := NewIOByteAccessor(rwseeker)
	assert.NotNil(t, ba)

	assert.Equal(t, int64(7248), ba.Length())
}

func TestIOByteAccessor_Reset(t *testing.T) {
	rwseeker, teardown := setupTestDataFile(t)
	defer teardown()
	ba := NewIOByteAccessor(rwseeker)
	assert.NotNil(t, ba)

	at7247, ok := ba.At(7247)
	assert.True(t, ok)
	assert.Equal(t, rawAt(rwseeker, 7247), at7247)
	assert.Equal(t, rawAt(rwseeker, 7247-(4096/2)), ba.buffer[0])
	assert.Equal(t, rawAt(rwseeker, 7247), ba.buffer[len(ba.buffer)-1])

	ba.Reset()

	at0, ok := ba.At(0)
	assert.True(t, ok)
	assert.Equal(t, rawAt(rwseeker, 0), at0)
	assert.Equal(t, rawAt(rwseeker, 0), ba.buffer[0])
	assert.Equal(t, rawAt(rwseeker, 4095), ba.buffer[len(ba.buffer)-1])
}
