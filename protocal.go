package main

// Message type
const (
	FileName = uint16(iota)
	Md5
	StatCode
	File
)

// Message stat code
const (
	TransportOK    uint16 = 200
	TransportError uint16 = 400

	FileExist   uint16 = 304
	RequestFile uint16 = 201

	Md5Error uint16 = 500
)

type Packet interface {
	Pack() []byte
}

// Message format
type Message struct {
	Len  uint32
	Type uint16
	Data []byte
}

func NewMessage(msgType uint16, length uint32, data []byte) *Message {
	return &Message{Type: msgType, Len: length, Data: data}
}

func (m *Message) Pack() []byte {
	var bufMsg []byte = make([]byte, 0)
	bufMsg = append(bufMsg, Uint32ToBytes(m.Len)...)
	bufMsg = append(bufMsg, Uint16ToBytes(m.Type)...)
	bufMsg = append(bufMsg, m.Data...)
	return bufMsg
}
