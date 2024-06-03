package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/parallel"
)

// Debug print
func print(item ...interface{}) { //利用 interface{} 来传递任意参数, 用...表示不限参数的个数
	//fmt.Print("[Debug]")
	fmt.Printf("%c[31;40;5m%s%c[0m", 0x1B, "[Debug Print]", 0x1B) //打印高亮文本
	for i := range item {
		fmt.Print(" ", item[i])
	}
	fmt.Print("\n")
}

//已经被 Hook 方法代替
// 解决了问题：如何读取指定区块的信息（交易，gas等）？
// 读取指定区块的交易信息,	并将交易转换为消息 Message
// A Message contains the data derived from a single transaction that is relevant to state processing.
// func ReadBlockTx(block *types.Block, db ethdb.Database, cacheConfig *core.CacheConfig) {
// 	print("\t\t\t\t\t\t---------------------------------------------------------------------------------------")
// 	print("\t\t\t\t\t\t|                                    Read Block Tx                                    |")
// 	print("\t\t\t\t\t\t---------------------------------------------------------------------------------------")
// 	for i, tx := range block.Transactions() {
// 		fmt.Printf("\n------------------------------------Transaction %d------------------------------------\n", i)

// 		//初始化一些参数
// 		var genesis *core.Genesis = nil
// 		triedb := triedb.NewDatabase(db, cacheConfig.TriedbConfig(genesis != nil && genesis.IsVerkle()))
// 		chainConfig, _, err := core.SetupGenesisBlockWithOverride(db, triedb, genesis, nil)
// 		if err != nil {
// 			print("👎Fail! ", err)
// 		}

// 		//将交易转换为消息
// 		signer := types.MakeSigner(chainConfig, block.Header().Number, block.Header().Time)
// 		msg, err := core.TransactionToMessage(tx, signer, block.Header().BaseFee)
// 		if err != nil {
// 			print("👎Transaction To Message fail! ", err)
// 		}

// 		//打印区块中每一笔 Transaction 的信息
// 		print("Tx From: ", msg.From)
// 		print("Tx To: ", msg.To)
// 		print("Tx Value: ", msg.Value)
// 		print("Tx GasLimit: ", msg.GasLimit)
// 		print("Tx Data: ", msg.Data)

// 	}
// }

type hook struct {
	Hash common.Hash `json:"hash"`
}

// 打印 Hook 信息并以 Json 形式返回
func OutputBlockHookInfo() {

	// //打印Hook从程序中勾取的信息, 包括 contract 的调用以及执行的 opcode
	print("Block Hash: ", parallel.GetBlockInfo().BlockHash)
	print("GasLimit: ", parallel.GetBlockInfo().GasLimit)
	for i, tx := range parallel.GetBlockInfo().Tx {
		fmt.Printf("\n\n\n------------------------------------Transaction %d------------------------------------\n", i)
		print("Tx Hash", tx.TxHash)
		print("Tx From: ", tx.From)
		print("Tx To: ", tx.To)
		print("Tx Value: ", tx.Value)
		print("Tx GasPrice: ", tx.GasPrice)
		//print("Tx Data: ", tx.Data)
		for _, q := range tx.CallQueue {
			//fmt.Printf("\nInvoke Contract: %d\n", j+1)
			fmt.Print("\n")
			print("Invoke Layer: ", q.Layer)
			print("Contract Address: ", q.ContractAddr)
			//print("Contract Opcode: ", q.OpcodeList)
			for _, op := range q.KeyOpcode {
				print("Key Opcode: ", op)
			}
			//print("Last Opcode: ", q.OpcodeList[len(q.OpcodeList)-1])
		}
	}

	//	将 BlockInfo 对象转化为 Json 对象
	jsonData, _ := json.Marshal(parallel.GetBlockInfo())
	file, err := os.Create("./output/txLog.json") //创建输出文件
	if err != nil {
		print(err)
	}
	file.Write(jsonData)
	defer file.Close()

}

// 从一个区块执行前的全局状态模拟执行一个区块
func DoProcess() {

	//读取数据库
	chainDataDir := "/home/user/common/docker/volumes/cp1_eth-docker_geth-eth1-data/_data/geth/chaindata"
	ancientDir := chainDataDir + "/ancient"
	db, err := rawdb.Open(
		rawdb.OpenOptions{
			Directory:         chainDataDir,
			AncientsDirectory: ancientDir,
			Ephemeral:         true,
		},
	)
	if err != nil {
		print("👎Open rawdb fail!", err)
	}

	//用读取的数据新建数据链
	bc, _ := core.NewBlockChain(db, core.DefaultCacheConfigWithScheme(rawdb.HashScheme), nil, nil, ethash.NewFaker(), vm.Config{}, nil, nil)

	//读取特定的区块
	//var blockNumber uint64 = 9800644
	var blockNumber uint64 = 9833300 //包含创建合约的 Transaction (TODO:需要特殊处理不然报错)
	//var blockNumber uint64 = 9831292                              // Nice Picture
	blockHash := rawdb.ReadCanonicalHash(db, blockNumber)         //当前选取的区块 Hash
	parentBlockHash := rawdb.ReadCanonicalHash(db, blockNumber-1) //父区块 Hash
	block := rawdb.ReadBlock(db, blockHash, blockNumber)
	parentBlock := rawdb.ReadBlock(db, parentBlockHash, blockNumber-1)
	if block == nil || parentBlock == nil {
		print("👎Read block or parent block fail!", err)
	}

	//用父区块获得当前区块执行前的区块链全局状态
	parentBlockRoot := parentBlock.Root()
	stateDb, err := bc.StateAt(parentBlockRoot)
	if err != nil {
		print("👎Get State fail!", err)
	}

	//ReadBlockTx(block, db, core.DefaultCacheConfigWithScheme(rawdb.HashScheme))

	_, _, usedGas, err := bc.Processor().Process(block, stateDb, vm.Config{})
	if err != nil {
		print("👎Blockchain process fail!", err)
	}
	print("Gas Used: ", usedGas)

	//OutputBlockHookInfo()
}

func main() {
	fmt.Print("DoProcess()\n")
	DoProcess()
	fmt.Print("\n\n")

	// fmt.Print("\n\nGetGraphDemo()\n")
	// GetGraphDemo("/home/user/data/Brian/brian_eth_runner/go_runner/output", "demo")

	// fmt.Print("\n\nOutputGraph()\n")
	// parallel.OutputGraph()

	// fmt.Print("BuildGraph()\n")
	// var graph *parallel.Graph = parallel.BuildDependencyGraph()
	// //GetGraphFromRelationship(graph, "/home/user/data/Brian/brian_eth_runner/go_runner/output", "demo")
	// GetGraphFromRelationship(graph, "/home/user/data/Brian/brian_eth_runner/go_runner/output", "demo")
	// fmt.Print("\n\n")

	fmt.Print("BuildGraph()\n")
	parallel.BuildTxRelationGraph()
	fmt.Print("\n\n")
}
