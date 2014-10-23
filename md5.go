package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

func FileMd5(fileName string) (string, error) {
	file, _ := os.Open(fileName)
	fmt.Println(fileName)
	rd := bufio.NewReader(file)
	md5h := md5.New()
	for {
		line, err := rd.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				io.WriteString(md5h, line)
				break
			} else {
				fmt.Println(err)
			}
		}
		io.WriteString(md5h, line)
	}
	return hex.EncodeToString(md5h.Sum(nil)), nil
}
