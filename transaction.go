package mscutil

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/conformal/btcscript"
	"github.com/conformal/btcutil"
	"code.google.com/p/godec/dec"
	"log"
	"strconv"
	"strings"
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

		// There are some bogus transactions out there that don't generate a valid address, skip these outputs
		if len(a) > 0 {
			out := Output{Value: txOut.Value, Addr: a[0].Addr}

			addrs = append(addrs, out) // a is one address guaranteed
		}
	}

	return addrs
}

func FindInOutputs(outs []Output, finder func(output Output) bool) Output {
	for _, out := range outs {
		if finder(out) {
			return out
		}
	}
	return Output{}
}

func GetExodus(outs []Output) Output {
	return FindInOutputs(outs, func(output Output) bool {
		return output.Addr == ExodusAddress
	})
}

func MultipleSha(target []byte, times int) []byte {
	if times == 0 {
		return target
	}

	result := sha256.Sum256(MultipleSha(target, times-1))

	return []byte(strings.ToUpper(hex.EncodeToString(result[:])))
}

// Transformers obfuscated public keys in clear text keys
func DeobfuscatePublicKeys(multiSig []Address, sender Address) []string {
	data := make([]string, len(multiSig)-1)
	fmt.Println("Got keys:", multiSig[1:])
	// For each public key create a hash out of the reciever
	for i, sig := range multiSig[1:] {
		// Hash receiver x times
		hash := MultipleSha([]byte(sender.Addr), i+1)

		// Deobfuscated strings (which have been xor'ed) will be written and concatenated
		deobfuscatedBytes := make([]string, 32)

		// Skip first byte
		k := 2
		for j := 0; j < 62; j += 2 {
			hashPart, _ := strconv.ParseInt(string(hash[j:j+2]), 16, 16)
			pubKeyPart, _ := strconv.ParseInt(string(sig.Raw[k:k+2]), 16, 16)

			// Format base 16 and pad with 0 when applicable
			deobfuscatedBytes[k/2] = fmt.Sprintf("%02s", strconv.FormatInt(int64(byte(hashPart)^byte(pubKeyPart)), 16))

			k += 2
		}

		// Concatenate strings together
		data[i] = strings.Join(deobfuscatedBytes, "")
		fmt.Println("Data[i]", data[i])
	}

	return data
}

func GetAddrsClassB(tx *btcutil.Tx, sender Address) ([]string, string, error) {
	var addrs []Output

	// TODO if in the future we require multiple multisigs change this to a slice
	var multiSig []Address
	// var multiSigVal int64 

	tmx := tx.MsgTx()
	// Gather outputs from this tx, filter out the MultiSig
	for _, txOut := range tmx.TxOut {
		a, scriptType := GetAddrs(txOut.PkScript)
		// Assign the multi sig so we can work with it later on
		if scriptType == btcscript.MultiSigTy {
			// TODO: In the future we might want to add support for multiple outputs with multisig support
			multiSig = a
		} else {
			// Create regular outs for non multi sigs outputs
			addrs = append(addrs, Output{Value: txOut.Value, Addr: a[0].Addr})
		}
	}

	// Get the exodus address
	exodus := GetExodus(addrs)
	receiver := FindInOutputs(addrs, func(output Output) bool {
		return output.Value == exodus.Value
	})

	// If the receiver is empty
	if receiver == (Output{}) {
		return nil, "", errors.New("Unable to find recipient")
	}

	if len(multiSig) < 2 {
		return nil,"", errors.New("Invalid multisignature data, can't create Mastercoin transaction")
	}
	plainTextKeys := DeobfuscatePublicKeys(multiSig, sender)

	return plainTextKeys, receiver.Addr, nil

}

type SimpleTransaction struct {
	Receiver string
	Sender string
	Data     *SimpleSend
}

// Create simple transaction out of class B outputs
func MakeClassBSimpleSend(plainTextKeys []string, receiver string, sender Address) (*SimpleTransaction, error) {
	log.Println("Making class b simple transaction")

	// XXX if in the future simple sends may contain multiple data strings, update.
	simpleSend := DecodeFromPublicKeys(plainTextKeys)
	if !(simpleSend.TransactionType == 0 && (simpleSend.CurrencyId == 2 || simpleSend.CurrencyId == 1)) {
		//return nil, fmt.Errorf("Unable to create simple send from %s = %v", datOut.Addr, simpleSend)
		return nil, fmt.Errorf("Wutwut")
	}

	return &SimpleTransaction{
		Receiver: receiver,
		Data:     &simpleSend,
		Sender: sender.Addr,
	}, nil
}

// Create a simple transaction out of class A outputs
func MakeClassASimpleSend(sender Address, outputs []Output) (*SimpleTransaction, error) {
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
	log.Println("Looking for sequence number: ", data.Sequence+1)

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
	}else{
		log.Println("Found sequence number, receiver is", receiver)
	}

	return &SimpleTransaction{
		Receiver: receiver,
		Data:     data,
		Sender: sender.Addr,
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
	Value *dec.Dec
	Time  int64
}

func NewFundraiserTransaction(addr Address, value int64, time int64) (*FundraiserTransaction, error) {
	tx := &FundraiserTransaction{
		Addr:  addr.Addr,
		Time:  time,
	}

	// Base line amount bought from fundraiser, 1 btc gets you 100 msc
	mastercoinBought := dec.NewDecInt64((value * 100))
	mastercoinBought.Quo(mastercoinBought, dec.NewDecInt64(1e8), dec.Scale(18), dec.RoundHalfUp)

	fmt.Println("Baseline MSC Bought:", mastercoinBought)

	// Bonus amount received
	diff := dec.NewDecInt64(FundraiserEndTime - time)
	timeDifference := diff.Quo(diff, dec.NewDecInt64(604800), dec.Scale(18), dec.RoundHalfUp)

	fmt.Println("Time difference", timeDifference)

	if timeDifference.Cmp(dec.NewDecInt64(0)) > 0 {
		// bought += bought * (timediff * 0.1)

		x := new(dec.Dec)
		x.SetString("0.1")

		ratio := new(dec.Dec).Mul(timeDifference,x)
		fmt.Println("Ratio:", ratio)
		bonus := new(dec.Dec).Mul(ratio, mastercoinBought)
		bonus.Round(bonus, dec.Scale(18), dec.RoundDown)

		mastercoinBought.Add(mastercoinBought, bonus)
		mastercoinBought.Round(mastercoinBought, dec.Scale(8), dec.RoundHalfUp)
	}

	tx.Value = mastercoinBought

	Logger.Println("Created fundraiser transaction", tx)

	return tx, nil
}
