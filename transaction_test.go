package mscutil

import (
	"testing"
	"code.google.com/p/godec/dec"
	_"fmt"
)
func DecFromString(input string) *dec.Dec{
	x := new(dec.Dec)
	x.SetString(input)
	return x
}

func TestNewFundraiserTransaction(t *testing.T){
	tests := []struct {
		address Address
		value int64
		time int64
		tx FundraiserTransaction
	}{
		{Address{Addr: "1JvXUatqyiBs1gsp2vYopyc9QnStfVn2F1"}, 5e8, 1375306351, FundraiserTransaction{Addr: "1JvXUatqyiBs1gsp2vYopyc9QnStfVn2F1", Value: DecFromString("722.16013558"), Time:1375306351}},
		{Address{Addr: "1AsX2fv1QRFgCYT3F1YMWm37nWnW8TkZnm"}, 102e6, 1375429010, FundraiserTransaction{Addr: "1AsX2fv1QRFgCYT3F1YMWm37nWnW8TkZnm", Value: DecFromString("145.25201389"), Time:1375429010}},
		{Address{Addr: "163dXoX3bfrM8t8DwbhzfHdGYNQfYfXhK6"},50e8,1375531887, FundraiserTransaction{Addr: "163dXoX3bfrM8t8DwbhzfHdGYNQfYfXhK6", Value: DecFromString("7035.14632937"), Time:1375531887}},
	}
	for _, pair := range tests{
		result,_ := NewFundraiserTransaction(pair.address, pair.value, pair.time)

		if pair.tx.Value.Cmp(result.Value) != 0{
			t.Error("for", pair.value, 
			"and", pair.time,
			"Expected", pair.tx,
			"But got",result)
		}

	}
}

func TestMultipleSha(t *testing.T){
  tests := []struct {
	  source string
	  result string
	  times int
  }{
	  {"1CdighsfdfRcj4ytQSskZgQXbUEamuMUNF", "1D9A3DE5C2E22BF89A1E41E6FEDAB54582F8A0C3AE14394A59366293DD130C59", 1}, 
	  {"1CdighsfdfRcj4ytQSskZgQXbUEamuMUNF", "0800ED44F1300FB3A5980ECFA8924FEDB2D5FDBEF8B21BBA6526B4FD5F9C167C", 2}, 
  }

  for _, pair := range tests {
	  byteString := []byte(pair.source)
	  result := string(MultipleSha(byteString, pair.times))
	  if pair.result != result{
		  t.Error("for", pair.source, 
		  "Expected", pair.result,
			  "But got",result)
	  }
  }
}

func TestDeobfuscatePublicKeys(t *testing.T) {
	key1 := []byte("04e6da9c60084b43d28266243c636bcdaf4d8f17b5954e078d2dece7d4659e0dee3419a40b939c24ac813c692a323ca5207a6fb387ffe28e48f706c95dbf46648f")
	key2 := []byte("0226cb0561151d9045f6371cb09086ba7148d9942328bcf1de90c76edb35ccdda6")

	tests := []struct {
		publicKeys []Address
		ctPublicKey string
		receiver Address
	}{
		{ []Address{Address{Raw: key1},Address{Raw: key2}}, "0100000000000000020000000005f5e1000000000000000000000000000000", Address{Addr: "13NRX88EZbS5q81x6XFrTECzrciPREo821"}},
	}

	for _, pair := range tests {
		v := DeobfuscatePublicKeys(pair.publicKeys, pair.receiver) 
		if v[0] != pair.ctPublicKey{
			t.Error("For", pair.publicKeys,
				"Expected", pair.ctPublicKey,
				"Using xor", pair.receiver.Addr,
				"Got", v,
			)

		}
	}
}

