package mscutil

import (
	"github.com/conformal/btcscript"
	"github.com/conformal/btcutil"
)

// Helper function
func getAddrsClassA(tx *btcutil.Tx) []string {
	var addrs []string

	mtx := tx.MsgTx()
	for _, txOut := range mtx.TxOut {
		a, _ := GetAddrs(txOut.PkScript)
		addrs = append(addrs, a...) // a is one address guaranteed
	}

	return addrs
}

type TxType byte

const (
	InvalidType = iota
	ClassAType
	ClassBType
)

var txTypeToString = []string{
	InvalidType: "Invalid",
	ClassAType:  "ClassA",
	ClassBType:  "ClassB",
}

func (m TxType) String() string {
	if int(m) > len(txTypeToString) && int(m) < 0 {
		return "Invalid type"
	}

	return txTypeToString[m]
}

type Tx struct {
	Type     TxType
	Receiver string
	Data     SimpleSend
}

func isClassA(tx *btcutil.Tx) bool {
	mtx := tx.MsgTx()
	for _, txOut := range mtx.TxOut {
		_, scriptType := GetAddrs(txOut.PkScript)
		if scriptType == btcscript.MultiSigTy {
			return false
		}
	}

	// If it wasn't multi sig it's class a
	return true
}

func isClassB(tx *btcutil.Tx) bool {
	return !isClassA(tx)
}

func getTxType(tx *btcutil.Tx) TxType {
	if isClassA(tx) {
		return ClassAType
	}

	return ClassBType
}

func MakeClassATx(addrs []string) *Tx {
	var data SimpleSend
	// Find the data address
	for _, val := range addrs {
		simpleSend := DecodeFromAddress(val)
		if simpleSend.TransactionType == 0 && (simpleSend.CurrencyId == 2 || simpleSend.CurrencyId == 1) {
			data = simpleSend
			break
		}
	}

	var receiver string
	for _, addr := range addrs {
		if addr[0]+1 == data.Sequence {
			receiver = addr
		}
	}

	return &Tx{
		Type:     ClassAType,
		Receiver: receiver,
		Data:     data,
	}
}

// Creates a Mastercoin Tx out of a regular bitcoin tx
func MakeTx(tx *btcutil.Tx) *Tx {
	txType := getTxType(tx)

	switch txType {
	case ClassAType:
		return MakeClassATx(getAddrsClassA(tx))
	case ClassBType:
	}

	return nil
}
