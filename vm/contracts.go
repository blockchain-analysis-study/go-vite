package vm

import (
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/vm/abi"
	"github.com/vitelabs/go-vite/vm/contracts"
	cabi "github.com/vitelabs/go-vite/vm/contracts/abi"
	"github.com/vitelabs/go-vite/vm/util"
)

type precompiledContract struct {
	m   map[string]contracts.PrecompiledContractMethod
	abi abi.ABIContract
}

var simpleContracts = map[types.Address]*precompiledContract{
	types.AddressRegister: {
		map[string]contracts.PrecompiledContractMethod{
			cabi.MethodNameRegister:       &contracts.MethodRegister{},
			cabi.MethodNameCancelRegister: &contracts.MethodCancelRegister{},
			// TODO not support reward this version cabi.MethodNameReward:             &contracts.MethodReward{},
			cabi.MethodNameUpdateRegistration: &contracts.MethodUpdateRegistration{},
		},
		cabi.ABIRegister,
	},
	types.AddressVote: {
		map[string]contracts.PrecompiledContractMethod{
			cabi.MethodNameVote:       &contracts.MethodVote{},
			cabi.MethodNameCancelVote: &contracts.MethodCancelVote{},
		},
		cabi.ABIVote,
	},
	types.AddressPledge: {
		map[string]contracts.PrecompiledContractMethod{
			cabi.MethodNamePledge:       &contracts.MethodPledge{},
			cabi.MethodNameCancelPledge: &contracts.MethodCancelPledge{},
		},
		cabi.ABIPledge,
	},
	/* TODO not support consensus group this version
	types.AddressConsensusGroup: {
		map[string]contracts.PrecompiledContractMethod{
			contracts.MethodNameCreateConsensusGroup:   &contracts.MethodCreateConsensusGroup{},
			contracts.MethodNameCancelConsensusGroup:   &contracts.MethodCancelConsensusGroup{},
			contracts.MethodNameReCreateConsensusGroup: &contracts.MethodReCreateConsensusGroup{},
		},
		contracts.ABIConsensusGroup,
	},*/
	types.AddressMintage: {
		map[string]contracts.PrecompiledContractMethod{
			cabi.MethodNameMintage:             &contracts.MethodMintage{},
			cabi.MethodNameMintageCancelPledge: &contracts.MethodMintageCancelPledge{},
			cabi.MethodNameMint:                &contracts.MethodMint{},
			cabi.MethodNameIssue:               &contracts.MethodIssue{},
			cabi.MethodNameBurn:                &contracts.MethodBurn{},
			cabi.MethodNameTransferOwner:       &contracts.MethodTransferOwner{},
			cabi.MethodNameChangeTokenType:     &contracts.MethodChangeTokenType{},
		},
		cabi.ABIMintage,
	},
	types.AddressDexFund: {
		map[string]contracts.PrecompiledContractMethod{
			contracts.MethodNameDexFundUserDeposit:             &contracts.MethodDexFundUserDeposit{},
			contracts.MethodNameDexFundUserWithdraw:             &contracts.MethodDexFundUserWithdraw{},
			contracts.MethodNameDexFundNewOrder:             &contracts.MethodDexFundNewOrder{},
			contracts.MethodNameDexFundSettleOrders:             &contracts.MethodDexFundSettleOrders{},
			contracts.MethodNameDexFundFeeDividend:             &contracts.MethodDexFundFeeDividend{},
			contracts.MethodNameDexFundMinedVxDividend:             &contracts.MethodDexFundMinedVxDividend{},
			contracts.MethodNameDexFundNewMarket:             &contracts.MethodDexFundNewMarket{},
		},
		contracts.ABIDexFund,
	},
	types.AddressDexTrade: {
		map[string]contracts.PrecompiledContractMethod{
			contracts.MethodNameDexTradeNewOrder:             &contracts.MethodDexTradeNewOrder{},
			contracts.MethodNameDexTradeCancelOrder:             &contracts.MethodDexTradeCancelOrder{},
		},
		contracts.ABIDexTrade,
	},
}

func GetPrecompiledContract(addr types.Address, methodSelector []byte) (contracts.PrecompiledContractMethod, bool, *abi.Method, error) {
	var (
		method *abi.Method
		err error
	)
	p, ok := simpleContracts[addr]
	if ok {
		if method, err = p.abi.MethodById(methodSelector); err == nil {
			c, ok := p.m[method.Name]
			return c, ok, method, nil
		} else {
			return nil, ok, nil, util.ErrAbiMethodNotFound
		}
	}
	return nil, ok, nil, nil
}
