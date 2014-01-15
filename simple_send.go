package mscutil

import (
	"encoding/binary"
	"fmt"
	"github.com/conformal/btcutil"
	"log"
	"math/big"
	"strconv"
	"strings"
)

//func main(){
//  logging.SetFormatter(logging.MustStringFormatter("[%{level}] %{message}"))
//  logging.SetLevel(logging.DEBUG, "")

//  ss := DecodeFromAddress("15NoSD4F1ULYHGW3TK6khe1rLSS2qoysaX")
//  ss.Explain()

//  a := ss.SerializeToAddress()

//  ss2 := DecodeFromAddress(a)
//  ss2.Explain()

//  ss2.SerializeToKey()
//}

type Message struct {
}

type SimpleSend struct {
	CurrencyId      uint32
	Sequence        byte
	TransactionType uint32
	Amount          uint64
}

func (ss *SimpleSend) Explain() {
	fmt.Println("This is a Simple Send transaction for currency id", ss.CurrencyId, "and Amount", ss.Amount)
}

// Converts a number to a binary byte slice
// i.e. 4 => [0,0,0,4]
// 256 => [0,0,1,0]
// or in the case of 64
// 4 = > [0,0,0,0,0,0,0,4]
func makeBinary(value interface{}) []byte {
	var z []byte
	var val uint64

	amount := 4

	if v, ok := value.(uint64); ok {
		amount = 8
		val = v
	} else if v, ok := value.(uint32); ok {
		val = uint64(v)
	} else {
		log.Panic("makeBinary requires a value that's either a uint32 or an uint64, got:", value)
	}

	str := strconv.FormatUint(val, 10)

	number := new(big.Int)
	number.SetString(str, 10)

	template := make([]byte, amount)

	x := number.Bytes()
	z = append(template[:(amount-len(x))], x...)

	return z
}

// Converts a number to a string array
// i.e. 100,8 => [0,0,0,0,0,1,0,0]
// i.e. 66 ,4 => [0,0,6,6]
func makeStringArray(value string, length int) []string {
	z := make([]string, length)
	for i, _ := range z {
		z[i] = "0"
	}

	pointer := length - len(value)
	for _, val := range value {
		z[pointer] = fmt.Sprintf("%c", val)
		pointer++
	}
	return z
}

// Takes SerializeToKey output and builds a valid, obfuscated public key
func (ss *SimpleSend) SerializeToCompressedPublicKey(xor_target string) string {
	// 1. Create a value to XOR our data with by SHA-ing the xor_target a couple of times
	// 2. Use the XOR data, add the public key type (02) and add the random brute force package (00)
	// 3. Mange the last two characters until we have a valid key
	return "nothing"
}

// Encodes as Class B
// Encodes the data to a format that will be used as Obfuscate source
func (ss *SimpleSend) SerializeToKey() string {
	log.Println("Encoding data to KEY")

	raw := make([]string, 62)
	// Initialises our raw data with all zeros, except for the Sequence number.
	for i, _ := range raw {
		raw[i] = "0"

		// This is the 'fake' Sequence number, which we don't really need for Class B
		if i == 1 {
			raw[i] = "1"
		}
	}

	transactionType := makeStringArray(strconv.FormatUint(uint64(ss.TransactionType), 16), 8)
	log.Println("Transaction type: ", transactionType)

	currencyId := makeStringArray(strconv.FormatUint(uint64(ss.CurrencyId), 16), 8)
	log.Println("Currency ID: ", currencyId)

	amount := makeStringArray(strconv.FormatUint(ss.Amount, 16), 16)
	log.Println("Amount: ", amount)

	// Start of the data
	pointer := 2

	// Takes our 62 character string array and imposes the serialized values over it
	// [0,1,0,0,0,0,0,0,0,0,0....] becomes [0,1,0,0,0,0,0,1,2,3,4....] etc.
	// TODO: Perhaps make this a bit DRYer if there is no other way of doing it
	for _, value := range transactionType {
		raw[pointer] = value
		pointer++
	}
	for _, value := range currencyId {
		raw[pointer] = value
		pointer++
	}
	for _, value := range amount {
		raw[pointer] = value
		pointer++
	}

	rawString := strings.Join(raw, "")

	log.Println("Raw string: ", rawString)

	return rawString
}

// Encodes as Class A
func (ss *SimpleSend) SerializeToAddress() string {
	log.Println("Encoding data to address")

	raw := make([]byte, 25)
	var sequence byte = ss.Sequence
	raw[1] = sequence

	transactionType := makeBinary(ss.TransactionType)
	currencyId := makeBinary(ss.CurrencyId)
	amount := makeBinary(ss.Amount)

	//TODO: Can we optimise this?
	pointer := 2
	for _, value := range transactionType {
		raw[pointer] = value
		pointer++
	}
	for _, value := range currencyId {
		raw[pointer] = value
		pointer++
	}
	for _, value := range amount {
		raw[pointer] = value
		pointer++
	}
	//////////////////////////////

	rawData := btcutil.Base58Encode(raw)
	log.Println("Raw information: ", raw)
	log.Println("Encoded to address", rawData)
	return rawData
}

// Decodes Class A - Simple Sends
func DecodeFromAddress(address string) SimpleSend {
	log.Println("Decoding address '%s'.\n", address)

	rawData := btcutil.Base58Decode(address)

	log.Println("Base58 decoded data: %v \n", rawData)

	sequence := rawData[1]
	log.Println("Sequence %v", sequence)

	// Takes a byte array value and makes it an integer.
	// i.e. [0,0,1,2] becomes 257
	transactionType := binary.BigEndian.Uint32(rawData[2:6])
	log.Println("Transaction type: %v", transactionType)

	currencyId := binary.BigEndian.Uint32(rawData[6:10])
	log.Println("Currency id: %v ", currencyId)

	amount := binary.BigEndian.Uint64(rawData[10:18])
	log.Println("Amount: %v", amount)

	ss := SimpleSend{Amount: amount, CurrencyId: currencyId, TransactionType: transactionType, Sequence: sequence}
	return ss
}
