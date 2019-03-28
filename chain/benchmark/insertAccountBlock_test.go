package chain_benchmark

import (
	"github.com/vitelabs/go-vite/chain/test_tools"

	"fmt"

	"github.com/vitelabs/go-vite/vm_context"
	"math/rand"
	"testing"
	"time"
)

func Benchmark_InsertAccountBlock(b *testing.B) {
	chainInstance := newChainInstance("insertAccountBlock", true)
	//chainInstance := newRocketChainInstance("insertAccountBlock", true)
	const (
		ACCOUNT_NUMS        = 100000
		ACCOUNT_BLOCK_LIMIT = 1000 * 10000

		PRINT_PER_COUNT = 1000

		CREATE_REQUEST_TX_PROBABILITY = 50

		LOOP_INSERT_SNAPSHOTBLOCK = false

		INSERT_SNAPSHOTBLOCK_INTERVAL = time.Millisecond * 1000

		INSERT_ACCOUNTBLOCK_INTERVAL = 0
	)

	cTxOptions := &test_tools.CreateTxOptions{
		MockVmContext: true,
		MockSignature: true,
	}

	tps := newTps(tpsOption{
		name:          "insertAccountBlock",
		printPerCount: PRINT_PER_COUNT,
	})

	accounts := test_tools.MakeAccounts(ACCOUNT_NUMS, chainInstance)
	accountLength := len(accounts)
	fmt.Printf("Account length is %d\n", accountLength)

	tps.Start()

	var loopTerminal chan struct{}
	if LOOP_INSERT_SNAPSHOTBLOCK {
		loopTerminal = loopInsertSnapshotBlock(chainInstance, INSERT_SNAPSHOTBLOCK_INTERVAL)
	}

	for tps.Ops() < ACCOUNT_BLOCK_LIMIT {
		for _, account := range accounts {
			createRequestTx := true

			if account.HasUnreceivedBlock() {
				randNum := rand.Intn(100)
				if randNum > CREATE_REQUEST_TX_PROBABILITY {
					createRequestTx = false
				}
			}
			var tx []*vm_context.VmAccountBlock
			if createRequestTx {
				toAccount := accounts[rand.Intn(accountLength)]
				tx = account.CreateRequestTx(toAccount, cTxOptions)
			} else {
				tx = account.CreateResponseTx(cTxOptions)
			}

			chainInstance.InsertAccountBlocks(tx)
			tps.doOne()
			if INSERT_ACCOUNTBLOCK_INTERVAL > 0 {
				time.Sleep(INSERT_ACCOUNTBLOCK_INTERVAL)
			}
		}
	}
	loopTerminal <- struct{}{}
	tps.Stop()
	tps.Print()
}
