package mscutil

import (
	"bytes"
	"encoding/binary"
	"github.com/conformal/btcutil"
)

type TxPack struct {
	Time   int64
	Height int64
	Txs    []*btcutil.Tx
}

//|time|height |len(txs)|len(tx)|tx bytes|len(tx)|tx bytes|
func (pack *TxPack) Deserialize(data []byte) error {
	buff := bytes.NewBuffer(data) 

	var err error
	err = binary.Read(buff, binary.LittleEndian, &pack.Time)

	if err != nil {
		return err
	}

	err = binary.Read(buff, binary.LittleEndian, &pack.Height)

	if err != nil {
		return err
	}

	var txSize uint32
	err = binary.Read(buff, binary.LittleEndian, &txSize)
	if err != nil {
		return err
	}

	pack.Txs = make([]*btcutil.Tx, txSize)

	// Loop for each transaction
	for i := uint32(0); i < txSize; i++ {
		// Get the length of the transaction in bytes
		var byteLength uint32
		err = binary.Read(buff, binary.LittleEndian, &byteLength)
		if err != nil {
			return err
		}
		// Make the buffer equal to the amount written (= len(tx)
		txBuff := make([]byte, byteLength)
		// Read the tx
		err = binary.Read(buff, binary.LittleEndian, &txBuff)
		if err != nil {
			return err
		}
		// Create a new transaction
		pack.Txs[i], err = btcutil.NewTxFromBytes(txBuff)
		if err != nil {
			return err
		}
	}

	return nil
}

func (pack *TxPack) Serialize() ([]byte, error) {
	var buff bytes.Buffer
	var err error
	err = binary.Write(&buff, binary.LittleEndian, pack.Time)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buff, binary.LittleEndian, pack.Height)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buff, binary.LittleEndian, uint32(len(pack.Txs)))
	if err != nil {
		return nil, err
	}

	for _, tx := range pack.Txs {
		var b bytes.Buffer
		tx.MsgTx().Serialize(&b)

		err = binary.Write(&buff, binary.LittleEndian, uint32(b.Len()))
		if err != nil {
			return nil, err
		}

		err = binary.Write(&buff, binary.LittleEndian, b.Bytes())
		if err != nil {
			return nil, err
		}

	}
	return buff.Bytes(), nil
}
