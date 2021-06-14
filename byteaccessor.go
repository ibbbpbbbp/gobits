package gobits

type ByteAccessor interface {
	At(byteOffset int64) (b byte, ok bool)
	Slice(byteOffset, length int64) []byte
	Put(bytes []byte, byteOffset int64) bool
	Length() int64
}
