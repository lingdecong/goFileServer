package main

import (
	"fmt"
	"net"
)

func ServerStop() {
	close(stopChan)
	wg.Wait()
}

func StartLoop(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		wg.Add(1)
		go handleClient(conn)
	}
}
