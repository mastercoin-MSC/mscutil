package mscutil

import (
	"log"
	"os/user"
	"os"
	"path"
)

type LogSystem struct {
	log *log.Logger
}

func NewLogSystem() *LogSystem {
	user, _ := user.Current()
	file, _ := os.OpenFile(path.Join(user.HomeDir, "mscd.log"), os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModePerm)

	return &LogSystem{log: log.New(file, "[MSCD] ", log.LstdFlags)}
}

func (log *LogSystem) Pprintln(v ...interface{}) {
	log.log.Println(v...)
}
func (log *LogSystem) Println(v ...interface{}) {
	log.log.Println(v...)
}

func (log *LogSystem) Fatal(v ...interface{}) {
	log.log.Fatal(v...)
}

var Logger *LogSystem = NewLogSystem()
