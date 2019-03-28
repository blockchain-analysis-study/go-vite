package contracts

import (
	"github.com/vitelabs/go-vite/vm/util"
	"math/big"
)

const (
	RegisterGas               uint64 = 62200
	UpdateRegistrationGas     uint64 = 62200
	CancelRegisterGas         uint64 = 83200
	RewardGas                 uint64 = 238800
	VoteGas                   uint64 = 62000
	CancelVoteGas             uint64 = 62000
	PledgeGas                 uint64 = 21000
	CancelPledgeGas           uint64 = 21000
	CreateConsensusGroupGas   uint64 = 62200
	CancelConsensusGroupGas   uint64 = 83200
	ReCreateConsensusGroupGas uint64 = 62200
	MintageGas                uint64 = 83200
	MintageCancelPledgeGas    uint64 = 83200
	MintGas                   uint64 = 104525
	IssueGas                  uint64 = 69325
	BurnGas                   uint64 = 48837
	TransferOwnerGas          uint64 = 58981
	ChangeTokenTypeGas        uint64 = 63125

	cgNodeCountMin   uint8 = 3       // Minimum node count of consensus group
	cgNodeCountMax   uint8 = 101     // Maximum node count of consensus group
	cgIntervalMin    int64 = 1       // Minimum interval of consensus group in second
	cgIntervalMax    int64 = 10 * 60 // Maximum interval of consensus group in second
	cgPerCountMin    int64 = 1
	cgPerCountMax    int64 = 10 * 60
	cgPerIntervalMin int64 = 1
	cgPerIntervalMax int64 = 10 * 60

	RewardDayLimit     uint64 = 90
	rewardPrecForFloat uint   = 18

	registrationNameLengthMax int = 40

	tokenNameLengthMax   int = 40 // Maximum length of a token name(include)
	tokenSymbolLengthMax int = 10 // Maximum length of a token symbol(include)
)

var (
	viteTotalSupply                  = new(big.Int).Mul(big.NewInt(1e9), util.AttovPerVite)
	rewardPerBlock                   = big.NewInt(951293759512937595) // Reward pre snapshot block, rewardPreBlock * blockNumPerYear / viteTotalSupply = 3%
	pledgeAmountMin                  = new(big.Int).Mul(big.NewInt(10), util.AttovPerVite)
	pledgeAmountMin2                 = new(big.Int).Mul(big.NewInt(1000), util.AttovPerVite)
	mintageFee                       = new(big.Int).Mul(big.NewInt(1e3), util.AttovPerVite) // Mintage cost choice 1, destroy ViteToken
	mintagePledgeAmount              = new(big.Int).Mul(big.NewInt(1e5), util.AttovPerVite) // Mintage cost choice 2, pledge ViteToken for 3 month
	createConsensusGroupPledgeAmount = new(big.Int).Mul(big.NewInt(1000), util.AttovPerVite)

	float1                = new(big.Float).SetPrec(rewardPrecForFloat).SetInt64(1)
	additionForVoteReward = new(big.Int).Mul(big.NewInt(5e5), util.AttovPerVite)
)

type ContractsParams struct {
	MinPledgeHeight                  uint64 // Minimum pledge height
	CreateConsensusGroupPledgeHeight uint64 // Pledge height for registering to be a super node of snapshot group and common delegate group
	MintagePledgeHeight              uint64 // Pledge height for mintage if choose to pledge instead of destroy vite token
	RewardEndTimeLimit               uint64 // Cannot get snapshot block reward of current few blocks, for latest snapshot block could be reverted
	RewardTimeUnit                   uint64
}

var (
	ContractsParamsTest = ContractsParams{
		MinPledgeHeight:                  1,
		CreateConsensusGroupPledgeHeight: 1,
		MintagePledgeHeight:              1,
		RewardEndTimeLimit:               75,
		RewardTimeUnit:                   75 * 2,
	}
	ContractsParamsMainNet = ContractsParams{
		MinPledgeHeight:                  3600 * 24 * 3,
		CreateConsensusGroupPledgeHeight: 3600 * 24 * 3,
		MintagePledgeHeight:              3600 * 24 * 30 * 3,
		RewardEndTimeLimit:               3600 * 24,
		RewardTimeUnit:                   1152 * 75,
	}
)
