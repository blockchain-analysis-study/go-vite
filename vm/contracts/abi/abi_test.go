package abi

import (
	"bytes"
	"fmt"
	"github.com/vitelabs/go-vite/common/helper"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/vm/abi"
	"strconv"
	"strings"
	"testing"
)

func TestContractsABIInit(t *testing.T) {
	tests := []string{jsonRegister, jsonVote, jsonPledge, jsonConsensusGroup, jsonMintage}
	for _, data := range tests {
		if _, err := abi.JSONToABIContract(strings.NewReader(jsonRegister)); err != nil {
			t.Fatalf("json to abi failed, %v, %v", data, err)
		}
	}
}

func BenchmarkRegisterUnpackVariable(b *testing.B) {
	value := helper.HexToBytes("0000000000000000000000000000000000000000000000000000000000000100000000000000000000000000988dd19d15702dbf8a4d316f920b1fdcf57d4a50000000000000000000000000988dd19d15702dbf8a4d316f920b1fdcf57d4a50000000000000000000000000988dd19d15702dbf8a4d316f920b1fdcf57d4a500000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005c18dd820000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000066e6f646532350000000000000000000000000000000000000000000000000000")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registration := new(types.Registration)
		ABIRegister.UnpackVariable(registration, VariableNameRegistration, value)
	}
}

func BenchmarkVoteUnpackVariable(b *testing.B) {
	value := helper.HexToBytes("0000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000a73757065724e6f64653100000000000000000000000000000000000000000000")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vote := new(types.VoteInfo)
		ABIVote.UnpackVariable(vote, VariableNameVoteStatus, value)
	}
}

func BenchmarkConsensusGroupUnpackVariable(b *testing.B) {
	value := helper.HexToBytes("0000000000000000000000000000000000000000000000000000000000000019000000000000000000000000000000000000000000000000000000000000000300000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000160000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000001a000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000220000000000000000000000000b3db179e6ae63aa8d2114c386ebe91c9aa470ab50000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005ba78e1600000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000d3c21bcecceda10000000000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000076a7000000000000000000000000000000000000000000000000000000000000000000")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		info := new(types.ConsensusGroupInfo)
		ABIConsensusGroup.UnpackVariable(info, VariableNameConsensusGroupInfo, value)
	}
}

var (
	t1 = types.TokenTypeId{0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	t2 = types.TokenTypeId{0, 0, 0, 0, 0, 0, 0, 0, 0, 2}
	t3 = types.TokenTypeId{0, 0, 0, 0, 0, 0, 0, 0, 0, 3}
)

func TestDeleteTokenId(t *testing.T) {
	tests := []struct {
		input   []types.TokenTypeId
		tokenId types.TokenTypeId
		output  []types.TokenTypeId
	}{
		{[]types.TokenTypeId{t1}, t1, []types.TokenTypeId{}},
		{[]types.TokenTypeId{t1}, t2, []types.TokenTypeId{t1}},
		{[]types.TokenTypeId{t1, t2}, t1, []types.TokenTypeId{t2}},
		{[]types.TokenTypeId{t1, t2}, t2, []types.TokenTypeId{t1}},
		{[]types.TokenTypeId{t1, t2}, t3, []types.TokenTypeId{t1, t2}},
		{[]types.TokenTypeId{t1, t2, t3}, t1, []types.TokenTypeId{t2, t3}},
		{[]types.TokenTypeId{t1, t2, t3}, t2, []types.TokenTypeId{t1, t3}},
		{[]types.TokenTypeId{t1, t2, t3}, t3, []types.TokenTypeId{t1, t2}},
	}
	for _, test := range tests {
		var idList []byte
		for _, tid := range test.input {
			idList = AppendTokenId(idList, tid)
		}
		result := DeleteTokenId(idList, test.tokenId)
		var target []byte
		for _, tid := range test.output {
			target = AppendTokenId(target, tid)
		}
		if !bytes.Equal(result, target) {
			t.Fatalf("delete token id failed, delete %v from input %v, expected %v, got %v", test.tokenId, test.input, target, result)
		}
	}
}

func TestABIContract_MethodById(t *testing.T) {
	for _, e := range ABIMintage.Events {
		data := e.Id().Bytes()
		result := "{"
		for _, d := range data {
			result = result + strconv.Itoa(int(d)) + ","
		}
		result = result[:len(result)-1] + "}"
		fmt.Printf("%v: %v\n", e.Name, result)
	}
}
