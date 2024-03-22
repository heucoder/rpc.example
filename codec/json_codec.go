package codec

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

//如果不实现该接口，就会报错
var _ Codec = (*JsonCodec)(nil)

type JsonCodec struct {
	conn io.ReadWriteCloser
	buf  *bufio.Writer
	enc  json.Encoder
	dec  json.Decoder
}

func NewJsonCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &JsonCodec{
		conn: conn,
		buf:  buf,
		enc:  *json.NewEncoder(buf),
		dec:  *json.NewDecoder(conn),
	}
}

func (j *JsonCodec) ReadHeader(head *Header) error {
	return j.dec.Decode(head)
}

func (j *JsonCodec) ReadBody(body interface{}) error {
	return j.dec.Decode(body)
}

func (j *JsonCodec) Write(head *Header, body interface{}) (err error) {

	defer func() {
		_ = j.buf.Flush()
		if err != nil {
			_ = j.Close()
		}
	}()

	err = j.enc.Encode(head)
	if err != nil {
		fmt.Printf("[JsonCodec.Write] Encode head:%v", head)
		return err
	}
	err = j.enc.Encode(body)
	if err != nil {
		fmt.Printf("[JsonCodec.Write] Encode body:%v", body)
		return err
	}
	return nil
}

func (c *JsonCodec) Close() error {
	return c.conn.Close()
}
