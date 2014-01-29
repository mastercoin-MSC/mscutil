package mscutil

import (
	"github.com/conformal/btcscript"
	"github.com/conformal/btcutil"
	"github.com/conformal/btcwire"
	"github.com/conformal/btcdb"
)

const ExodusAddress string = "1EXoDusjGwvnjZUyKkxZ4UHEf77z6A5S4P"

const (
	FundraiserEndBlock    = int64(255365)
	FundraiserEndTime     = int64(1377993600) // For bonus generation, not Development coins
	DevelopmentCoinsStart = int64(1377993874) // At this time we start Development coins generation
)

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
				if addr.Addr == ExodusAddress {
					txs = append(txs, tx)
					// Continue, we don't care if there are more exodus addresses
					continue
				}
			}
		}
	}

	return txs
}

type Address struct {
	Addr string
	Raw  []byte
}

func GetAddrs(pkScript []byte) (ret []Address, scriptClass btcscript.ScriptClass) {
	// Extract the address from the script pub key
	scriptClass, addrs, _, _ := btcscript.ExtractPkScriptAddrs(pkScript, btcwire.MainNet)
	// Check each output address and if there's an address going to the exodus address
	// we add it to tx slice
	for _, addr := range addrs {
		// Script address returns the public key if it's a multi sig
		ret = append(ret, Address{Addr: addr.EncodeAddress(), Raw: addr.ScriptAddress()})
	}

	return
}

/*
// Generates an address from a input script signature
func GetAddressFromScriptSig(scriptSig []byte) (string, error) {
  pubkey, _ := hex.DecodeString("040df5ef88d24e2414ad47c9a59a367c96120ab7c5f13a0683e243b1a0747ebd2a740d3eec1f7bd0cf17b85c0e5aa8801a7400eda229f3e0e40e40c0313d6ab5a8")
  fmt.Println(pubkey)

  a, err := btcutil.NewAddressPubKey(pubkey, btcwire.MainNet)
  if err != nil {
    return err
  }

  return a.EncodeAddress()
}
*/

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
func FindSender(txIns []*btcwire.TxIn, btcdb btcdb.Db) (Address, error) {
	inputs := make(map[string]int64)

	for _, txIn := range txIns {
		op := txIn.PreviousOutpoint
		hash := op.Hash
		index := op.Index
		transactions, err := btcdb.FetchTxBySha(&hash)
		if err != nil {
			return Address{}, err
		}

		previousOutput := transactions[0].Tx.TxOut[index]

		// The largest contributor receives the Mastercoins, so add multiple address values together
		address, _ := GetAddrs(previousOutput.PkScript)
		inputs[address[0].Addr] += previousOutput.Value
	}

	// Decide which input has the most value so we know who is sending this transaction
	var highest int64
	var highestAddress string

	for k, v := range inputs {
		if v > highest {
			highest = v
			highestAddress = k
		}
	}
      return Address{Addr: highestAddress}, nil
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
