package mscutil

import (
	"errors"
	"github.com/conformal/btcscript"
	"github.com/conformal/btcutil"
	"log"
        "fmt"
	"crypto/sha256"
)

type Output struct {
	Value int64
	Addr  string
}

// Helper function
func GetAddrsClassA(tx *btcutil.Tx) []Output {
	var addrs []Output

	mtx := tx.MsgTx()
	for _, txOut := range mtx.TxOut {
		a, _ := GetAddrs(txOut.PkScript)

		out := Output{Value: txOut.Value, Addr: a[0].Addr}

		addrs = append(addrs, out) // a is one address guaranteed
	}

	return addrs
}

func FindInOutputs(outs []Output, finder func(output Output) bool) Output {
	for _, out := range outs {
		if finder(out) {
			return out
		}
	}
}

func GetExodus(outs []Output) Output {
	return FindInOutputs(outs, func(output Output) {
		return output.Addr == ExodusAddress
	})
}

func MultipleSha(target []byte, times int) []byte {
	if times == 0 {
		return target
	}

	result := sha256.Sum256(MultipleSha(target, times - 1))

	return []byte(strings.ToUpper(hex.EncodeToString(result[:])))
}

// Transformers obfuscated public keys in clear text keys
func DeobfuscatePublicKeys(multiSig []Output, receiver string) []string {
	data := make([]string, len(multiSig)-1)
	// For each public key create a hash out of the reciever
	for i, sig := range multiSig[1:] {
		// Hash receiver x times
		hash := MultipleSha([]byte(receiver.Addr), i)
		// xor each byte
		for j, val := range hash[:31] {
			// XOR First byte and last byte are ignored
			// SIG: 02|1C9A3DE5C2E22BF89B1E41E6FED84FB502F8A0C3AE14394A59366293DD130C|33
			// OBV:    AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
			data[i][j] = val ^ sig.Raw[j+1]
		}
	}
	return append(data, receiver)
}

func GetAddrsClassB(tx *btcutil.Tx) []string {
	var addrs []Output

	// TODO if in the future we require multiple multisigs change this to a slice
	var multiSig []Address
	var multiSigVal int64

	tmx := tx.MsgTx()
	// Gather outputs from this tx, filter out the MultiSig
	for _, txOut := range tmx.TxOut {
		a, scriptType := GetAddrs(txOut.PkScript)
		// Assign the multi sig so we can work with it later on
		if scriptType == btcscript.MultiSigTy {
			multiSig = a
			multiSigVal = txOut.Value
		} else {
			// Create regular outs for non multi sigs outputs
			addrs = append(addrs, Output{Value: txOut.Value, Addr: a[0].Addr})
		}
	}


	// Get the exodus address
	exodus := GetExodus(addrs)
	receiver := FindInOutputs(addrs, func(output Output) {
		return output.Value == exodus.Value
	})

	// If the receiver is empty
	if receiver == Output{} {
		return nil, errors.New("Unable to find recipient")
	}

	return DeobfuscatePublicKeys(multiSig, receiver)

}

type SimpleTransaction struct {
	Receiver string
	Data     *SimpleSend
}


// Create simple transaction out of class B outputs
func MakeClassBSimpleSend(outputs []string) (*SimpleTransaction, error) {
	log.Println("Making class b simple transaction")

	// Receiver is the last in the outputs
	receiver := outputs[len(outputs)-1]
	// First X is the data
	data := outputs[:len(outputs)-1]

	// XXX if in the future simple sends may contain multiple data strings, update.
	simpleSend := DecodeFromAddress(data[0])
	if !(simpleSend.TransactionType == 0 && (simpleSend.CurrencyId == 2 || simpleSend.CurrencyId == 1)) {
		return nil, fmt.Errorf("Unable to create simple send from %s = %v", datOut.Addr, simpleSend)
	}

	return &SimpleTransaction{
		Receiver: recOut.Addr,
		Data:     &simpleSend,
	}, nil
}


// Create a simple transaction out of class A outputs
func MakeClassASimpleSend(outputs []Output) (*SimpleTransaction, error) {
	var data *SimpleSend

	// Find the data address
	log.Println("Locating data-address.")
	for _, output := range outputs {
		simpleSend := DecodeFromAddress(output.Addr)
		if simpleSend.TransactionType == 0 && (simpleSend.CurrencyId == 2 || simpleSend.CurrencyId == 1) {
			data = &simpleSend
			break
		}
	}

	if data == nil {
		return nil, errors.New("Not a mastercoin simple send")
	}

	log.Println("Data address found:", data)

	// Level 1 - Loop over all Exodus-valued outputs to assume this is a perfect transaction
	log.Println("Locating sequence number: ", data.Sequence+1)

	log.Println("Attempting Level 1 search")
	receiver, err := locateRecipientAddress(outputs, data.Sequence, false)
	if err != nil {
		return nil, err
	}

	// Level 2 - Loop over all outputs and locate correct sequence
	if receiver == "" {
		log.Println("No recipient found, Attempting Level 2 search")
		receiver, err = locateRecipientAddress(outputs, data.Sequence, true)
		if err != nil {
			return nil, err
		}

		// We still don't have anything, invalidate
		if receiver == "" {
			return nil, errors.New("No recipient address found, invalidating")
		}
	}

	return &SimpleTransaction{
		Receiver: receiver,
		Data:     data,
	}, nil
}

// Loops through an given amount of outputs and locate recipient address
// with the required sequence number
func locateRecipientAddress(outputs []Output, seqNumber byte, checkAll bool) (address string, err error) {
	var exodusValue int64
	for _, output := range outputs {
		if output.Addr == ExodusAddress {
			exodusValue = output.Value
		}
	}

	for _, output := range outputs {
		if output.Addr == ExodusAddress || (output.Value != exodusValue && !checkAll) {
			continue
		}

		rawData := btcutil.Base58Decode(output.Addr)
		sequence := rawData[1]

		log.Println("Address", output.Addr, "has sequence", sequence)

		if sequence == seqNumber+1 {
			if address != "" {
				err = errors.New("Multiple recipients found, invalidating")
				return
			}
			log.Println("Located recipient address:", address)
			address = output.Addr
		}
	}
	return
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

type FundraiserTransaction struct {
	Addr  string
	Value int64
	Time  int64
}

func NewFundraiserTransaction(addr string, value int64, time int64) (*FundraiserTransaction, error) {
	tx := &FundraiserTransaction{
		Addr:  addr,
		Value: 0,
		Time:  time,
	}

	// Base line amount bought from fundraiser
	mastercoinBought := value / 1e8 * 100

	// Bonus amount received
	timeDifference := FundraiserEndTime - time
	if timeDifference > 0 {
		// TODO THIS IS WRONG! (DON'T USE FLOATS)
		// Ceil floor? What do we do.
		mastercoinBought += (mastercoinBought * (timeDifference * 0.1))
	}

	tx.Value = mastercoinBought

	log.Println("Attempting to create fundraiser tx", tx)

	return tx, nil
}
