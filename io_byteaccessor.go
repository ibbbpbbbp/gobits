package gobits

import (
	"io"
)

const (
	maxBufferSize = int64(4 * 1024)
)

type IOByteAccessor struct {
	rwseeker    io.ReadWriteSeeker
	buffer      []byte
	bufferIndex int64
	bufferSize  int64
}

func (ba *IOByteAccessor) renewBuffer(byteOffset int64) bool {
	if ba.bufferIndex <= byteOffset && byteOffset < ba.bufferIndex+ba.bufferSize {
		return true
	}

	newByteOffset := byteOffset - (maxBufferSize / 2)
	if newByteOffset < 0 {
		newByteOffset = 0
	}

	bufferIndex, err := ba.rwseeker.Seek(newByteOffset, 0)
	if err != nil {
		return false
	}

	ba.bufferIndex = bufferIndex
	bufferSize, err := ba.rwseeker.Read(ba.buffer)
	ba.bufferSize = int64(bufferSize)

	ba.buffer = ba.buffer[:ba.bufferSize]

	if err != nil {
		return false
	}

	return true
}

func (ba *IOByteAccessor) At(byteOffset int64) (byte, bool) {
	if !ba.renewBuffer(byteOffset) {
		return 0, false
	}

	if byteOffset < ba.bufferIndex || ba.bufferIndex+ba.bufferSize <= byteOffset {
		return 0, false
	}

	at := byteOffset - ba.bufferIndex
	return ba.buffer[at], true
}

func (ba *IOByteAccessor) Slice(byteOffset, length int64) []byte {
	if length <= 0 {
		return []byte{}
	}

	_, err := ba.rwseeker.Seek(byteOffset, 0)
	if err != nil {
		return []byte{}
	}

	bytes := make([]byte, length)
	actualLength, err := ba.rwseeker.Read(bytes)
	if err != nil {
		return []byte{}
	}

	bytes = bytes[:actualLength]

	return bytes
}

func (ba *IOByteAccessor) Put(bytes []byte, byteOffset int64) bool {
	if byteOffset < 0 || bytes == nil {
		return false
	}
	if len(bytes) == 0 {
		return true
	}

	_, err := ba.rwseeker.Seek(byteOffset, 0)
	if err != nil {
		return false
	}

	actualLength, err := ba.rwseeker.Write(bytes)
	if err != nil {
		return false
	}

	// sync the buffer
	byteStart := byteOffset - ba.bufferIndex
	byteEnd := byteStart + int64(actualLength)
	bufStart := int64(0)
	bufEnd := ba.bufferSize
	byteStartInRange := false
	byteEndInRange := false

	if bufStart <= byteStart && byteStart < bufEnd {
		byteStartInRange = true
	}
	if bufStart < byteEnd && byteEnd <= bufEnd {
		byteEndInRange = true
	}

	/*
		if byteStartInRange && byteEndInRange {
			copy(ba.buffer[byteStart:], bytes)
		} else if byteStartInRange && !byteEndInRange {
			copy(ba.buffer[byteStart:], bytes[:bufEnd])
		} else if !byteStartInRange && byteEndInRange {
			copy(ba.buffer[:], bytes[bufStart-byteStart:])
		}
	*/
	// The code below is the same as the code above.
	if byteStartInRange {
		copy(ba.buffer[byteStart:], bytes)
	} else if byteEndInRange {
		copy(ba.buffer[:], bytes[bufStart-byteStart:])
	}

	return actualLength == len(bytes)
}

func (ba *IOByteAccessor) Length() int64 {
	e, _ := ba.rwseeker.Seek(0, 2)
	return e
}

func (ba *IOByteAccessor) Reset() {
	if int64(len(ba.buffer)) < maxBufferSize {
		ba.buffer = make([]byte, maxBufferSize)
	}
	ba.bufferIndex = 0
	ba.bufferSize = 0
}

func NewIOByteAccessor(rwseeker io.ReadWriteSeeker) *IOByteAccessor {
	return &IOByteAccessor{
		rwseeker: rwseeker,
		buffer:   make([]byte, maxBufferSize),
	}
}
