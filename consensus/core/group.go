package core

import (
	"fmt"
	"math/big"
	"time"

	"github.com/vitelabs/go-vite/common/types"
)

type MemberPlan struct {
	STime  time.Time
	ETime  time.Time
	Member types.Address
	Name   string
}

type GroupInfo struct {
	types.ConsensusGroupInfo

	// todo
	Repeat int32
	// checkLevel = 0  check all info
	// checkLevel = 1  Just check if it is in the producer collection
	// todo
	CheckLevel uint8

	genesisTime  time.Time
	seed         *big.Int
	PlanInterval uint64
	//countingTokenId types.TokenTypeId
}

func NewGroupInfo(genesisTime time.Time, info types.ConsensusGroupInfo) *GroupInfo {
	groupInfo := &GroupInfo{
		ConsensusGroupInfo: info,
		genesisTime:        genesisTime,
		seed:               new(big.Int).SetBytes(info.Gid.Bytes()),
	}

	if info.Gid == types.DELEGATE_GID {
		groupInfo.Repeat = 48
		groupInfo.CheckLevel = 1
	}

	if info.Gid == types.SNAPSHOT_GID {
		groupInfo.Repeat = 1
		groupInfo.CheckLevel = 0
	}
	groupInfo.PlanInterval = planInterval(groupInfo)
	return groupInfo
}

func planInterval(info *GroupInfo) uint64 {
	return uint64(info.Interval) * uint64(info.NodeCount) * uint64(info.PerCount) * uint64(info.Repeat)
}

func (self *GroupInfo) Time2Index(t time.Time) uint64 {
	subSec := int64(t.Sub(self.genesisTime).Seconds())

	i := uint64(subSec) / self.PlanInterval
	return i
}
func (self *GroupInfo) GenSTime(index uint64) time.Time {
	planInterval := self.PlanInterval
	return self.genesisTime.Add(time.Duration(planInterval*index) * time.Second)
}

func (self *GroupInfo) GenETime(index uint64) time.Time {
	planInterval := self.PlanInterval
	return self.genesisTime.Add(time.Duration(planInterval*(index+1)) * time.Second)
}

func (self *GroupInfo) Index2Time(index uint64) (time.Time, time.Time) {
	sTime := self.GenSTime(index)
	eTime := self.GenETime(index)
	return sTime, eTime
}

func (self *GroupInfo) GenVoteTime(index uint64) time.Time {
	if index < 2 {
		index = 2
	}
	planInterval := self.PlanInterval
	return self.genesisTime.Add(time.Duration(planInterval*(index-1)) * time.Second)
}

func (self *GroupInfo) GenPlan(index uint64, members []*Vote) []*MemberPlan {
	sTime := self.GenSTime(index)
	var plans []*MemberPlan
	for _, member := range members {
		for i := int64(0); i < self.PerCount; i++ {
			etime := sTime.Add(time.Duration(self.Interval) * time.Second)
			plan := MemberPlan{STime: sTime, ETime: etime, Member: member.Addr, Name: member.Name}
			plans = append(plans, &plan)
			sTime = etime
		}
	}
	return plans
}
func (self *GroupInfo) GenPlanByAddress(index uint64, members []types.Address) []*MemberPlan {
	sTime := self.GenSTime(index)
	var plans []*MemberPlan

	for j := int32(0); j < self.Repeat; j++ {
		for _, member := range members {
			for i := int64(0); i < self.PerCount; i++ {
				etime := sTime.Add(time.Duration(self.Interval) * time.Second)
				plan := MemberPlan{STime: sTime, ETime: etime, Member: member}
				plans = append(plans, &plan)
				sTime = etime
			}
		}
	}
	return plans
}

func (self *GroupInfo) String() string {
	return fmt.Sprintf("genesisTime:%s, memberCnt:%d, interval:%d, perCnt:%d, randCnt:%d, randRange:%d, seed:%s, countingTokenId:%s",
		self.genesisTime.String(), self.NodeCount, self.Interval, self.PerCount, self.RandCount, self.RandRank, self.seed, self.CountingTokenId.String())

}
