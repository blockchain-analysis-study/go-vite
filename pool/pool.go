package pool

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	ch "github.com/vitelabs/go-vite/chain"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/ledger"
	"github.com/vitelabs/go-vite/log15"
	"github.com/vitelabs/go-vite/monitor"
	"github.com/vitelabs/go-vite/verifier"
	"github.com/vitelabs/go-vite/vm_context"
	"github.com/vitelabs/go-vite/wallet"
	"github.com/vitelabs/go-vite/wallet/keystore"
	"github.com/viteshan/naive-vite/common/face"
)

type PoolWriter interface {
	// for normal account
	AddDirectAccountBlock(address types.Address, vmAccountBlock *vm_context.VmAccountBlock) error

	// for contract account
	AddDirectAccountBlocks(address types.Address, received *vm_context.VmAccountBlock, sendBlocks []*vm_context.VmAccountBlock) error
}

type SnapshotProducerWriter interface {
	Lock()

	UnLock()

	AddDirectSnapshotBlock(block *ledger.SnapshotBlock) error

	RollbackAccountTo(addr types.Address, hash types.Hash, height uint64) error
}

type PoolReader interface {
	// received block in current? (key is requestHash)
	ExistInPool(address types.Address, requestHash types.Hash) bool
}

type BlockPool interface {
	PoolWriter
	PoolReader
	Start()
	Stop()
	// todo need use subscribe net
	Init(s syncer)
	Info(addr *types.Address) string
}

type commonBlock interface {
	Height() uint64
	Hash() types.Hash
	PrevHash() types.Hash
	checkForkVersion() bool
	resetForkVersion()
	forkVersion() int
}

func newForkBlock(v *ForkVersion) *forkBlock {
	return &forkBlock{firstV: v.Val(), v: v}
}

type forkBlock struct {
	firstV int
	v      *ForkVersion
}

func (self *forkBlock) forkVersion() int {
	return self.v.Val()
}
func (self *forkBlock) checkForkVersion() bool {
	return self.firstV == self.v.Val()
}
func (self *forkBlock) resetForkVersion() {
	val := self.v.Val()
	self.firstV = val
}

type pool struct {
	pendingSc *snapshotPool
	pendingAc sync.Map // key:address v:*accountPool

	sync syncer
	bc   ch.Chain
	wt   wallet.Manager

	snapshotVerifier *verifier.SnapshotVerifier
	accountVerifier  *verifier.AccountVerifier

	accountSubId  int
	snapshotSubId int

	rwMutex sync.RWMutex
	version *ForkVersion

	closed      chan struct{}
	wg          sync.WaitGroup
	accountCond *sync.Cond // if new block add, notify

	log log15.Logger
}

func (self *pool) Lock() {
	self.rwMutex.Lock()
}

func (self *pool) UnLock() {
	self.rwMutex.Unlock()
}

func (self *pool) RLock() {
	self.rwMutex.RLock()
}

func (self *pool) RUnLock() {
	self.rwMutex.RUnlock()
}

func NewPool(bc ch.Chain) BlockPool {
	self := &pool{bc: bc, rwMutex: sync.RWMutex{}, version: &ForkVersion{}, accountCond: sync.NewCond(&sync.Mutex{})}
	self.log = log15.New("module", "pool")
	return self
}

func (self *pool) Init(s syncer) {
	self.sync = s
	rw := &snapshotCh{version: self.version, bc: self.bc}
	fe := &snapshotSyncer{fetcher: s}
	v := &snapshotVerifier{}
	snapshotPool := newSnapshotPool("snapshotPool", self.version, v, fe, rw, self.log)
	snapshotPool.init(
		newTools(fe, rw),
		self)

	self.pendingSc = snapshotPool
}
func (self *pool) Info(addr *types.Address) string {
	if addr == nil {
		bp := self.pendingSc.blockpool
		cp := self.pendingSc.chainpool

		freeSize := len(bp.freeBlocks)
		compoundSize := len(bp.compoundBlocks)
		snippetSize := len(cp.snippetChains)
		currentLen := cp.current.size()
		chainSize := len(cp.chains)
		return fmt.Sprintf("freeSize:%d, compoundSize:%d, snippetSize:%d, currentLen:%d, chainSize:%d",
			freeSize, compoundSize, snippetSize, currentLen, chainSize)
	} else {
		ac := self.selfPendingAc(*addr)
		if ac == nil {
			return "pool not exist."
		}
		bp := ac.blockpool
		cp := ac.chainpool

		freeSize := len(bp.freeBlocks)
		compoundSize := len(bp.compoundBlocks)
		snippetSize := len(cp.snippetChains)
		currentLen := cp.current.size()
		chainSize := len(cp.chains)
		return fmt.Sprintf("freeSize:%d, compoundSize:%d, snippetSize:%d, currentLen:%d, chainSize:%d",
			freeSize, compoundSize, snippetSize, currentLen, chainSize)
	}
}
func (self *pool) Start() {
	self.closed = make(chan struct{})

	self.accountSubId = self.sync.SubscribeAccountBlock(self.AddAccountBlock)
	self.snapshotSubId = self.sync.SubscribeSnapshotBlock(self.AddSnapshotBlock)

	self.pendingSc.Start()
	go self.loopTryInsert()
	go self.loopCompact()
	go self.loopBroadcastAndDel()
}
func (self *pool) Stop() {
	self.sync.UnsubscribeAccountBlock(self.accountSubId)
	self.accountSubId = 0
	self.sync.UnsubscribeSnapshotBlock(self.snapshotSubId)
	self.snapshotSubId = 0

	self.pendingSc.Stop()
	close(self.closed)
	self.wg.Wait()
}

func (self *pool) AddSnapshotBlock(block *ledger.SnapshotBlock) {

	self.log.Info("receive snapshot block from network. height:" + strconv.FormatUint(block.Height, 10) + ", hash:" + block.Hash.String() + ".")

	self.pendingSc.AddBlock(newSnapshotPoolBlock(block, self.version))
}

func (self *pool) AddDirectSnapshotBlock(block *ledger.SnapshotBlock) error {
	cBlock := newSnapshotPoolBlock(block, self.version)
	err := self.pendingSc.AddDirectBlock(cBlock)
	if err != nil {
		self.pendingSc.f.broadcastBlock(block)
	}
	return err
}

func (self *pool) AddAccountBlock(address types.Address, block *ledger.AccountBlock) {
	self.log.Info(fmt.Sprintf("receive account block from network. addr:%s, height:%d, hash:%s.", address, block.Height, block.Hash))
	self.selfPendingAc(address).AddBlock(newAccountPoolBlock(block, nil, self.version))
	self.selfPendingAc(address).AddReceivedBlock(block)

	self.accountCond.L.Lock()
	defer self.accountCond.L.Unlock()
	self.accountCond.Broadcast()
}

func (self *pool) AddDirectAccountBlock(address types.Address, block *vm_context.VmAccountBlock) error {
	defer monitor.LogTime("pool", "addDirectAccount", time.Now())
	self.RLock()
	defer self.RUnLock()

	ac := self.selfPendingAc(address)
	cBlock := newAccountPoolBlock(block.AccountBlock, block.VmContext, self.version)
	err := ac.AddDirectBlocks(cBlock, nil)
	if err != nil {
		ac.f.broadcastBlock(block.AccountBlock)
	}
	self.accountCond.L.Lock()
	defer self.accountCond.L.Unlock()
	self.accountCond.Broadcast()
	return err

}
func (self *pool) AddAccountBlocks(address types.Address, blocks []*ledger.AccountBlock) error {
	defer monitor.LogTime("pool", "addAccountArr", time.Now())

	for _, b := range blocks {
		self.AddAccountBlock(address, b)
	}

	self.accountCond.L.Lock()
	defer self.accountCond.L.Unlock()
	self.accountCond.Broadcast()
	return nil
}

func (self *pool) AddDirectAccountBlocks(address types.Address, received *vm_context.VmAccountBlock, sendBlocks []*vm_context.VmAccountBlock) error {
	defer monitor.LogTime("pool", "addDirectAccountArr", time.Now())
	self.RLock()
	defer self.RUnLock()
	ac := self.selfPendingAc(address)
	// todo
	var accountPoolBlocks []*accountPoolBlock
	for _, v := range sendBlocks {
		accountPoolBlocks = append(accountPoolBlocks, newAccountPoolBlock(v.AccountBlock, v.VmContext, self.version))
	}
	err := ac.AddDirectBlocks(newAccountPoolBlock(received.AccountBlock, received.VmContext, self.version), accountPoolBlocks)
	if err != nil {
		ac.f.broadcastReceivedBlocks(received, sendBlocks)
	}

	self.accountCond.L.Lock()
	defer self.accountCond.L.Unlock()
	self.accountCond.Broadcast()
	return err
}

func (self *pool) ExistInPool(address types.Address, requestHash types.Hash) bool {
	return self.selfPendingAc(address).ExistInCurrent(requestHash)
}

func (self *pool) ForkAccounts(accounts map[types.Address][]commonBlock) error {

	for k, v := range accounts {
		self.selfPendingAc(k).rollbackCurrent(v)
	}
	return nil
}

func (self *pool) PendingAccountTo(addr types.Address, h *ledger.HashHeight) error {
	this := self.selfPendingAc(addr)

	targetChain := this.findInTree(h.Hash, h.Height)
	if targetChain != nil {
		this.CurrentModifyToChain(targetChain)
		return nil
	}
	inPool := this.findInPool(h.Hash, h.Height)
	if !inPool {
		this.f.fetch(ledger.HashHeight{Hash: h.Hash, Height: h.Height}, 5)
	}
	return nil
}

func (self *pool) ForkAccountTo(addr types.Address, h *ledger.HashHeight) error {
	this := self.selfPendingAc(addr)
	err := self.RollbackAccountTo(addr, h.Hash, h.Height)

	if err != nil {
		return err
	}
	// find in tree
	targetChain := this.findInTree(h.Hash, h.Height)

	if targetChain == nil {
		cnt := h.Height - this.chainpool.diskChain.Head().Height()
		this.f.fetch(ledger.HashHeight{Height: h.Height, Hash: h.Hash}, cnt)
		err = this.CurrentModifyToEmpty()
	} else {
		err = this.CurrentModifyToChain(targetChain)
	}
	self.version.Inc()
	return err
}

func (self *pool) RollbackAccountTo(addr types.Address, hash types.Hash, height uint64) error {
	p := self.selfPendingAc(addr)

	// del some blcoks
	snapshots, accounts, e := p.rw.delToHeight(height)
	if e != nil {
		return e
	}

	// rollback snapshot chain in pool
	err := self.pendingSc.rollbackCurrent(snapshots)
	if err != nil {
		return err
	}
	// rollback accounts chain in pool
	for k, v := range accounts {
		err = self.selfPendingAc(k).rollbackCurrent(v)
		if err != nil {
			return err
		}
	}
	return err
}

func (self *pool) selfPendingAc(addr types.Address) *accountPool {
	chain, ok := self.pendingAc.Load(addr)

	if ok {
		return chain.(*accountPool)
	}

	// lazy load
	rw := &accountCh{address: addr, rw: self.bc, version: self.version}
	f := &accountSyncer{address: addr, fetcher: self.sync}
	v := &accountVerifier{}
	p := newAccountPool("accountChainPool-"+addr.Hex(), rw, self.version, self.log)

	p.Init(newTools(f, v, rw), self)

	chain, _ = self.pendingAc.LoadOrStore(addr, p)
	return chain.(*accountPool)
}
func (self *pool) loopTryInsert() {
	self.wg.Add(1)
	defer self.wg.Done()

	t := time.NewTicker(time.Millisecond * 20)
	defer t.Stop()
	sum := 0
	for {
		select {
		case <-self.closed:
			return
		case <-t.C:
			if sum == 0 {
				self.accountCond.L.Lock()
				self.accountCond.Wait()
				self.accountCond.L.Unlock()
				monitor.LogEvent("pool", "tryInsertSleep")
			}
			sum = 0
			sum += self.accountsTryInsert()
		default:
			sum += self.accountsTryInsert()
		}
	}
}

func (self *pool) accountsTryInsert() int {
	monitor.LogEvent("pool", "tryInsert")
	sum := 0
	var pending []*accountPool
	self.pendingAc.Range(func(_, v interface{}) bool {
		p := v.(*accountPool)
		pending = append(pending, p)
		return true
	})
	var tasks []verifyTask
	for _, p := range pending {
		task := p.TryInsert()
		if task != nil {
			self.fetchForTask(task)
			tasks = append(tasks, task)
			sum = sum + 1
		}
	}
	return sum
}

func (self *pool) loopCompact() {
	self.wg.Add(1)
	defer self.wg.Done()

	t := time.NewTicker(time.Millisecond * 40)
	defer t.Stop()
	sum := 0
	for {
		select {
		case <-self.closed:
			return
		case <-t.C:
			if sum == 0 {
				self.accountCond.L.Lock()
				self.accountCond.Wait()
				self.accountCond.L.Unlock()
			}
			sum = 0

			sum += self.accountsCompact()
		default:
			sum += self.accountsCompact()
		}
	}
}
func (self *pool) loopBroadcastAndDel() {
	self.wg.Add(1)
	defer self.wg.Done()

	broadcastT := time.NewTicker(time.Second * 30)
	delT := time.NewTicker(time.Second * 40)
	delUselessChainT := time.NewTicker(time.Minute)

	defer broadcastT.Stop()
	defer delT.Stop()
	for {
		select {
		case <-self.closed:
			return
		case <-broadcastT.C:
			addrList := self.listUnlockedAddr()
			for _, addr := range addrList {
				self.selfPendingAc(addr).broadcastUnConfirmedBlocks()
			}
		case <-delT.C:
			addrList := self.listUnlockedAddr()
			for _, addr := range addrList {
				self.delTimeoutUnConfirmedBlocks(addr)
			}
		case <-delUselessChainT.C:
			// del some useless chain in pool
			self.delUseLessChains()
		}
	}
}

func (self *pool) delUseLessChains() {
	self.pendingSc.loopDelUselessChain()
	var pendings []*accountPool
	self.pendingAc.Range(func(_, v interface{}) bool {
		p := v.(*accountPool)
		pendings = append(pendings, p)
		return true
	})
	for _, v := range pendings {
		v.loopDelUselessChain()
	}
}

func (self *pool) listUnlockedAddr() []types.Address {
	var todoAddress []types.Address
	status, e := self.wt.KeystoreManager.Status()
	if e != nil {
		return todoAddress
	}
	for k, v := range status {
		if v == keystore.Locked {
			todoAddress = append(todoAddress, k)
		}
	}
	return todoAddress
}

func (self *pool) accountsCompact() int {
	sum := 0
	var pendings []*accountPool
	self.pendingAc.Range(func(_, v interface{}) bool {
		p := v.(*accountPool)
		pendings = append(pendings, p)
		return true
	})
	for _, p := range pendings {
		sum = sum + p.Compact()
	}
	return sum
}
func (self *pool) fetchForTask(task verifyTask) []*face.FetchRequest {
	reqs := task.requests()
	if len(reqs) <= 0 {
		return nil
	}
	// if something in pool, deal with it.
	var existReqs []*face.FetchRequest
	for _, r := range reqs {
		exist := false
		//if r.Chain == "" {
		//	exist = self.pendingSc.ExistInCurrent(r)
		//} else {
		//	exist = self.selfPendingAc(r.Chain).ExistInCurrent(r)
		//}

		if !exist {
			self.sync.Fetch(r)
		} else {
			self.log.Info(fmt.Sprintf("block[%s] exist, should not fetch.", r.String()))
			existReqs = append(existReqs, &r)
		}
	}
	return existReqs
}
func (self *pool) delTimeoutUnConfirmedBlocks(addr types.Address) {
	headSnapshot := self.pendingSc.rw.headSnapshot()
	ac := self.selfPendingAc(addr)
	firstUnconfirmedBlock := ac.rw.getFirstUnconfirmedBlock(headSnapshot)
	if firstUnconfirmedBlock == nil {
		return
	}
	referSnapshot := self.pendingSc.rw.getSnapshotBlockByHash(firstUnconfirmedBlock.SnapshotHash)

	// verify account timeout
	if !self.pendingSc.v.verifyAccountTimeout(headSnapshot, referSnapshot) {
		self.Lock()
		defer self.UnLock()
		self.RollbackAccountTo(addr, firstUnconfirmedBlock.Hash, firstUnconfirmedBlock.Height)
	}
}