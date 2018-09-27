package discovery

import (
	"crypto/rand"
	"fmt"
	"github.com/pkg/errors"
	"github.com/vitelabs/go-vite/crypto/ed25519"
	"github.com/vitelabs/go-vite/monitor"
	"github.com/vitelabs/go-vite/p2p/block"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const seedCount = 20
const seedMaxAge = 7 * 24 * time.Hour

var tExpire = 24 * time.Hour        // nodes have been ping-pong checked during this time period are considered valid
var tRefresh = 1 * time.Hour        // refresh the node table at tRefresh intervals
var storeInterval = 5 * time.Minute // store nodes in table to db at storeDuration intervals
var checkInterval = 3 * time.Minute // check the oldest node in table at checkInterval intervals
var stayInTable = 5 * time.Minute   // minimal duration node stay in table can be store in db

var errUnsolicitedMsg = errors.New("unsolicited message")
var errMsgExpired = errors.New("message has expired")
var errUnkownMsg = errors.New("unknown message")
var errExtractPing = errors.New("extract message error, should be ping")
var errExtractFindNode = errors.New("extract message error, should be findnode")

func Priv2NodeID(priv ed25519.PrivateKey) (NodeID, error) {
	pub := priv.PubByte()
	return Bytes2NodeID(pub)
}

// @section Discovery
type Config struct {
	Priv      ed25519.PrivateKey
	DBPath    string
	BootNodes []*Node
	Conn      *net.UDPConn
	Self      *Node
}

type Discovery struct {
	bootNodes   []*Node
	self        *Node
	agent       *agent
	tab         *table
	db          *nodeDB
	term        chan struct{}
	refreshing  int32 // atomic, whether indicate node table is refreshing
	refreshDone chan struct{}
	wg          sync.WaitGroup
	blockList   *block.CuckooSet
}

func (d *Discovery) Start() {
	d.wg.Add(1)
	go d.tableLoop()

	d.wg.Add(1)
	go d.agent.start()

	discvLog.Info(fmt.Sprintf("discovery server %s start", d.self))
}

func (d *Discovery) Stop() {
	select {
	case <-d.term:
	default:
		close(d.term)
	}

	d.wg.Wait()

	discvLog.Info(fmt.Sprintf("discovery server %s stop", d.self))
}

func (d *Discovery) Block(ID NodeID, IP net.IP) {
	if !ID.IsZero() {
		d.blockList.Add(ID[:])
	}
	if IP != nil {
		d.blockList.Add(IP)
	}
}

func (d *Discovery) tableLoop() {
	defer d.wg.Done()

	checkTicker := time.NewTicker(checkInterval)
	refreshTicker := time.NewTimer(tRefresh)
	storeTicker := time.NewTicker(storeInterval)

	defer checkTicker.Stop()
	defer refreshTicker.Stop()
	defer storeTicker.Stop()

	d.RefreshTable()

	for {
		select {
		case <-refreshTicker.C:
			d.RefreshTable()

		case <-checkTicker.C:
			n := d.tab.pickOldest()
			if n != nil {
				d.agent.ping(n, func(node *Node, err error) {
					if err == nil {
						d.tab.bubble(n)
					} else {
						discvLog.Error(fmt.Sprintf("check oldest node error: %v", err))
						d.tab.removeNode(n)
					}
				})
			}

		case <-storeTicker.C:
			now := time.Now()
			d.tab.traverse(func(n *Node) {
				if now.Sub(n.addAt) > stayInTable {
					d.db.storeNode(n)
				}
			})

		case <-d.term:
			return
		}
	}
}

func (d *Discovery) findNode(to *Node, target NodeID, callback func(n *Node, nodes []*Node)) {
	d.agent.findnode(to, target, func(nodes []*Node, err error) {
		callback(to, nodes)
		if len(nodes) == 0 {
			discvLog.Error(fmt.Sprintf("find %s to %s, got %d neighbors, error: %v", target, to.UDPAddr(), len(nodes), err))
		} else {
			discvLog.Info(fmt.Sprintf("find %s to %s, got %d neighbors", target, to.UDPAddr(), len(nodes)))

			// add as many nodes as possible
			for _, n := range nodes {
				d.tab.addNode(n)
			}
		}
	})
}

func (d *Discovery) RandomNodes(result []*Node) int {
	return d.tab.randomNodes(result)
}

func (d *Discovery) Lookup(id NodeID) []*Node {
	return d.lookup(id, true)
}

func (d *Discovery) lookup(id NodeID, refreshIfNull bool) []*Node {
	var result *neighbors

	for {
		result = d.tab.findNeighbors(id, K)

		if len(result.nodes) > 0 || !refreshIfNull {
			break
		}

		d.RefreshTable()

		// table has refreshed, should not refresh again in a short time
		refreshIfNull = false
	}

	asked := make(map[NodeID]struct{}) // nodes has sent findnode message
	asked[d.self.ID] = struct{}{}

	// all nodes of responsive neighbors, use for filter to ensure the same node pushed once
	hasPushedIntoResult := make(map[NodeID]struct{})

	reply := make(chan []*Node, alpha)
	queries := 0

	for {
		for i := 0; i < len(result.nodes) && queries < alpha; i++ {
			n := result.nodes[i]
			if _, ok := asked[n.ID]; !ok {
				asked[n.ID] = struct{}{}
				d.findNode(n, id, func(n *Node, nodes []*Node) {
					reply <- nodes
				})
				queries++
			}
		}

		if queries == 0 {
			break
		}

		nodes := <- reply
		queries--
		for _, n := range nodes {
			if n != nil {
				if _, ok := hasPushedIntoResult[n.ID]; !ok {
					hasPushedIntoResult[n.ID] = struct{}{}
					result.push(n)
				}
			}
		}
	}

	return result.nodes
}

// find Node who`s equal id
func (d *Discovery) Resolve(id NodeID) *Node {
	nodes := d.tab.findNeighbors(id, 1).nodes

	if len(nodes) > 0 && id.Equal(nodes[0].ID) {
		return nodes[0]
	}

	nodes = d.Lookup(id)
	for _, n := range nodes {
		if n != nil && id.Equal(n.ID) {
			return n
		}
	}

	return nil
}

func (d *Discovery) HandleMsg(res *packet) {
	if res.msg.isExpired() {
		return
	}

	discvLog.Info(fmt.Sprintf("receive %s from %s@%s", res.msg, res.fromID, res.from))

	switch res.code {
	case pingCode:
		monitor.LogEvent("p2p/discv", "ping-receive")
		ping, _ := res.msg.(*Ping)

		node := &Node{
			ID:  res.fromID,
			IP:  res.from.IP,
			UDP: uint16(res.from.Port), // use the remote address
			TCP: ping.TCP,              // extract from the message
		}

		d.db.setLastPing(res.fromID, time.Now())
		d.agent.pong(node, res.hash)
		d.tab.addNode(node)

	case pongCode:
		monitor.LogEvent("p2p/discv", "pong-receive")
		d.db.setLastPong(res.fromID, time.Now())

	case findnodeCode:
		monitor.LogEvent("p2p/discv", "find-receive")

		findMsg, _ := res.msg.(*FindNode)
		nodes := d.tab.findNeighbors(findMsg.Target, K).nodes
		node := &Node{
			ID:  res.fromID,
			IP:  res.from.IP,
			UDP: uint16(res.from.Port), // use the remote address
		}

		d.agent.sendNeighbors(node, nodes)
	case neighborsCode:
		monitor.LogEvent("p2p/discv", "neighbors-receive")

	default:
		d.agent.send(&sendPkt{
			addr: res.from,
			code: exceptionCode,
			msg:  &Exception{
				Code: eUnKnown,
			},
		})
	}
}

func (d *Discovery) RefreshTable() {
	if !atomic.CompareAndSwapInt32(&d.refreshing, 0, 1) {
		select {
		case <-d.term:
		case <-d.refreshDone:
		}
		return
	}

	discvLog.Info("refresh table")

	monitor.LogEvent("p2p/discv", "refreshTable")

	// do refresh routine
	d.tab.initRand()
	d.loadInitNodes()
	d.lookup(d.self.ID, false)

	// find random NodeID in order to improve the defense of eclipse attack
	for i := 0; i < alpha; i++ {
		var id NodeID
		rand.Read(id[:])
		d.lookup(id, false)
	}

	// set the right state after refresh
	atomic.CompareAndSwapInt32(&d.refreshing, 1, 0)
	close(d.refreshDone)
	d.refreshDone = make(chan struct{})
	discvLog.Info("refresh table done")
}

func (d *Discovery) loadInitNodes() {
	nodes := d.db.randomNodes(seedCount, seedMaxAge) // get random nodes from db
	nodes = append(nodes, d.bootNodes...)
	for _, node := range nodes {
		d.tab.addNode(node)
	}
}

func New(cfg *Config) *Discovery {
	db, err := newDB(cfg.DBPath, 2, cfg.Self.ID)
	if err != nil {
		discvLog.Crit("create p2p db", "error", err)
	}

	d := &Discovery{
		bootNodes:   cfg.BootNodes,
		self:        cfg.Self,
		tab:         newTable(cfg.Self.ID, N),
		db:          db,
		term:        make(chan struct{}),
		refreshing:  0,
		refreshDone: make(chan struct{}),
		blockList:   block.NewCuckooSet(1000),
	}

	d.agent = &agent{
		self:       cfg.Self,
		conn:       cfg.Conn,
		priv:       cfg.Priv,
		term:       make(chan struct{}),
		pktHandler: d.HandleMsg,
		pool:       newWtPool(),
		write: make(chan *sendPkt, 100),
		read: make(chan *packet, 100),
	}

	return d
}