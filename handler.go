package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	//"sync"
	"time"

	"github.com/lingdecong/golib"
)

func handleClient(conn net.Conn) {
	defer func() {
		fmt.Println("conn exit...")
	}()
	defer wg.Done()
	defer conn.Close()
	fmt.Println(conn.RemoteAddr())

	recvMessageChan := make(chan *Message, 100)
	sendMessageChan := make(chan *Message, 100)
	doneCh := make(chan int)
	go handleRecvMessage(recvMessageChan, sendMessageChan, doneCh)
	go handleSendMessage(conn, sendMessageChan)
	request := make([]byte, 1024)
	buf := make([]byte, 0)
	for {
		select {
		case <-doneCh:
			//close(recvMessageChan)
			break
			//fmt.Println("handler is existing...")
			return
		case <-stopChan:
			//fmt.Println("step 1-4")
			//close(recvMessageChan)
			break
			//fmt.Println("step 1-5")
			return
		default:
			//fmt.Println("step 1-3")
		}
		//fmt.Println("step 2")
		conn.SetReadDeadline(time.Now().Add(time.Duration(conf.DeadTime) * time.Second))
		readSize, err := conn.Read(request)
		//fmt.Println(readSize)
		if err != nil {
			fmt.Println(err)
			if err == io.EOF {
				fmt.Println("eof exit...")
				//close(recvMessageChan)
				break
			}
			continue
		}

		buf = append(buf, request[:readSize]...)
		bufSize := uint32(len(buf))

		// 此处需要错误处理
		//fmt.Println("step 3")
		for {
			if bufSize > 6 {
				msgLen := BytesToUint32(buf[:4])
				msgType := BytesToUint16(buf[4:6])
				if msgLen <= bufSize {
					data := buf[6:msgLen]
					// 发送数据
					//fmt.Println("step 4")
					recvMessageChan <- NewMessage(msgType, msgLen, data)
					// 处理buf
					buf = buf[msgLen:]
					bufSize = bufSize - msgLen
				} else {
					break
				}
			} else {
				break
			}
		}
	}
	close(recvMessageChan)
}

func handleSendMessage(conn net.Conn, sendMessageChan chan *Message) {
	//defer doneWg.Done()
	defer func() {
		fmt.Println("send over")
	}()
	for {
		m, ok := <-sendMessageChan
		//fmt.Println("send", BytesToUint16(m.Data))

		if ok == false {
			break
		}
		conn.SetWriteDeadline(time.Now().Add(time.Duration(conf.DeadTime) * time.Second))
		_, err := conn.Write(m.Pack())
		if err != nil {
			fmt.Println(err)
		}
	}
}

func handleRecvMessage(recvMessageChan chan *Message, sendMessageChan chan *Message, doneChan chan int) {
	defer func() {
		fmt.Println("recv over")
		//doneChan <- 1
	}()
	//defer doneWg.Done()
	defer close(sendMessageChan)
	var fileName string
	var tmpFileName string
	var clientFileMd5 string
	var fileFd *os.File = nil
	var wr *bufio.Writer = nil
	for {
		m, ok := <-recvMessageChan
		//fmt.Println(m)
		if ok == false {
			break
		}
		switch m.Type {
		case FileName:
			fileName = path.Join(conf.DstDir, string(m.Data))
			tmpFileName = conf.DstDir + "/." + string(m.Data)
			fmt.Println("fffffff", fileName, "tmp: ", tmpFileName)

		case Md5:
			// 文件已经存在
			var data []byte
			clientFileMd5 = string(m.Data)
			if golib.IsExists(fileName) {
				fileMd5, err := FileMd5(fileName)
				if err != nil {
					fmt.Println(err)
					// 此处是否需要返回err，还需要思考一下
					data = Uint16ToBytes(TransportError)
					sendMessageChan <- NewMessage(StatCode, uint32(6+len(data)), data)
					break
				}
				if clientFileMd5 == fileMd5 {
					// 发送消息告知客户端该文件已经存在
					data = Uint16ToBytes(FileExist)
					sendMessageChan <- NewMessage(StatCode, uint32(6+len(data)), data)
					break
				}
			}
			// 需要发送文件
			data = Uint16ToBytes(RequestFile)
			sendMessageChan <- NewMessage(StatCode, uint32(6+len(data)), Uint16ToBytes(RequestFile))

		case File:
			if fileFd == nil {
				fd, err := os.OpenFile(tmpFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 775)
				if err != nil {
					fmt.Println(err)
				}
				fileFd = fd
				wr = bufio.NewWriter(fd)
			}
			_, err := wr.Write(m.Data)
			wr.Flush()
			if err != nil {
				fmt.Println("hello:", err)
				// 删除文件
				var data []byte
				data = Uint16ToBytes(TransportError)
				sendMessageChan <- NewMessage(StatCode, uint32(6+len(data)), data)
				os.Remove(fileName)
				break
			}

		case StatCode:
			var data []byte
			switch BytesToUint16(m.Data) {
			case TransportOK:
				// 对比文件MD5
				if fileFd != nil {
					wr.Flush()
					fileFd.Close()
				}
				servFileMd5, err := FileMd5(tmpFileName)
				fmt.Println("md5", servFileMd5, "cmd5", clientFileMd5)
				if err != nil {
					fmt.Println(err)
					data = Uint16ToBytes(Md5Error)
					sendMessageChan <- NewMessage(StatCode, uint32(6+len(data)), data)
					break
				}
				if servFileMd5 == clientFileMd5 {
					data = Uint16ToBytes(TransportOK)
					sendMessageChan <- NewMessage(StatCode, uint32(6+len(data)), data)
					os.Rename(tmpFileName, fileName)
					fmt.Println("yyy")
					// 成功
					break
				} else {
					fmt.Println("zzz")
					//os.Exit(1)
					data = Uint16ToBytes(TransportError)
					fmt.Println(BytesToUint16(data))
					sendMessageChan <- NewMessage(StatCode, uint32(6+len(data)), data)
					os.Remove(tmpFileName)
					break
				}

			case TransportError:
				// 删除文件
				os.Remove(fileName)
				break

			default:
				fmt.Println("Unknow message type")
				break
			}
		}
	}

}
