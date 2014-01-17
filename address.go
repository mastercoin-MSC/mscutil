package mscutil

import (
	"github.com/conformal/btcscript"
	"github.com/conformal/btcutil"
	"github.com/conformal/btcwire"
)

const ExodusAddress string = "1EXoDusjGwvnjZUyKkxZ4UHEf77z6A5S4P"

func GetExodusTransactions(block *btcutil.Block) []*btcutil.Tx {
	var txs []*btcutil.Tx

	for _, tx := range block.Transactions() {
		mtx := tx.MsgTx()
		for _, txOut := range mtx.TxOut {
			// Extract the address from the script pub key
			addrs, _ := GetAddrs(txOut.PkScript)
			// Check each output address and if there's an address going to the exodus address
			// we add it to tx slice
			for _, addr := range addrs {
				if addr == ExodusAddress {
					txs = append(txs, tx)
					// Continue, we don't care if there are more exodus addresses
					continue
				}
			}
		}
	}

	return txs
}

func GetAddrs(pkScript []byte) (ret []string, scriptClass btcscript.ScriptClass) {
	// Extract the address from the script pub key
	scriptClass, addrs, _, _ := btcscript.ExtractPkScriptAddrs(pkScript, btcwire.MainNet)
	// Check each output address and if there's an address going to the exodus address
	// we add it to tx slice
	for _, addr := range addrs {
		ret = append(ret, addr.EncodeAddress())
	}

	return
}

type MsgType byte

const (
	InvalidMsgTy = iota
	TxMsgTy
	DexMsgTy
)

var msgTypeToString = []string{
	InvalidMsgTy: "Invalid",
	TxMsgTy:      "Transaction",
	DexMsgTy:     "Dex",
}

func (m MsgType) String() string {
	if int(m) > len(msgTypeToString) && int(m) < 0 {
		return "Invalid type"
	}

	return msgTypeToString[m]
}

func GetType(tx *btcutil.Tx) (t MsgType) {
	// Defaults to invalid type
	t = InvalidMsgTy

	mtx := tx.MsgTx()
	for _, txOut := range mtx.TxOut {
		_, scriptType := GetAddrs(txOut.PkScript)
		var mt MsgType

		// Check the btc Tx type and determine our own type
		switch scriptType {
		default:
			fallthrough
		case btcscript.MultiSigTy:
			mt = TxMsgTy
		}

		// Set the tx type if it's greater (class b is higher than a)
		if mt > t {
			t = mt
		}
	}

	return
}
