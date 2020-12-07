package codec

import "io"

//Header 请求头
type Header struct {
	ServiceMethod string // 服务名和方法
	Seq           uint64 //请求的序号
	Error         string //错误信息
}

//Codec 抽象成接口
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

//类似工厂模式，但不存实例，存构造函数
var NewCodecFuncMap map[Type]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)
	//默认gob编码
	NewCodecFuncMap[GobType] = NewGobCodec
}
