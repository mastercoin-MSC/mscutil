package mscutil

import (
	"os/user"
	"bytes"
	"path"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"encoding/binary"
	"encoding/gob"
)

type LDBDatabase struct {
	db *leveldb.DB
}

func (db *LDBDatabase) GetDb() *leveldb.DB{
	return db.db
}

func NewLDBDatabase(name string) (*LDBDatabase, error) {
	// This will eventually have to be something like a resource folder.
	// it works on my system for now. Probably won't work on Windows
	usr, _ := user.Current()
	dbPath := path.Join(usr.HomeDir, ".mastercoin", name)

	// Open the db
	db, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		return nil, err
	}

	database := &LDBDatabase{db: db}

	return database, nil
}
func (db *LDBDatabase) Close() {
	db.db.Close()
}

func (db *LDBDatabase) PutMap(key []byte, value interface{}){
	b := new(bytes.Buffer)
	e := gob.NewEncoder(b)

	err := e.Encode(value)
	if err != nil {
		panic(err)
	}

	db.Put(key, b.Bytes())
}

func (db *LDBDatabase) PutAccount(addr string, sheet map[uint32]uint64){
	db.PutMap([]byte(addr), sheet)
	return
}

func (db *LDBDatabase) GetAccount(addr string) map[uint32]uint64 {
	var account map[uint32]uint64

	db.GetMap([]byte(addr), &account)
	if len(account) == 0 {
		account = map[uint32]uint64{1: uint64(0), 2: uint64(0)}
	}

	return account
}

func (db *LDBDatabase) GetMap(key []byte, value interface{}){
	rawData, err := db.Get(key)

	if err == nil && len(rawData) > 0{
		b := bytes.NewBuffer(rawData)
		e := gob.NewDecoder(b)

		err := e.Decode(value)
		if err != nil {
			fmt.Println("Could not get/decode key", err.Error())
		}
	}
}

func (db *LDBDatabase) Put(key []byte, value []byte) {
	err := db.db.Put(key, value, nil)
	if err != nil {
		fmt.Println("Error put", err)
	}
}

func (db *LDBDatabase) Get(key []byte) ([]byte, error) {
	return db.db.Get(key, nil)
}

func (db *LDBDatabase) CreateTxPack(key int64, pack []byte) {
	buff := new(bytes.Buffer)
	binary.Write(buff, binary.BigEndian, key)

	db.Put(buff.Bytes(), pack)
}

func (db *LDBDatabase) GetTxPack(key int64) []byte {
	buff := new(bytes.Buffer)
	binary.Write(buff, binary.BigEndian, key)

	data, _ := db.Get(buff.Bytes())
	return data
}
