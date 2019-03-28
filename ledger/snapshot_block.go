package ledger

import (
	"bytes"
	"encoding/binary"
	"github.com/golang/protobuf/proto"
	"github.com/vitelabs/go-vite/common/fork"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/crypto"
	"github.com/vitelabs/go-vite/crypto/ed25519"
	"github.com/vitelabs/go-vite/log15"
	"github.com/vitelabs/go-vite/trie"
	"github.com/vitelabs/go-vite/vitepb"
	"sort"
	"time"
)

var snapshotBlockLog = log15.New("module", "ledger/snapshot_block")

type SnapshotContent map[types.Address]*HashHeight

type SnapshotContentItem struct {
	Address    *types.Address
	HashHeight *HashHeight
}

const ScItemBytesLen = types.AddressSize + types.HashSize + 8

func (scItem *SnapshotContentItem) Bytes() []byte {
	bytes := make([]byte, 0, ScItemBytesLen)
	// Address
	bytes = append(bytes, scItem.Address.Bytes()...)

	// Hash
	bytes = append(bytes, scItem.HashHeight.Hash.Bytes()...)

	// Height
	heightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBytes, scItem.HashHeight.Height)
	bytes = append(bytes, heightBytes...)

	return bytes
}

type SnapshotContentList []*SnapshotContentItem

func (scList SnapshotContentList) Len() int      { return len(scList) }
func (scList SnapshotContentList) Swap(i, j int) { scList[i], scList[j] = scList[j], scList[i] }
func (scList SnapshotContentList) Less(i, j int) bool {
	return bytes.Compare(scList[i].Address.Bytes(), scList[j].Address.Bytes()) <= 0
}

func NewSnapshotContentList(sc SnapshotContent) SnapshotContentList {
	scList := make(SnapshotContentList, 0, len(sc))
	for addr, hashHeight := range sc {
		scItemAddr := addr
		scList = append(scList, &SnapshotContentItem{
			Address: &scItemAddr,
			HashHeight: &HashHeight{
				Hash:   hashHeight.Hash,
				Height: hashHeight.Height,
			},
		})
	}

	return scList
}

func (sc SnapshotContent) Bytes() []byte {
	scList := NewSnapshotContentList(sc)
	sort.Sort(scList)

	bytes := make([]byte, 0, ScItemBytesLen*len(scList))
	for _, scItem := range scList {
		bytes = append(bytes, scItem.Bytes()...)
	}
	return bytes
}

func (sc SnapshotContent) DeProto(pb *vitepb.SnapshotContent) {
	for addrString, snapshotItem := range pb.Content {
		addr, _ := types.HexToAddress(addrString)
		accountBlockHash, _ := types.BytesToHash(snapshotItem.AccountBlockHash)

		sc[addr] = &HashHeight{
			Height: snapshotItem.AccountBlockHeight,
			Hash:   accountBlockHash,
		}
	}
}

func (sc SnapshotContent) Proto() *vitepb.SnapshotContent {
	pb := &vitepb.SnapshotContent{
		Content: make(map[string]*vitepb.SnapshotItem),
	}

	for addr, snapshotItem := range sc {
		pb.Content[addr.String()] = &vitepb.SnapshotItem{
			AccountBlockHash:   snapshotItem.Hash.Bytes(),
			AccountBlockHeight: snapshotItem.Height,
		}
	}
	return pb
}

func (sc *SnapshotContent) Serialize() ([]byte, error) {
	pb := sc.Proto()
	buf, err := proto.Marshal(pb)
	if err != nil {
		snapshotBlockLog.Error("proto.Marshal failed, error is "+err.Error(), "method", "SnapshotContent.Serialize")
		return nil, err
	}
	return buf, nil
}
func (sc *SnapshotContent) Deserialize(buf []byte) error {
	pb := &vitepb.SnapshotContent{}
	unmarshalErr := proto.Unmarshal(buf, pb)
	if unmarshalErr != nil {
		snapshotBlockLog.Error("proto.Unmarshal failed, error is "+unmarshalErr.Error(), "method", "SnapshotContent.Deserialize")
		return unmarshalErr
	}

	sc.DeProto(pb)
	return nil
}

type SnapshotBlock struct {
	Hash types.Hash `json:"hash"`

	PrevHash types.Hash `json:"prevHash"`
	Height   uint64     `json:"height"`
	producer *types.Address

	PublicKey ed25519.PublicKey `json:"publicKey"`
	Signature []byte            `json:"signature"`

	Timestamp *time.Time `json:"timestamp"`

	StateHash types.Hash `json:"stateHash"`
	StateTrie *trie.Trie `json:"-"`

	SnapshotContent SnapshotContent `json:"snapshotContent"`
}

func (sb *SnapshotBlock) ComputeHash() types.Hash {
	var source []byte
	// PrevHash
	source = append(source, sb.PrevHash.Bytes()...)

	// Height
	heightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBytes, sb.Height)
	source = append(source, heightBytes...)

	// Timestamp
	unixTimeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(unixTimeBytes, uint64(sb.Timestamp.Unix()))
	source = append(source, unixTimeBytes...)

	// StateHash
	source = append(source, sb.StateHash.Bytes()...)

	// Add fork name
	forkName := fork.GetRecentForkName(sb.Height)
	if forkName != "" {
		source = append(source, []byte(forkName)...)
	}

	// Snapshot Content
	if fork.IsMintFork(sb.Height) {
		scBytes := sb.SnapshotContent.Bytes()
		source = append(source, scBytes...)
	}

	hash, _ := types.BytesToHash(crypto.Hash256(source))
	return hash
}

func (sb *SnapshotBlock) Producer() types.Address {
	if sb.producer == nil {
		producer := types.PubkeyToAddress(sb.PublicKey)
		sb.producer = &producer
	}
	return *sb.producer
}

func (sb *SnapshotBlock) VerifySignature() bool {
	isVerified, verifyErr := crypto.VerifySig(sb.PublicKey, sb.Hash.Bytes(), sb.Signature)
	if verifyErr != nil {
		accountBlockLog.Error("crypto.VerifySig failed, error is "+verifyErr.Error(), "method", "VerifySignature")
	}
	return isVerified
}

func (sb *SnapshotBlock) proto() *vitepb.SnapshotBlock {
	pb := &vitepb.SnapshotBlock{}
	pb.Hash = sb.Hash.Bytes()
	pb.PrevHash = sb.PrevHash.Bytes()
	pb.Height = sb.Height
	pb.PublicKey = sb.PublicKey
	pb.Signature = sb.Signature
	pb.Timestamp = sb.Timestamp.UnixNano()
	pb.StateHash = sb.StateHash.Bytes()
	return pb
}

func (sb *SnapshotBlock) Proto() *vitepb.SnapshotBlock {
	pb := sb.proto()
	if sb.SnapshotContent != nil {
		pb.SnapshotContent = sb.SnapshotContent.Proto()
	}

	return pb
}

func (sb *SnapshotBlock) DeProto(pb *vitepb.SnapshotBlock) {
	sb.Hash, _ = types.BytesToHash(pb.Hash)
	sb.PrevHash, _ = types.BytesToHash(pb.PrevHash)
	sb.Height = pb.Height
	sb.PublicKey = pb.PublicKey
	sb.Signature = pb.Signature

	timestamp := time.Unix(0, pb.Timestamp)
	sb.Timestamp = &timestamp

	sb.StateHash, _ = types.BytesToHash(pb.StateHash)

	if pb.SnapshotContent != nil {
		sb.SnapshotContent = SnapshotContent{}
		sb.SnapshotContent.DeProto(pb.SnapshotContent)
	}
}

func (sb *SnapshotBlock) DbSerialize() ([]byte, error) {
	pb := sb.proto()
	buf, err := proto.Marshal(pb)
	if err != nil {
		snapshotBlockLog.Error("proto.Marshal failed, error is "+err.Error(), "method", "SnapshotBlock.DbSerialize")
		return nil, err
	}
	return buf, nil
}

func (sb *SnapshotBlock) Serialize() ([]byte, error) {
	pb := sb.Proto()
	buf, err := proto.Marshal(pb)
	if err != nil {
		snapshotBlockLog.Error("proto.Marshal failed, error is "+err.Error(), "method", "SnapshotBlock.Serialize")
		return nil, err
	}
	return buf, nil
}

func (sb *SnapshotBlock) Deserialize(buf []byte) error {
	pb := &vitepb.SnapshotBlock{}
	unmarshalErr := proto.Unmarshal(buf, pb)
	if unmarshalErr != nil {
		snapshotBlockLog.Error("proto.Unmarshal failed, error is "+unmarshalErr.Error(), "method", "SnapshotBlock.Deserialize")
		return unmarshalErr
	}

	sb.DeProto(pb)
	return nil
}
