package codec

const MagicNumber = 0x3bef5c

type Option struct {
	MagicNumber int  // MagicNumber marks this's a geerpc request
	CodecType   Type // client may choose different Codec to encode body
}

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   JsonType,
}
