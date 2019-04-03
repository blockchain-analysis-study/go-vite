package net

import (
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/vitelabs/go-vite/common"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/ledger"
	"github.com/vitelabs/go-vite/monitor"
	"github.com/vitelabs/go-vite/p2p"
	"github.com/vitelabs/go-vite/tools/list"
	"github.com/vitelabs/go-vite/vite/net/message"
)

type code p2p.Code

const (
	HandshakeCode code = iota
	StatusCode
	ForkCode // tell peer it has forked, use for respond GetSnapshotBlocksCode
	GetSubLedgerCode
	GetSnapshotBlocksCode // get snapshotblocks without content
	GetSnapshotBlocksContentCode
	GetFullSnapshotBlocksCode   // get snapshotblocks with content
	GetSnapshotBlocksByHashCode // chain batch of hash
	GetSnapshotBlocksContentByHashCode
	GetFullSnapshotBlocksByHashCode
	GetAccountBlocksCode       // query single AccountChain
	GetMultiAccountBlocksCode  // query multi AccountChain
	GetAccountBlocksByHashCode // query accountBlocks by hashList
	GetFilesCode
	GetChunkCode
	SubLedgerCode
	FileListCode
	SnapshotBlocksCode
	SnapshotBlocksContentCode
	FullSnapshotBlocksCode
	AccountBlocksCode
	NewSnapshotBlockCode
	NewAccountBlockCode

	ExceptionCode = 127
)

var msgNames = [...]string{
	HandshakeCode:                      "HandShakeMsg",
	StatusCode:                         "StatusMsg",
	ForkCode:                           "ForkMsg",
	GetSubLedgerCode:                   "GetSubLedgerMsg",
	GetSnapshotBlocksCode:              "GetSnapshotBlocksMsg",
	GetSnapshotBlocksContentCode:       "GetSnapshotBlocksContentMsg",
	GetFullSnapshotBlocksCode:          "GetFullSnapshotBlocksMsg",
	GetSnapshotBlocksByHashCode:        "GetSnapshotBlocksByHashMsg",
	GetSnapshotBlocksContentByHashCode: "GetSnapshotBlocksContentByHashMsg",
	GetFullSnapshotBlocksByHashCode:    "GetFullSnapshotBlocksByHashMsg",
	GetAccountBlocksCode:               "GetAccountBlocksMsg",
	GetMultiAccountBlocksCode:          "GetMultiAccountBlocksMsg",
	GetAccountBlocksByHashCode:         "GetAccountBlocksByHashMsg",
	GetFilesCode:                       "GetFileMsg",
	GetChunkCode:                       "GetChunkMsg",
	SubLedgerCode:                      "SubLedgerMsg",
	FileListCode:                       "FileListMsg",
	SnapshotBlocksCode:                 "SnapshotBlocksMsg",
	SnapshotBlocksContentCode:          "SnapshotBlocksContentMsg",
	FullSnapshotBlocksCode:             "FullSnapshotBlocksMsg",
	AccountBlocksCode:                  "AccountBlocksMsg",
	NewSnapshotBlockCode:               "NewSnapshotBlockMsg",
	NewAccountBlockCode:                "NewAccountBlockMsg",
}

func (t code) String() string {
	if t == ExceptionCode {
		return "ExceptionMsg"
	}

	if t > NewAccountBlockCode {
		return "UnkownMsg"
	}

	return msgNames[t]
}

type MsgHandler interface {
	ID() string
	Cmds() []code
	Handle(msg p2p.Msg, sender Peer) error
}

// @section statusHandler
type _statusHandler func(msg p2p.Msg, sender Peer) error

func statusHandler(msg p2p.Msg, sender Peer) error {
	defer monitor.LogTime("net", "handle_StatusMsg", time.Now())

	status := new(ledger.HashHeight)

	if err := status.Deserialize(msg.Payload); err != nil {
		return err
	}

	sender.setHead(status.Hash, status.Height)
	return nil
}

func (s _statusHandler) ID() string {
	return "status handler"
}

func (s _statusHandler) Cmds() []code {
	return []code{StatusCode}
}

func (s _statusHandler) Handle(msg p2p.Msg, sender Peer) error {
	return s(msg, sender)
}

// @section queryHandler
type queryHandler struct {
	lock     sync.RWMutex
	queue    list.List
	handlers map[code]MsgHandler
	term     chan struct{}
	wg       sync.WaitGroup
}

func newQueryHandler(chain Chain) *queryHandler {
	q := &queryHandler{
		handlers: make(map[code]MsgHandler),
		queue:    list.New(),
	}

	//q.addHandler(&getSubLedgerHandler{chain})
	q.addHandler(&getSnapshotBlocksHandler{chain})
	q.addHandler(&getAccountBlocksHandler{chain})
	//q.addHandler(&getChunkHandler{chain})

	return q
}

func (q *queryHandler) start() {
	q.term = make(chan struct{})

	q.wg.Add(1)
	common.Go(q.loop)
}

func (q *queryHandler) stop() {
	if q.term == nil {
		return
	}

	select {
	case <-q.term:
	default:
		close(q.term)
		q.wg.Wait()
	}
}

func (q *queryHandler) addHandler(handler MsgHandler) {
	for _, cmd := range handler.Cmds() {
		q.handlers[cmd] = handler
	}
}

func (q *queryHandler) ID() string {
	return "query handler"
}

func (q *queryHandler) Cmds() []code {
	return []code{GetSubLedgerCode, GetSnapshotBlocksCode, GetAccountBlocksCode, GetChunkCode}
}

func (q *queryHandler) Handle(msg p2p.Msg, sender Peer) error {
	q.lock.Lock()
	//q.queue.Append(e)
	q.lock.Unlock()

	return nil
}

func (q *queryHandler) loop() {
	defer q.wg.Done()
	/*
		const batch = 10
		tasks := make([]*queryTask, batch)
		index := 0
		var ele interface{}

		for {
			select {
			case <-q.term:
				return
			default:
				// next
			}

			q.lock.Lock()
			for index, ele = 0, q.queue.Shift(); ele != nil; ele = q.queue.Shift() {
				tasks[index] = ele.(*queryTask)
				index++
				if index >= batch {
					break
				}
			}
			q.lock.Unlock()

			if index == 0 {
				time.Sleep(200 * time.Millisecond)
			} else {
				netLog.Info(fmt.Sprintf("retrive %d query tasks", index))

				for _, event := range tasks[:index] {
					cmd := code(event.Msg.Cmd)
					if h, ok := q.handlers[cmd]; ok {
						if err := h.Handle(event.Msg, event.Sender); err != nil {
							event.Sender.Report(err)
						}
					}
				}
			}
		}
	*/
}

type getSnapshotBlocksHandler struct {
	chain snapshotBlockReader
}

func (s *getSnapshotBlocksHandler) ID() string {
	return "GetSnapshotBlocks"
}

func (s *getSnapshotBlocksHandler) Cmds() []code {
	return []code{GetSnapshotBlocksCode}
}

func (s *getSnapshotBlocksHandler) Handle(msg p2p.Msg, sender Peer) (err error) {
	defer monitor.LogTime("net", "handle_GetSnapshotBlocksMsg", time.Now())

	req := new(message.GetSnapshotBlocks)

	if err = req.Deserialize(msg.Payload); err != nil {
		return
	}

	netLog.Info(fmt.Sprintf("receive %s from %s", req, sender.Address()))

	var block *ledger.SnapshotBlock
	if req.From.Hash != types.ZERO_HASH {
		block, err = s.chain.GetSnapshotBlockByHash(req.From.Hash)
	} else {
		block, err = s.chain.GetSnapshotBlockByHeight(req.From.Height)
	}

	if err != nil || block == nil {
		netLog.Warn(fmt.Sprintf("handle %s from %s error: %v", req, sender.Address(), err))
		return sender.send(ExceptionCode, msg.Id, message.Missing)
	}

	// use for split
	var from, to uint64
	if req.Forward {
		from = block.Height
		to = from + req.Count - 1
	} else {
		to = block.Height
		if to >= req.Count {
			from = to - req.Count + 1
		} else {
			from = 0
		}
	}
	chunks := splitChunk(from, to, downloadTaskSize)

	var blocks []*ledger.SnapshotBlock
	for _, c := range chunks {
		blocks, err = s.chain.GetSnapshotBlocksByHeight(c[0], true, c[1]-c[0]+1)
		if err != nil || len(blocks) == 0 {
			netLog.Warn(fmt.Sprintf("handle %s from %s error: %v", req, sender.Address(), err))
			monitor.LogEvent("net/handle", "GetSnapshotBlocks_Fail")
			return sender.send(ExceptionCode, msg.Id, message.Missing)
		}
		monitor.LogEvent("net/handle", "GetSnapshotBlocks_Success")

		if err = sender.sendSnapshotBlocks(blocks, msg.Id); err != nil {
			netLog.Error(fmt.Sprintf("send %d SnapshotBlocks to %s error: %v", len(blocks), sender.Address(), err))
			return
		} else {
			netLog.Info(fmt.Sprintf("send %d SnapshotBlocks to %s done", len(blocks), sender.Address()))
		}
	}

	return
}

// @section get account blocks
type getAccountBlocksHandler struct {
	chain accountBockReader
}

func (a *getAccountBlocksHandler) ID() string {
	return "GetAccountBlocks Handler"
}

func (a *getAccountBlocksHandler) Cmds() []code {
	return []code{GetAccountBlocksCode}
}

var NULL_ADDRESS = types.Address{}
var errGetABlocksMissingParam = errors.New("missing param to GetAccountBlocks")

func (a *getAccountBlocksHandler) Handle(msg p2p.Msg, sender Peer) (err error) {
	defer monitor.LogTime("net", "handle_GetAccountBlocksMsg", time.Now())

	req := new(message.GetAccountBlocks)

	if err = req.Deserialize(msg.Payload); err != nil {
		return
	}

	netLog.Info(fmt.Sprintf("receive %s from %s", req, sender.Address()))

	var block *ledger.AccountBlock
	if req.From.Hash != types.ZERO_HASH {
		// only need hash
		block, err = a.chain.GetAccountBlockByHash(req.From.Hash)
	} else if req.Address == NULL_ADDRESS {
		// missing start hash and address, so we can`t handle it
		return errGetABlocksMissingParam
	} else {
		// address and height
		block, err = a.chain.GetAccountBlockByHeight(req.Address, req.From.Height)
	}

	if err != nil || block == nil {
		netLog.Warn(fmt.Sprintf("handle %s from %s error: %v", req, sender.Address(), err))
		monitor.LogEvent("net/handle", "GetAccountBlocks_Fail")
		return sender.send(ExceptionCode, msg.Id, message.Missing)
	}

	address := block.AccountAddress

	// use for split
	var from, to uint64
	if req.Forward {
		from = block.Height
		to = from + req.Count - 1
	} else {
		to = block.Height
		if to >= req.Count {
			from = to - req.Count + 1
		} else {
			from = 0
		}
	}

	chunks := splitChunk(from, to, downloadTaskSize)

	var blocks []*ledger.AccountBlock
	for _, c := range chunks {
		blocks, err = a.chain.GetAccountBlocksByHeight(address, c[0], c[1]-c[0]+1)
		if err != nil || len(blocks) == 0 {
			netLog.Warn(fmt.Sprintf("handle %s from %s error: %v", req, sender.Address(), err))
			monitor.LogEvent("net/handle", "GetAccountBlocks_Fail")
			return sender.send(ExceptionCode, msg.Id, message.Missing)
		}

		monitor.LogEvent("net/handle", "GetAccountBlocks_Success")

		if err = sender.sendAccountBlocks(blocks, msg.Id); err != nil {
			netLog.Error(fmt.Sprintf("send %d AccountBlocks to %s error: %v", len(blocks), sender.Address(), err))
			return
		} else {
			netLog.Info(fmt.Sprintf("send %d AccountBlocks to %s done", len(blocks), sender.Address()))
		}
	}

	return
}

// helper
type accountBlockMap = map[types.Address][]*ledger.AccountBlock

func countAccountBlocks(mblocks accountBlockMap) (count uint64) {
	for _, blocks := range mblocks {
		for range blocks {
			count++
		}
	}

	return
}

func mapToSlice(mblocks accountBlockMap) []*ledger.AccountBlock {
	total := countAccountBlocks(mblocks)
	ret := make([]*ledger.AccountBlock, 0, total)
	for _, ablocks := range mblocks {
		ret = append(ret, ablocks...)
	}

	return ret
}

func splitAccountMap(mblocks accountBlockMap) (ret [][]*ledger.AccountBlock) {
	const batch = 1000
	var index, end int
	var add bool
	s := make([]*ledger.AccountBlock, 0, batch)

	for _, blocks := range mblocks {
		index = 0

		for index < len(blocks) {
			slotLen := cap(s) - len(s)
			if len(blocks[index:]) <= slotLen {
				s = append(s, blocks[index:]...)
				add = false
				break
			} else {
				end = index + slotLen
				s = append(s, blocks[index:end]...)
				index = end

				ret = append(ret, s)
				add = true
				s = make([]*ledger.AccountBlock, 0, batch)
			}
		}
	}

	if !add && len(s) != 0 {
		ret = append(ret, s)
	}

	return
}
