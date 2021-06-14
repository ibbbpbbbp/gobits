package gobits

type SliceByteAccessor struct {
	bytes []byte
}

func (ba *SliceByteAccessor) At(byteOffset int64) (byte, bool) {
	if byteOffset < 0 || int64(len(ba.bytes)) <= byteOffset {
		return 0, false
	}
	return ba.bytes[byteOffset], true
}

func (ba *SliceByteAccessor) Slice(byteOffset, length int64) []byte {
	if byteOffset < 0 || length <= 0 {
		return []byte{}
	}

	bytesLen := int64(len(ba.bytes))
	last := byteOffset + length
	if bytesLen < last {
		length -= last - bytesLen
		last = bytesLen
	}

	if length == 0 {
		return []byte{}
	}

	bytes := make([]byte, length)
	copy(bytes, ba.bytes[byteOffset:last])
	return bytes
}

func (ba *SliceByteAccessor) Put(bytes []byte, byteOffset int64) bool {
	if byteOffset < 0 || bytes == nil {
		return false
	}

	copy(ba.bytes[byteOffset:], bytes)
	return true
}

func (ba *SliceByteAccessor) Length() int64 {
	return int64(len(ba.bytes))
}

func NewSliceByteAccessor(bytes []byte) *SliceByteAccessor {
	return &SliceByteAccessor{bytes: bytes}
}
