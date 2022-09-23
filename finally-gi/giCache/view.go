package giCache

type ByteView struct {
	b []byte
}

func (b ByteView) Len() int {
	return len(b.b)
}

func (b ByteView) String() string {
	return string(b.b)
}

func (b ByteView) ByteSlice() []byte {
	return copyByte(b.b)
}

func copyByte(b []byte) []byte {
	data := make([]byte, len(b))
	copy(data, b)
	return data
}
