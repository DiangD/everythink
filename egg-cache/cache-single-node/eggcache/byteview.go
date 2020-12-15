package eggcache

type ByteView struct {
	bytes []byte
}

func (v ByteView) Len() int {
	return len(v.bytes)
}

func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.bytes)
}

func (v ByteView) String() string {
	return string(v.bytes)
}

func cloneBytes(src []byte) []byte {
	tmp := make([]byte, len(src))
	copy(tmp, src)
	return tmp
}
