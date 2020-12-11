package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

type GobCodec struct {
	conn io.ReadWriteCloser
	buf  *bufio.Writer
	dec  *gob.Decoder
	enc  *gob.Encoder
}

func NewGobCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn: conn,
		buf:  buf,
		dec:  gob.NewDecoder(conn),
		enc:  gob.NewEncoder(buf),
	}
}

//Close 关闭io
func (c *GobCodec) Close() error {
	return c.conn.Close()
}

//ReadHeader 读取header
func (c *GobCodec) ReadHeader(h *Header) error {
	return c.dec.Decode(h)
}

//ReadBody 读取body
func (c *GobCodec) ReadBody(obj interface{}) error {
	return c.dec.Decode(obj)
}

//Write 写入到header、body
func (c *GobCodec) Write(h *Header, obj interface{}) error {
	defer func() {
		err := c.buf.Flush()
		if err != nil {
			_ = c.Close()
		}
	}()

	if err := c.enc.Encode(h); err != nil {
		log.Println("rpc: gob error encoding header:", err)
		return err
	}

	if err := c.enc.Encode(obj); err != nil {
		log.Println("rpc: gob error encoding body:", err)
		return err
	}

	return nil
}
