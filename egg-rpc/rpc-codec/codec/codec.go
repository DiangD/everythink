package codec

import "io"

type Header struct {
	ServiceMethod string // 服务名和方法
	Seq           uint64 //请求的序号
	Error         string //错误信息
}

type Codec interface {
	io.Closer
	ReadHeader(h *Header) error
	ReadBody(obj interface{}) error
	Write(h *Header, obj interface{}) error
}

type NewCodecFunc func(closer io.ReadWriteCloser) Codec

type Type string

const (
	GobType  Type = "application/gob"
	JsonType Type = "application/json"
)

var NewCodecFuncMap map[Type]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCodec
}
