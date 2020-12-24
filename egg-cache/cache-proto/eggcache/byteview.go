package eggcache

//ByteView 只读数据结构来表示缓存值
type ByteView struct {
	bytes []byte
}

//实现Value接口
func (v ByteView) Len() int {
	return len(v.bytes)
}

//ByteSlice copy
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.bytes)
}

//String toString
func (v ByteView) String() string {
	return string(v.bytes)
}

//cloneBytes 由于只读我们要对数据进行复制
func cloneBytes(src []byte) []byte {
	tmp := make([]byte, len(src))
	copy(tmp, src)
	return tmp
}
