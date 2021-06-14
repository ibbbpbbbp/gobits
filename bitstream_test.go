package gobits

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBitStream(t *testing.T) {
	bs := NewBitStream(NewSliceByteAccessor([]byte{1}))
	assert.NotNil(t, bs)
}

func TestBitStream_RemainingBits(t *testing.T) {
	t.Run("slice_byteaccessor", func(t *testing.T) {
		bs := NewBitStream(NewSliceByteAccessor([]byte{1, 2, 3, 4, 5}))

		assert.True(t, bs.RemainingBits(0))
		assert.True(t, bs.RemainingBits(39))
		assert.True(t, bs.RemainingBits(40))
		assert.False(t, bs.RemainingBits(41))

		assert.True(t, bs.Seek(4, 7))

		assert.True(t, bs.RemainingBits(0))
		assert.True(t, bs.RemainingBits(1))
		assert.False(t, bs.RemainingBits(2))
	})
	t.Run("io_byteaccessor", func(t *testing.T) {
		rwseeker, teardown := setupTestDataFile(t)
		defer teardown()
		bs := NewBitStream(NewIOByteAccessor(rwseeker))

		assert.True(t, bs.RemainingBits(0))
		assert.True(t, bs.RemainingBits(39))
		assert.True(t, bs.RemainingBits(40))
		assert.True(t, bs.RemainingBits(7248*8))
		assert.False(t, bs.RemainingBits(7248*8+1))

		assert.True(t, bs.Seek(7247, 7))

		assert.True(t, bs.RemainingBits(0))
		assert.True(t, bs.RemainingBits(1))
		assert.False(t, bs.RemainingBits(2))
	})
}

// Test together at "TestBitStream_ReadBits"
/*
func TestBitStream_PeekBits {
}
*/

func TestBitStream_ConsumeBits(t *testing.T) {
	t.Run("slice_byteaccessor", func(t *testing.T) {
		bs := NewBitStream(NewSliceByteAccessor([]byte{1, 2, 3, 4, 5}))

		assert.True(t, bs.ConsumeBits(20))
		assert.Equal(t, int64(2), bs.byteOffset)
		assert.Equal(t, byte(4), bs.bitOffset)

		assert.True(t, bs.ConsumeBits(19))
		assert.Equal(t, int64(4), bs.byteOffset)
		assert.Equal(t, byte(7), bs.bitOffset)

		assert.True(t, bs.ConsumeBits(1))
		assert.Equal(t, int64(5), bs.byteOffset)
		assert.Equal(t, byte(0), bs.bitOffset)

		assert.True(t, bs.ConsumeBits(0))
		assert.False(t, bs.ConsumeBits(1))
	})
	t.Run("io_byteaccessor", func(t *testing.T) {
		rwseeker, teardown := setupTestDataFile(t)
		defer teardown()
		bs := NewBitStream(NewIOByteAccessor(rwseeker))

		assert.True(t, bs.ConsumeBits(2020*8+4))
		assert.Equal(t, int64(2020), bs.byteOffset)
		assert.Equal(t, byte(4), bs.bitOffset)

		assert.True(t, bs.ConsumeBits(2020*8+4))
		assert.Equal(t, int64(4041), bs.byteOffset)
		assert.Equal(t, byte(0), bs.bitOffset)

		assert.True(t, bs.ConsumeBits(3206*8+7))
		assert.Equal(t, int64(7247), bs.byteOffset)
		assert.Equal(t, byte(7), bs.bitOffset)

		assert.True(t, bs.ConsumeBits(1))
		assert.Equal(t, int64(7248), bs.byteOffset)
		assert.Equal(t, byte(0), bs.bitOffset)

		assert.True(t, bs.ConsumeBits(0))
		assert.False(t, bs.ConsumeBits(1))
	})
}

func TestBitStream_ConsumeBytes(t *testing.T) {
	t.Run("slice_byteaccessor", func(t *testing.T) {
		bs := NewBitStream(NewSliceByteAccessor([]byte{1, 2, 3, 4, 5}))

		assert.True(t, bs.ConsumeBytes(2))
		assert.Equal(t, int64(2), bs.byteOffset)
		assert.Equal(t, byte(0), bs.bitOffset)

		assert.True(t, bs.ConsumeBytes(2))
		assert.Equal(t, int64(4), bs.byteOffset)
		assert.Equal(t, byte(0), bs.bitOffset)

		assert.True(t, bs.ConsumeBytes(1))
		assert.Equal(t, int64(5), bs.byteOffset)
		assert.Equal(t, byte(0), bs.bitOffset)

		assert.True(t, bs.ConsumeBits(0))
		assert.False(t, bs.ConsumeBits(1))
	})
	t.Run("io_byteaccessor", func(t *testing.T) {
		rwseeker, teardown := setupTestDataFile(t)
		defer teardown()
		bs := NewBitStream(NewIOByteAccessor(rwseeker))

		assert.True(t, bs.ConsumeBytes(2020))
		assert.Equal(t, int64(2020), bs.byteOffset)
		assert.Equal(t, byte(0), bs.bitOffset)

		assert.True(t, bs.ConsumeBytes(2020))
		assert.Equal(t, int64(4040), bs.byteOffset)
		assert.Equal(t, byte(0), bs.bitOffset)

		assert.True(t, bs.ConsumeBytes(3207))
		assert.Equal(t, int64(7247), bs.byteOffset)
		assert.Equal(t, byte(0), bs.bitOffset)

		assert.True(t, bs.ConsumeBytes(1))
		assert.Equal(t, int64(7248), bs.byteOffset)
		assert.Equal(t, byte(0), bs.bitOffset)

		assert.True(t, bs.ConsumeBits(0))
		assert.False(t, bs.ConsumeBits(1))
	})
}

func TestBitStream_ReadBits(t *testing.T) {
	t.Run("slice_byteaccessor", func(t *testing.T) {
		bs := NewBitStream(NewSliceByteAccessor(
			[]byte{
				0xa5, 0x5a, 0xa5, 0x5a,
				0xa5, 0x5a, 0xa5, 0x5a,
				0xa5, 0x5a, 0xa5, 0x5a,
			}))

		bits, ok := bs.ReadBits(0)
		assert.True(t, ok)
		assert.Equal(t, uint64(0), bits)

		bits, ok = bs.ReadBits(65)
		assert.False(t, ok)
		assert.Equal(t, uint64(0), bits)

		bits, ok = bs.ReadBits(8)
		assert.True(t, ok)
		assert.Equal(t, uint64(0xa5), bits)
		bs.Seek(0, 0)

		bits, ok = bs.ReadBits(64)
		assert.True(t, ok)
		assert.Equal(t, uint64(0xa55aa55aa55aa55a), bits)

		bits, ok = bs.ReadBits(32)
		assert.True(t, ok)
		assert.Equal(t, uint64(0xa55aa55a), bits)

		bs.Seek(0, 0)

		bits, ok = bs.ReadBits(3)
		assert.True(t, ok)
		assert.Equal(t, uint64(5), bits)

		bits, ok = bs.ReadBits(3)
		assert.True(t, ok)
		assert.Equal(t, uint64(1), bits)

		bits, ok = bs.ReadBits(3)
		assert.True(t, ok)
		assert.Equal(t, uint64(2), bits)

		bits, ok = bs.ReadBits(6)
		assert.True(t, ok)
		assert.Equal(t, uint64(45), bits)

		bits, ok = bs.ReadBits(64)
		assert.True(t, ok)
		assert.Equal(t, uint64(0x52ad52ad52ad52ad), bits)

		bits, ok = bs.ReadBits(18)
		assert.False(t, ok)
		assert.Equal(t, uint64(0), bits)

		bits, ok = bs.ReadBits(17)
		assert.True(t, ok)
		assert.Equal(t, uint64(0xa55a), bits)

		bits, ok = bs.ReadBits(0)
		assert.True(t, ok)
		assert.Equal(t, uint64(0), bits)

		bits, ok = bs.ReadBits(1)
		assert.False(t, ok)
		assert.Equal(t, uint64(0), bits)
	})
	t.Run("io_byteaccessor", func(t *testing.T) {
		rwseeker, teardown := setupTestDataFile(t)
		defer teardown()
		bs := NewBitStream(NewIOByteAccessor(rwseeker))

		bits, ok := bs.ReadBits(0)
		assert.True(t, ok)
		assert.Equal(t, uint64(0), bits)

		bits, ok = bs.ReadBits(65)
		assert.False(t, ok)
		assert.Equal(t, uint64(0), bits)

		bits, ok = bs.ReadBits(8)
		assert.True(t, ok)
		assert.Equal(t, uint64(0xff), bits)

		bits, ok = bs.ReadBits(16)
		assert.True(t, ok)
		assert.Equal(t, uint64(0xd8ff), bits)

		bits, ok = bs.ReadBits(24)
		assert.True(t, ok)
		assert.Equal(t, uint64(0xe00010), bits)

		bits, ok = bs.ReadBits(32)
		assert.True(t, ok)
		assert.Equal(t, uint64(0x4a464946), bits)

		bits, ok = bs.ReadBits(40)
		assert.True(t, ok)
		assert.Equal(t, uint64(0x0001010100), bits)

		bits, ok = bs.ReadBits(48)
		assert.True(t, ok)
		assert.Equal(t, uint64(0x4800480000ff), bits)

		bits, ok = bs.ReadBits(56)
		assert.True(t, ok)
		assert.Equal(t, uint64(0xdb004300050304), bits)

		bits, ok = bs.ReadBits(64)
		assert.True(t, ok)
		assert.Equal(t, uint64(0x404030504040405), bits)

		bits, ok = bs.ReadBits(5)
		assert.True(t, ok)
		assert.Equal(t, uint64(0), bits)

		bits, ok = bs.ReadBits(5)
		assert.True(t, ok)
		assert.Equal(t, uint64(0x14), bits)

		bits, ok = bs.ReadBits(6)
		assert.True(t, ok)
		assert.Equal(t, uint64(0x5), bits)

		assert.True(t, bs.Seek(7240, 0))

		bits, ok = bs.ReadBits(64)
		assert.True(t, ok)
		assert.Equal(t, uint64(0xb23484538a3fffd9), bits)

		bits, ok = bs.ReadBits(0)
		assert.True(t, ok)
		assert.Equal(t, uint64(0), bits)

		bits, ok = bs.ReadBits(1)
		assert.False(t, ok)
		assert.Equal(t, uint64(0), bits)
	})
}

func TestBitStream_ReadUint(t *testing.T) {
	bs := NewBitStream(NewSliceByteAccessor([]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88}))
	u8, ok := bs.ReadUint8()
	assert.True(t, ok)
	assert.Equal(t, uint8(0x11), u8)

	ule16, ok := bs.ReadUint16(binary.LittleEndian)
	assert.True(t, ok)
	assert.Equal(t, uint16(0x3322), ule16)

	ube16, ok := bs.ReadUint16(binary.BigEndian)
	assert.True(t, ok)
	assert.Equal(t, uint16(0x4455), ube16)

	assert.True(t, bs.Seek(0, 0))

	ule32, ok := bs.ReadUint32(binary.LittleEndian)
	assert.True(t, ok)
	assert.Equal(t, uint32(0x44332211), ule32)

	ube32, ok := bs.ReadUint32(binary.BigEndian)
	assert.True(t, ok)
	assert.Equal(t, uint32(0x55667788), ube32)

	assert.True(t, bs.Seek(0, 0))

	ule64, ok := bs.ReadUint64(binary.LittleEndian)
	assert.True(t, ok)
	assert.Equal(t, uint64(0x8877665544332211), ule64)

	assert.True(t, bs.Seek(0, 0))

	ube64, ok := bs.ReadUint64(binary.BigEndian)
	assert.True(t, ok)
	assert.Equal(t, uint64(0x1122334455667788), ube64)
}

func TestBitStream_Seek(t *testing.T) {
	t.Run("slice_byteaccessor", func(t *testing.T) {
		bs := NewBitStream(NewSliceByteAccessor([]byte{1, 2, 3, 4, 5}))

		assert.True(t, bs.Seek(2, 4))
		assert.Equal(t, int64(2), bs.byteOffset)
		assert.Equal(t, byte(4), bs.bitOffset)

		assert.True(t, bs.Seek(4, 7))
		assert.Equal(t, int64(4), bs.byteOffset)
		assert.Equal(t, byte(7), bs.bitOffset)

		assert.False(t, bs.Seek(4, 8))
		assert.Equal(t, int64(4), bs.byteOffset)
		assert.Equal(t, byte(7), bs.bitOffset)

		assert.True(t, bs.ConsumeBits(1))
		assert.False(t, bs.ConsumeBits(1))
	})
	t.Run("io_byteaccessor", func(t *testing.T) {
		rwseeker, teardown := setupTestDataFile(t)
		defer teardown()
		bs := NewBitStream(NewIOByteAccessor(rwseeker))

		assert.True(t, bs.Seek(2020, 4))
		assert.Equal(t, int64(2020), bs.byteOffset)
		assert.Equal(t, byte(4), bs.bitOffset)

		assert.True(t, bs.Seek(4041, 0))
		assert.Equal(t, int64(4041), bs.byteOffset)
		assert.Equal(t, byte(0), bs.bitOffset)

		assert.True(t, bs.Seek(7247, 7))
		assert.Equal(t, int64(7247), bs.byteOffset)
		assert.Equal(t, byte(7), bs.bitOffset)

		assert.False(t, bs.Seek(7247, 8))
		assert.Equal(t, int64(7247), bs.byteOffset)
		assert.Equal(t, byte(7), bs.bitOffset)

		assert.True(t, bs.ConsumeBits(1))
		assert.False(t, bs.ConsumeBits(1))
	})
}

func TestBitStream_SaveRestorePos(t *testing.T) {
	bs := NewBitStream(NewSliceByteAccessor([]byte{1, 2, 3, 4, 5}))

	assert.True(t, bs.Seek(2, 4))
	assert.Equal(t, int64(2), bs.byteOffset)
	assert.Equal(t, byte(4), bs.bitOffset)

	pos := bs.SavePos()

	assert.True(t, bs.Seek(4, 7))
	assert.Equal(t, int64(4), bs.byteOffset)
	assert.Equal(t, byte(7), bs.bitOffset)

	assert.False(t, bs.Seek(4, 8))
	assert.Equal(t, int64(4), bs.byteOffset)
	assert.Equal(t, byte(7), bs.bitOffset)

	bs.RestorePos(pos)
	assert.Equal(t, int64(2), bs.byteOffset)
	assert.Equal(t, byte(4), bs.bitOffset)

	bs.ResetPos()
	assert.Equal(t, int64(0), bs.byteOffset)
	assert.Equal(t, byte(0), bs.bitOffset)

	bs.RestorePos(pos)
	assert.Equal(t, int64(2), bs.byteOffset)
	assert.Equal(t, byte(4), bs.bitOffset)
}

func TestBitStream_ReadExponentialGolomb(t *testing.T) {
	bs := NewBitStream(NewSliceByteAccessor([]byte{0xa6, 0x42, 0x98, 0xe2, 0x04, 0x8a}))
	expg, ok := bs.ReadExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, uint64(0), expg)
	expg, ok = bs.ReadExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, uint64(1), expg)
	expg, ok = bs.ReadExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, uint64(2), expg)
	expg, ok = bs.ReadExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, uint64(3), expg)
	expg, ok = bs.ReadExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, uint64(4), expg)
	expg, ok = bs.ReadExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, uint64(5), expg)
	expg, ok = bs.ReadExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, uint64(6), expg)
	expg, ok = bs.ReadExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, uint64(7), expg)
	expg, ok = bs.ReadExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, uint64(8), expg)
	expg, ok = bs.ReadExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, uint64(9), expg)

	expg, ok = bs.ReadExponentialGolomb()
	assert.False(t, ok)
	assert.Equal(t, uint64(0), expg)
}

func TestBitStream_ReadSignedExponentialGolomb(t *testing.T) {
	bs := NewBitStream(NewSliceByteAccessor([]byte{0xa6, 0x42, 0x98, 0xe2, 0x04, 0x8a}))
	sexpg, ok := bs.ReadSignedExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, int64(0), sexpg)
	sexpg, ok = bs.ReadSignedExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, int64(1), sexpg)
	sexpg, ok = bs.ReadSignedExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, int64(-1), sexpg)
	sexpg, ok = bs.ReadSignedExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, int64(2), sexpg)
	sexpg, ok = bs.ReadSignedExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, int64(-2), sexpg)
	sexpg, ok = bs.ReadSignedExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, int64(3), sexpg)
	sexpg, ok = bs.ReadSignedExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, int64(-3), sexpg)
	sexpg, ok = bs.ReadSignedExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, int64(4), sexpg)
	sexpg, ok = bs.ReadSignedExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, int64(-4), sexpg)
	sexpg, ok = bs.ReadSignedExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, int64(5), sexpg)

	sexpg, ok = bs.ReadSignedExponentialGolomb()
	assert.False(t, ok)
	assert.Equal(t, int64(0), sexpg)
}

func TestBitStream_WriteBits(t *testing.T) {
	bs := NewBitStream(NewSliceByteAccessor(make([]byte, 8))) // 64bits

	assert.True(t, bs.WriteBits(1, 1))
	assert.True(t, bs.WriteBits(1, 1))
	assert.True(t, bs.WriteBits(1, 1))
	assert.True(t, bs.WriteBits(1, 1))
	assert.True(t, bs.ConsumeBits(1))           // 5bits
	assert.True(t, bs.WriteBits(0x1234567, 28)) // 33bits
	assert.True(t, bs.ConsumeBits(4))           // 37bits
	assert.True(t, bs.WriteBits(0xabc, 12))     // 49bits
	assert.True(t, bs.ConsumeBits(4))           // 53bits
	assert.True(t, bs.WriteBits(0x7fe, 11))     // 64bits
	assert.False(t, bs.WriteBits(1, 1))         // over
	assert.True(t, bs.WriteBits(1, 0))

	assert.True(t, bs.Seek(0, 0))

	b, ok := bs.ReadBits(5)
	assert.True(t, ok)
	assert.Equal(t, uint64(0x1e), b)
	b, ok = bs.ReadBits(32)
	assert.True(t, ok)
	assert.Equal(t, uint64(0x12345670), b)
	b, ok = bs.ReadBits(16)
	assert.True(t, ok)
	assert.Equal(t, uint64(0xabc0), b)
	b, ok = bs.ReadBits(11)
	assert.True(t, ok)
	assert.Equal(t, uint64(0x7fe), b)
}

func TestBitStream_WriteUint(t *testing.T) {
	bs := NewBitStream(NewSliceByteAccessor(make([]byte, 8)))
	assert.True(t, bs.WriteUint8(0x11))
	assert.True(t, bs.Seek(0, 0))
	u8, ok := bs.ReadUint8()
	assert.True(t, ok)
	assert.Equal(t, uint8(0x11), u8)
	assert.True(t, bs.Seek(0, 0))

	assert.True(t, bs.WriteUint16(0x1122, binary.LittleEndian))
	assert.True(t, bs.Seek(0, 0))
	ule16, ok := bs.ReadUint16(binary.LittleEndian)
	assert.True(t, ok)
	assert.Equal(t, uint16(0x1122), ule16)
	assert.True(t, bs.Seek(0, 0))

	assert.True(t, bs.WriteUint16(0x1122, binary.BigEndian))
	assert.True(t, bs.Seek(0, 0))
	ube16, ok := bs.ReadUint16(binary.BigEndian)
	assert.True(t, ok)
	assert.Equal(t, uint16(0x1122), ube16)
	assert.True(t, bs.Seek(0, 0))

	assert.True(t, bs.WriteUint32(0x11223344, binary.LittleEndian))
	assert.True(t, bs.Seek(0, 0))
	ule32, ok := bs.ReadUint32(binary.LittleEndian)
	assert.True(t, ok)
	assert.Equal(t, uint32(0x11223344), ule32)
	assert.True(t, bs.Seek(0, 0))

	assert.True(t, bs.WriteUint32(0x11223344, binary.BigEndian))
	assert.True(t, bs.Seek(0, 0))
	ube32, ok := bs.ReadUint32(binary.BigEndian)
	assert.True(t, ok)
	assert.Equal(t, uint32(0x11223344), ube32)
	assert.True(t, bs.Seek(0, 0))

	assert.True(t, bs.WriteUint64(0x1122334455667788, binary.LittleEndian))
	assert.True(t, bs.Seek(0, 0))
	ule64, ok := bs.ReadUint64(binary.LittleEndian)
	assert.True(t, ok)
	assert.Equal(t, uint64(0x1122334455667788), ule64)
	assert.True(t, bs.Seek(0, 0))

	assert.True(t, bs.WriteUint64(0x1122334455667788, binary.BigEndian))
	assert.True(t, bs.Seek(0, 0))
	ube64, ok := bs.ReadUint64(binary.BigEndian)
	assert.True(t, ok)
	assert.Equal(t, uint64(0x1122334455667788), ube64)
	assert.True(t, bs.Seek(0, 0))
}

func TestBitStream_WriteExponentialGolomb(t *testing.T) {
	bs := NewBitStream(NewSliceByteAccessor(make([]byte, 8)))

	assert.True(t, bs.WriteExponentialGolomb(0))
	assert.True(t, bs.WriteExponentialGolomb(1))
	assert.True(t, bs.WriteExponentialGolomb(2))
	assert.True(t, bs.WriteExponentialGolomb(3))
	assert.True(t, bs.WriteExponentialGolomb(4))
	assert.True(t, bs.WriteExponentialGolomb(5))
	assert.True(t, bs.WriteExponentialGolomb(6))
	assert.True(t, bs.WriteExponentialGolomb(7))
	assert.True(t, bs.WriteExponentialGolomb(8))
	assert.True(t, bs.WriteExponentialGolomb(9))

	assert.True(t, bs.Seek(0, 0))

	expg, ok := bs.ReadExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, uint64(0), expg)
	expg, ok = bs.ReadExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, uint64(1), expg)
	expg, ok = bs.ReadExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, uint64(2), expg)
	expg, ok = bs.ReadExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, uint64(3), expg)
	expg, ok = bs.ReadExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, uint64(4), expg)
	expg, ok = bs.ReadExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, uint64(5), expg)
	expg, ok = bs.ReadExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, uint64(6), expg)
	expg, ok = bs.ReadExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, uint64(7), expg)
	expg, ok = bs.ReadExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, uint64(8), expg)
	expg, ok = bs.ReadExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, uint64(9), expg)

	expg, ok = bs.ReadExponentialGolomb()
	assert.False(t, ok)
	assert.Equal(t, uint64(0), expg)
}

func TestBitStream_WriteSignedExponentialGolomb(t *testing.T) {
	bs := NewBitStream(NewSliceByteAccessor(make([]byte, 8)))

	assert.True(t, bs.WriteSignedExponentialGolomb(0))
	assert.True(t, bs.WriteSignedExponentialGolomb(1))
	assert.True(t, bs.WriteSignedExponentialGolomb(-1))
	assert.True(t, bs.WriteSignedExponentialGolomb(2))
	assert.True(t, bs.WriteSignedExponentialGolomb(-2))
	assert.True(t, bs.WriteSignedExponentialGolomb(3))
	assert.True(t, bs.WriteSignedExponentialGolomb(-3))
	assert.True(t, bs.WriteSignedExponentialGolomb(4))
	assert.True(t, bs.WriteSignedExponentialGolomb(-4))
	assert.True(t, bs.WriteSignedExponentialGolomb(5))
	assert.True(t, bs.WriteSignedExponentialGolomb(-5))

	assert.True(t, bs.Seek(0, 0))

	sexpg, ok := bs.ReadSignedExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, int64(0), sexpg)
	sexpg, ok = bs.ReadSignedExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, int64(1), sexpg)
	sexpg, ok = bs.ReadSignedExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, int64(-1), sexpg)
	sexpg, ok = bs.ReadSignedExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, int64(2), sexpg)
	sexpg, ok = bs.ReadSignedExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, int64(-2), sexpg)
	sexpg, ok = bs.ReadSignedExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, int64(3), sexpg)
	sexpg, ok = bs.ReadSignedExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, int64(-3), sexpg)
	sexpg, ok = bs.ReadSignedExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, int64(4), sexpg)
	sexpg, ok = bs.ReadSignedExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, int64(-4), sexpg)
	sexpg, ok = bs.ReadSignedExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, int64(5), sexpg)
	sexpg, ok = bs.ReadSignedExponentialGolomb()
	assert.True(t, ok)
	assert.Equal(t, int64(-5), sexpg)

	sexpg, ok = bs.ReadSignedExponentialGolomb()
	assert.False(t, ok)
	assert.Equal(t, int64(0), sexpg)
}
