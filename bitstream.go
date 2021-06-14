package gobits

import (
	"encoding/binary"
	"math"
)

type pos struct {
	byteOffset int64
	bitOffset  byte
}

type PosWrapper struct {
	pos
}

type BitStream struct {
	ba ByteAccessor
	pos
}

func lowerBits(byt, count byte) byte {
	return byt & ((1 << count) - 1)
}

func higherBits(byt, count byte) byte {
	shift := 8 - count
	mask := byte(0xff) << shift
	return (byt & mask) >> shift
}

func highestByte(val uint64) byte {
	return byte(val >> 56)
}

func countEffectiveBits(val uint64) byte {
	bitCount := byte(0)
	for ; val != 0; bitCount++ {
		val >>= 1
	}
	return bitCount
}

func writePartialByte(srcByte byte, srcBitCount byte, dstByte byte, dstBitOffset byte) byte {
	mask := byte(0xff<<(8-srcBitCount)) >> dstBitOffset
	return (dstByte &^ mask) | (srcByte >> dstBitOffset)
}

func (bs *BitStream) RemainingBits(bitCount int64) bool {
	bitCount += int64(bs.bitOffset)
	byteOffset := bs.byteOffset
	for bitCount > 0 {
		_, ok := bs.ba.At(byteOffset)
		if !ok {
			return false
		}
		byteOffset++
		bitCount -= 8
	}
	return true
}

func (bs *BitStream) PeekBits(bitCount byte) (uint64, bool) {
	if bitCount == 0 {
		return 0, true
	}
	if !bs.RemainingBits(int64(bitCount)) || bitCount > 64 {
		return 0, false
	}

	byteOffset := bs.byteOffset
	remainingBitsInCurrByte := 8 - bs.bitOffset
	byt, ok := bs.ba.At(byteOffset)
	if !ok {
		return 0, false
	}
	bits_ := lowerBits(byt, remainingBitsInCurrByte)
	byteOffset++
	if bitCount < remainingBitsInCurrByte {
		return uint64(higherBits(bits_, bs.bitOffset+bitCount)), true
	}

	bits := uint64(bits_)
	bitCount -= remainingBitsInCurrByte
	for bitCount >= 8 {
		byt, ok := bs.ba.At(byteOffset)
		if !ok {
			return 0, false
		}
		bits = (bits << 8) | uint64(byt)
		byteOffset++
		bitCount -= 8
	}

	if bitCount > 0 {
		byt, ok := bs.ba.At(byteOffset)
		if !ok {
			return 0, false
		}
		bits = (bits << bitCount) | uint64(higherBits(byt, bitCount))
	}

	return bits, true
}

func (bs *BitStream) ConsumeBits(bitCount int64) bool {
	if !bs.RemainingBits(bitCount) {
		return false
	}
	bs.byteOffset += (int64(bs.bitOffset) + bitCount) / 8
	bs.bitOffset = byte((int64(bs.bitOffset) + bitCount) % 8)
	return true
}

func (bs *BitStream) ConsumeBytes(byteCount int64) bool {
	return bs.ConsumeBits(byteCount * 8)
}

func (bs *BitStream) ReadBits(bitCount byte) (uint64, bool) {
	bits, ok := bs.PeekBits(bitCount)
	if !ok {
		return 0, false
	}

	if !bs.ConsumeBits(int64(bitCount)) {
		return 0, false
	}

	return bits, true
}

func (bs *BitStream) ReadUint8() (uint8, bool) {
	v, ok := bs.ReadBits(8)
	return uint8(v), ok
}

func (bs *BitStream) ReadUint16(bo binary.ByteOrder) (uint16, bool) {
	b, ok := bs.ReadBits(16)
	v := []byte{byte((b >> 8) & 0xff), byte(b & 0xff)}
	return bo.Uint16(v), ok
}

func (bs *BitStream) ReadUint32(bo binary.ByteOrder) (uint32, bool) {
	b, ok := bs.ReadBits(32)
	v := []byte{byte((b >> 24) & 0xff), byte((b >> 16) & 0xff), byte((b >> 8) & 0xff), byte(b & 0xff)}
	return bo.Uint32(v), ok
}

func (bs *BitStream) ReadUint64(bo binary.ByteOrder) (uint64, bool) {
	b, ok := bs.ReadBits(64)
	v := []byte{
		byte((b >> 56) & 0xff),
		byte((b >> 48) & 0xff),
		byte((b >> 40) & 0xff),
		byte((b >> 32) & 0xff),
		byte((b >> 24) & 0xff),
		byte((b >> 16) & 0xff),
		byte((b >> 8) & 0xff),
		byte(b & 0xff),
	}
	return bo.Uint64(v), ok
}

func (bs *BitStream) Seek(byteOffset int64, bitOffset byte) bool {
	_, ok := bs.ba.At(byteOffset)
	if ok && bitOffset < 8 {
		bs.byteOffset = byteOffset
		bs.bitOffset = bitOffset
		return true
	}

	return false
}

func (bs *BitStream) SavePos() PosWrapper {
	return PosWrapper{bs.pos}
}

func (bs *BitStream) RestorePos(pw PosWrapper) {
	bs.pos = pw.pos
}

func (bs *BitStream) ResetPos() {
	bs.pos = pos{
		byteOffset: 0,
		bitOffset:  0,
	}
}

func (bs *BitStream) ReadExponentialGolomb() (uint64, bool) {
	originalbyteOffset := bs.byteOffset
	originalBitOffset := bs.bitOffset
	zeroBitCount := 0
	peekedBit := uint64(0)
	ok := true

	for {
		if peekedBit, ok = bs.PeekBits(1); peekedBit != 0 || !ok {
			break
		}
		zeroBitCount++
		bs.ConsumeBits(1)
	}

	val := uint64(0)
	valueBitCount := zeroBitCount + 1

	if !ok {
		goto failed
	} else if valueBitCount > 64 {
		goto failed
	} else if val, ok = bs.ReadBits(byte(valueBitCount)); !ok {
		goto failed
	}

	return val - 1, true

failed:
	bs.Seek(originalbyteOffset, originalBitOffset)
	return 0, false
}

func (bs *BitStream) ReadSignedExponentialGolomb() (int64, bool) {
	val := uint64(0)
	ok := true

	if val, ok = bs.ReadExponentialGolomb(); !ok {
		return 0, false
	}

	if val&1 == 0 {
		return -int64(val / 2), true
	}
	return int64(val+1) / 2, true
}

func (bs *BitStream) WriteBits(val uint64, bitCount byte) bool {
	if bitCount == 0 {
		return true
	}
	if !bs.RemainingBits(int64(bitCount)) || bitCount > 64 {
		return false
	}

	consumeBits := int64(bitCount)
	val <<= 64 - uint64(bitCount)
	remainingBitsInCurrByte := 8 - bs.bitOffset
	bitsInFirstByte := bitCount
	if bitsInFirstByte > remainingBitsInCurrByte {
		bitsInFirstByte = remainingBitsInCurrByte
	}

	dstByteOffset := bs.byteOffset
	dstByte, ok := bs.ba.At(dstByteOffset)
	if !ok {
		return false
	}

	bytes := make([]byte, (bs.bitOffset+bitCount+7)/8)
	bytes[0] = writePartialByte(highestByte(val), bitsInFirstByte, dstByte, bs.bitOffset)

	if bitCount <= remainingBitsInCurrByte {
		goto fin
	}

	val <<= bitsInFirstByte
	bitCount -= bitsInFirstByte
	for bitCount >= 8 {
		dstByteOffset++
		bytes[dstByteOffset-bs.byteOffset] = highestByte(val)
		val <<= 8
		bitCount -= 8
	}

	if bitCount > 0 {
		dstByteOffset++
		dstByte, ok := bs.ba.At(dstByteOffset)
		if !ok {
			return false
		}
		bytes[dstByteOffset-bs.byteOffset] = writePartialByte(highestByte(val), bitCount, dstByte, 0)
	}

fin:
	if !bs.ba.Put(bytes, bs.byteOffset) {
		return false
	}

	return bs.ConsumeBits(consumeBits)
}

func (bs *BitStream) WriteUint8(val uint8) bool {
	return bs.WriteBits(uint64(val), 8)
}

func (bs *BitStream) WriteUint16(val uint16, bo binary.ByteOrder) bool {
	v := []byte{byte((val >> 8) & 0xff), byte(val & 0xff)}
	return bs.WriteBits(uint64(bo.Uint16(v)), 16)
}

func (bs *BitStream) WriteUint32(val uint32, bo binary.ByteOrder) bool {
	v := []byte{byte((val >> 24) & 0xff), byte((val >> 16) & 0xff), byte((val >> 8) & 0xff), byte(val & 0xff)}
	return bs.WriteBits(uint64(bo.Uint32(v)), 32)
}

func (bs *BitStream) WriteUint64(val uint64, bo binary.ByteOrder) bool {
	v := []byte{
		byte((val >> 56) & 0xff),
		byte((val >> 48) & 0xff),
		byte((val >> 40) & 0xff),
		byte((val >> 32) & 0xff),
		byte((val >> 24) & 0xff),
		byte((val >> 16) & 0xff),
		byte((val >> 8) & 0xff),
		byte(val & 0xff),
	}
	return bs.WriteBits(uint64(bo.Uint64(v)), 64)
}

func (bs *BitStream) WriteExponentialGolomb(val uint64) bool {
	if val == math.MaxUint64 {
		return false
	}
	val++
	return bs.WriteBits(val, countEffectiveBits(val)*2-1)
}

func (bs *BitStream) WriteSignedExponentialGolomb(val int64) bool {
	if val == 0 {
		return bs.WriteExponentialGolomb(0)
	} else if val > 0 {
		return bs.WriteExponentialGolomb(uint64(val)*2 - 1)
	} else {
		if val == math.MinInt64 {
			return false
		}
		return bs.WriteExponentialGolomb(uint64(-val) * 2)
	}
}

func NewBitStream(ba ByteAccessor) *BitStream {
	return &BitStream{
		ba: ba,
		pos: pos{
			byteOffset: 0,
			bitOffset:  0,
		},
	}
}
