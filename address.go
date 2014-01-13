package mscutil

import (
  "github.com/conformal/btcutil"
  "github.com/conformal/btcwire"
  "github.com/conformal/btcscript"
)

const ExodusAddress string = "1EXoDusjGwvnjZUyKkxZ4UHEf77z6A5S4P"

func GetExodusTransactions(block *btcutil.Block) []*btcutil.Tx {
  var txs []*btcutil.Tx

  for _, tx := range txs {
    mtx := tx.MsgTx()
    for _, txOut := range mtx.TxOut {
      // Extract the address from the script pub key
      _, addrs, _, _ := btcscript.ExtractPkScriptAddrs(txOut.PkScript, btcwire.MainNet)
      // Check each output address and if there's an address going to the exodus address
      // we add it to tx slice
      for _, addr := range addrs {
        if addr.EncodeAddress() == ExodusAddress {
          txs = append(txs, tx)
          // Continue, we don't care if there are more exodus addresses
          continue
        }
      }
    }
  }

  return txs
}

