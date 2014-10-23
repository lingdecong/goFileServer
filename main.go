package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"github.com/lingdecong/golib"
)

var (
	conf     appConf
	wg       sync.WaitGroup
	stopChan chan int
)

func init() {
	stopChan = make(chan int)
	LoadConf()
}

func main() {
	fmt.Println(conf)
	golib.CheckDir(conf.DstDir, 0775)
	port := strconv.Itoa(conf.Port)
	service := conf.IPAddr + ":" + port
	laddr, err := net.ResolveTCPAddr("tcp4", service)
	if err != nil {
		fmt.Println(err)
	}

	listener, err := net.ListenTCP("tcp4", laddr)
	if err != nil {
		fmt.Println(err)
	}

	go StartLoop(listener)

	// 处理信号
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	fmt.Printf("Signal:%v, the server stoping...\n", <-ch)
	ServerStop()
}
