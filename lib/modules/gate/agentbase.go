package gate

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"sync"
	"war-game/libs/gatelogicmodel"
	"war-game/libs/model/baseencryptmodel"

	"github.com/Pororochenzy/lego/sys/log"
	"github.com/Pororochenzy/lego/sys/proto"
	"github.com/Pororochenzy/lego/utils/container/id"
)

// 远程链接代理
type AgentBase struct {
	Module      IGateModule
	Agent       IAgent
	Conn        IConn
	id          string
	ip          string
	closeSignal chan bool
	writeChan   chan *[]byte
	Isclose     bool
	lock        sync.RWMutex
	wg          sync.WaitGroup
	r           *bufio.Reader
	w           *bufio.Writer
	rev_num     int64
	send_num    int64
}

func (this *AgentBase) Id() string {
	return this.id
}
func (this *AgentBase) IP() string {
	return this.ip
}
func (this *AgentBase) RevNum() int64 {
	return this.rev_num
}
func (this *AgentBase) SendNum() int64 {
	return this.send_num
}
func (this *AgentBase) IsClosed() bool {
	return this.Isclose
}
func (this *AgentBase) OnInit(module IGateModule, coon IConn, agent IAgent) (err error) {
	this.Module = module
	this.Agent = agent
	this.Conn = coon
	this.id = id.NewXId()
	this.ip = coon.RemoteAddr().String()
	this.closeSignal = make(chan bool)
	this.writeChan = make(chan *[]byte, 10)
	this.Isclose = false
	this.r = bufio.NewReaderSize(coon, 1<<17) // 128 kb
	this.w = bufio.NewWriterSize(coon, 1<<17) // 128 kb
	this.rev_num = 0
	this.send_num = 0
	this.Module.Connect(this.Agent) //发送链接消息
	return
}
func (this *AgentBase) OnRun() {
	this.wg.Add(1)
	go this.listenwrite()
loop:
	for {
		//读到结构体
		msg, err, pBuffer := this.DecodeMessage(this.r)
		if err != nil {
			log.Errorf("[%s]接收消息异常 err:%s", this.id, err.Error())
			this.OnClose()
			break loop
		}

		this.Agent.OnRecoverByte(msg.ComId, msg.MsgId, msg.MsgLen, pBuffer)
	}
}

func (this *AgentBase) OnRecoverByte(ComId uint16, MsgId uint16, MsgLen uint32, Buffer *[]byte) {

}

// 数组转结构体
func EncodeStrut(cbBuffer []byte) (gatelogicmodel.BaseMessage, error) {
	msg := gatelogicmodel.BaseMessage{}
	var t gatelogicmodel.BaseMessage
	buf := bytes.NewReader(cbBuffer)
	err := binary.Read(buf, binary.BigEndian, &t)

	if err != nil {
		return msg, err
	}

	return t, nil
}

func (this *AgentBase) readUInt64(r *bufio.Reader) (*gatelogicmodel.BaseMessage, error) {
	buf := make([]byte, 8)
	msg := gatelogicmodel.BaseMessage{}
	_, err := io.ReadFull(r, buf[:8])
	if err != nil {
		return &msg, err
	}

	mHead, err := EncodeStrut(buf)

	return &mHead, nil
}

//	"war-game/libs/model/baseencryptmodel"
//
// 读取二进制流到结构体
func (this *AgentBase) DecodeMessage(r *bufio.Reader) (*gatelogicmodel.BaseMessage, error, *[]byte) {
	var err error
	//声明头文件
	msg, err := this.readUInt64(r)
	if err != nil {
		return msg, err, nil
	}

	if msg.MsgLen > (uint32)(baseencryptmodel.SOCKET_TCP_BUFFER+gatelogicmodel.BaseMessageLen) {
		return msg, fmt.Errorf("DecodeMessageBybufio err msg.MsgLen:%d Super long", msg.MsgLen), nil
	}

	bBuffer := make([]byte, msg.MsgLen)

	_, err = io.ReadFull(r, bBuffer)

	return msg, err, &bBuffer
}

func (this *AgentBase) listenwrite() {
	defer this.wg.Done()
loop:
	for {
		select {
		case msg, ok := <-this.writeChan:
			if ok {
				//b := proto.EncodeToByte(msg)
				_, err := this.w.Write(*msg)
				if err != nil {
					this.OnClose()
					break loop
				}
				err = this.w.Flush()
				if err != nil {
					this.OnClose()
					break loop
				}
			} else {
				this.OnClose()
				break loop
			}
		case <-this.closeSignal:
			break loop
		}
	}
}
func (this *AgentBase) OnClose() {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.Isclose {
		return
	}
	this.Isclose = true
	this.closeSignal <- true
	this.Conn.Close()
}

func (this *AgentBase) Destory() {
	this.wg.Wait()
	this.Isclose = true
	close(this.writeChan)
	close(this.closeSignal)
	this.Module.DisConnect(this.Agent) //发送连接断开的事件
}
func (this *AgentBase) WriteMsg(msg *[]byte) error {
	this.lock.RLock()
	defer this.lock.RUnlock()
	if !this.Isclose {
		if msg != nil {
			this.send_num++
			this.writeChan <- msg
			return nil
		} else {
			return fmt.Errorf("异常写入空消息")
		}
	} else {
		return fmt.Errorf("连接已关闭无法写入消息")
	}
}
func (this *AgentBase) OnRecover(msg proto.IMessage) {
	this.rev_num++
}
