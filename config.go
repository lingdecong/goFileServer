package main

import (
	"os"

	"github.com/Unknwon/goconfig"
)

type appConf struct {
	DstDir   string
	Port     int
	IPAddr   string
	DeadTime int64
}

func LoadConf() {
	cfgFile, err := goconfig.LoadConfigFile("./conf.ini")
	if err != nil {
		os.Exit(-1)
	}
	conf.IPAddr = cfgFile.MustValue("app", "ip_addr", "127.0.0.1")
	conf.DstDir = cfgFile.MustValue("app", "dst_dir", "/tmp")
	conf.Port = int(cfgFile.MustInt("app", "port", 9099))
	conf.DeadTime = int64(cfgFile.MustInt("app", "deadtime", 60))
}
