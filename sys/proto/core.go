package proto

import (
	"bufio"
	"reflect"
)

type (
	IMessage interface {
		GetComId() uint16
		GetMsgId() uint16
		GetMsgLen() uint32
		GetBuffer() []byte
		ToStriing() string
	}
	IMessageFactory interface {
		SetMessageConfig(MsgProtoType ProtoType, IsUseBigEndian bool)
		DecodeMessageBybufio(r *bufio.Reader) (message IMessage, err error)
		DecodeMessageBybytes(buffer []byte) (message IMessage, err error)
		EncodeToMesage(comId uint16, msgId uint16, msg interface{}) (message IMessage)
		EncodeToByte(message *[]byte) (buffer []byte)
		RpcEncodeMessage(d interface{}) ([]byte, error)
		RpcDecodeMessage(dataType reflect.Type, d []byte) (interface{}, error)
	}
	IProto interface {
		DecodeMessageBybufio(r *bufio.Reader) (message IMessage, err error)
		DecodeMessageBybytes(buffer []byte) (message IMessage, err error)
		EncodeToMesage(comId uint16, msgId uint16, msg interface{}) (message IMessage)
		EncodeToByte(message *[]byte) (buffer []byte)
		ByteDecodeToStruct(t reflect.Type, d []byte) (interface{}, error)
	}
)

var (
	defsys IProto
)

func OnInit(config map[string]interface{}, option ...Option) (err error) {
	//log.Warnf("BB%v:", config)
	defsys, err = newSys(newOptions(config, option...))
	return
}

func NewSys(option ...Option) (sys IProto, err error) {
	sys, err = newSys(newOptionsByOption(option...))
	return
}

func DecodeMessageBybufio(r *bufio.Reader) (IMessage, error) {
	return defsys.DecodeMessageBybufio(r)
}
func DecodeMessageBybytes(buffer []byte) (msg IMessage, err error) {
	return defsys.DecodeMessageBybytes(buffer)
}
func EncodeToMesage(comId uint16, msgId uint16, msg interface{}) (message IMessage) {
	return defsys.EncodeToMesage(comId, msgId, msg)
}
func EncodeToByte(message *[]byte) (buffer []byte) {
	return defsys.EncodeToByte(message)
}

func ByteDecodeToStruct(t reflect.Type, d []byte) (interface{}, error) {
	return defsys.ByteDecodeToStruct(t, d)
}
