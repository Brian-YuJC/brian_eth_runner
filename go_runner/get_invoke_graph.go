package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/parallel"
)

// 实例的总数（Tx addr）用于给实例编号创建实例和编号的映射表
var instanceNum int = 0

// 实例编号和地址的映射表
var num2Addr map[int]string = make(map[int]string)

// 实例地址和编号的映射表
var addr2Num map[string]int = make(map[string]int)

// 维护映射表的函数
func addInstance(addr string) {
	if _, ok := addr2Num[addr]; !ok { //如果当前地址不在映射表中(未注册)
		instanceNum++
		addr2Num[addr] = instanceNum
		num2Addr[instanceNum] = addr
	}
}

// 往图中添加账号节点
func addAccountNode(addr string, graph *Graph) {
	if _, ok := addr2Num[addr]; !ok { //如果当前地址不在映射表中(未注册)，就注册该地址并添加节点
		addInstance(addr) // 注册节点
		node := Node{NodeName: "port_" + strconv.Itoa(addr2Num[addr])}
		node.AddAttr("style", "filled")
		node.AddAttr("shape", "Mrecord")
		node.AddAttr("penwidth", 1)
		node.AddAttr("fillcolor", "grey")
		node.AddAttr("fontname", "Courier New")
		node.AddAttr("label", "Account Address: "+addr)
		graph.AddNode(node)
	}
}

// 往图里添加边的封装函数
func addEdge(from string, to string, lineType string, label string, color string, graph *Graph) {
	edge := Edge{From: from, To: to, lineType: lineType}
	if len(label) > 0 {
		edge.AddAttr("label", label)
	}
	edge.AddAttr("color", color)
	graph.AddEdge(edge)
}

// 绘图方法
func GetGraphDemo(path string, fileName string) {

	blockInfo := parallel.GetBlockInfo() //获取 blockInfo 对象

	//	新建一张图并初始化
	graph := Graph{GraphName: "G"}
	graph.GraphAttr = append(graph.GraphAttr, "fontsize=30 labelloc=\"t\" label=\"\" splines=true overlap=false rankdir = \"LR\"")

	//获取块中的交易信息
	for i, tx := range blockInfo.Tx {

		//设置交易的图像节点
		node := Node{NodeName: "port_tx" + fmt.Sprintf("%d", i)}
		node.AddAttr("style", "filled")
		node.AddAttr("shape", "Mrecord")
		node.AddAttr("penwidth", 1)
		node.AddAttr("fillcolor", "white")
		node.AddAttr("fontname", "Courier New")

		//创建并初始化表格
		txTable := Table{}
		txTable.AddAttr("border", "0")
		txTable.AddAttr("cellborder", "0")
		txTable.AddAttr("cellpadding", "3")
		txTable.AddAttr("bgcolor", "white")

		//设置交易名称的cell
		row1 := TableRow{}
		txName := fmt.Sprintf("TX_%d", i)
		txNameCell := TableCell{Content: txName}
		txNameCell.AddAttr("bgcolor", "black")
		txNameCell.AddAttr("colspan", "2")
		txNameCell.AddFontAttr("color", "white")
		row1.AddCell(txNameCell)

		//设置 From To Value的cell
		row2 := TableRow{}
		from := tx.From.Hex()
		to := tx.To
		value := tx.Value.String()

		txInfoCell := TableCell{Content: "<b>From: </b>" + from + "<br/>" + "<b>To: </b>" + to + "<br/><b>Value: </b>" + value}
		txInfoCell.AddAttr("bgcolor", "white")
		//txInfoCell.AddAttr("align", "left")
		txInfoCell.AddAttr("colspan", "2")
		txInfoCell.AddFontAttr("color", "black")
		row2.AddCell(txInfoCell)

		//Node中添加表格标签，然后将节点加入图
		txTable.AddRow(row1)
		txTable.AddRow(row2)
		node.AddAttr("label", txTable.toString())
		graph.AddNode(node)

		//接下来开始画调用依赖
		//维护一个箭头表（address->read or write）,在表上就说明Transaction读或写了此账户(bool1:isRead? bool2:isWrite? bool3:isCreate?)
		edgeMap := make(map[string][3]bool)

		//Transaction 的图像节点标识符
		txPort := "port_tx" + fmt.Sprintf("%d", i)

		//Transaction调用的第一个地址From，肯定会读取和改变其 Nonce 所以有依赖关系
		addAccountNode(from, &graph)               //画上 from 节点
		edgeMap[from] = [3]bool{true, true, false} //添加进箭头表

		//处理创建合约的特殊情况
		if tx.To == "nil" {
			addAccountNode(tx.NewContractAddr.Hex(), &graph)
			edgeMap[tx.NewContractAddr.Hex()] = [3]bool{false, false, true}
		}

		//不管有无调用其他合约，只要 Transaction 的 value 不为空就会进行转账操作(创建合约的情况前面处理过了)
		if tx.Value.Sign() > 0 && tx.To != "nil" { //	value > 0 需要转账
			addAccountNode(to, &graph) //画上 to 节点
			edgeMap[to] = [3]bool{true, true, false}
		}

		if len(tx.CallQueue) > 0 { //有合约调用情况

			for _, contractInfo := range tx.CallQueue { //进入每个调用过程的循环，如果调用的合约有对自身的读写操作，则 tx 图像实例直接指向它

				//是否有读写的标识
				doRead := false
				doWrite := false

				for _, keyOpcode := range contractInfo.KeyOpcode {
					split := strings.Split(keyOpcode, " ")
					opcode := split[1]
					if opcode == "BALANCE" || opcode == "SELFBALANCE" { //opcode 为BALANCE需要记录BALANCE访问的地址，因为涉及读操作
						addAccountNode(split[2], &graph)        //加入新的图节点
						if value, ok := edgeMap[split[2]]; ok { //	判断防止之前访问过的记录被覆盖
							value[0] = true
							edgeMap[split[2]] = value
						} else {
							edgeMap[split[2]] = [3]bool{true, false, false}
						}
					}
					if opcode == "SLOAD" {
						doRead = true
					}
					if opcode == "SSTORE" {
						doWrite = true
					}
					if opcode == "CREATE" || opcode == "CREATE2" { //opcode为 create or create2 则新建新创建的合约的图节点，并加入 edgemap箭头标签为“create”，
						addAccountNode(split[2], &graph) //加入新的图节点
						edgeMap[split[2]] = [3]bool{false, false, true}
						if split[3] == "doTransfer_true" { //create 有转账发生
							doRead = true
							doWrite = true
						}
					}
					if opcode == "CALL" {
						if split[3] == "doTransfer_true" { //需要转账
							addAccountNode(split[2], &graph)
							if value, ok := edgeMap[split[2]]; ok { //	判断防止之前访问过的记录被覆盖(如果不是有 create 标记这里也不用判断，因为 read write 都为true)
								value[0] = true
								value[1] = true
								edgeMap[split[2]] = value
							} else {
								edgeMap[split[2]] = [3]bool{true, true, false}
							}
							doRead = true
							doWrite = true
						}
					}
					if opcode == "SELFDESTRUCT" {
						addAccountNode(split[2], &graph)
						if value, ok := edgeMap[split[2]]; ok { //	判断防止之前访问过的记录被覆盖(如果不是有 create 标记这里也不用判断，因为 read write 都为true)
							value[0] = true
							value[1] = true
							edgeMap[split[2]] = value
						} else {
							edgeMap[split[2]] = [3]bool{true, true, false}
						}
						doRead = true
						doWrite = true
					}
				}

				//判断箭头上应该标注什么信息，并判断当前调用到的地址是否是被转账的地址
				label := [3]bool{false, false, false}
				if doRead && doWrite {
					label = [3]bool{true, true, false}
				} else if doRead && !doWrite {
					label = [3]bool{true, false, false}
				} else if !doRead && doWrite {
					label = [3]bool{false, true, false}
				} else { //如果该调用合约没有涉及读或写则继续
					continue
				}

				//如果涉及读写则加入图节点
				addAccountNode(contractInfo.ContractAddr.Hex(), &graph)
				if value, ok := edgeMap[contractInfo.ContractAddr.Hex()]; ok { //	判断防止之前访问过的记录被覆盖
					//只要有 1 就是 1
					edgeMap[contractInfo.ContractAddr.Hex()] = [3]bool{value[0] || label[0], value[1] || label[1], value[2] || label[2]}
				} else {
					edgeMap[contractInfo.ContractAddr.Hex()] = label
				}

			}
		}

		//根据 EdgeMap 往图里添加边
		for addr, value := range edgeMap {
			var label string
			if value[2] { //要先判断 create 不然会覆盖掉
				label = "[Create]"
			} else if value[0] && value[1] {
				label = "[Read & Write]"
			} else if !value[0] && value[1] {
				label = "[Write]"
			} else if value[0] && !value[1] {
				label = "[Read]"
			} else {
				print("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
			}
			addEdge(txPort, "port_"+strconv.Itoa(addr2Num[addr]), "->", label, "black", &graph)
		}
	}

	//让 Transaction 按顺序连接
	// for i := 0; i < len(blockInfo.Tx)-1; i++ {
	// 	fromPort := "port_tx" + fmt.Sprintf("%d", i)
	// 	toPort := "port_tx" + fmt.Sprintf("%d", i+1)
	// 	addEdge(fromPort, toPort, "->", "", "grey", &graph)
	// }
	graph.Draw(path, fileName)

}

// func graphTest() {
// 	Test Graph
// 	graph := Graph{GraphName: "G"}
// 	graph.GraphAttr = append(graph.GraphAttr, "fontsize=30 labelloc=\"t\" label=\"\" splines=true overlap=false rankdir = \"LR\"")
// 	node1 := Node{NodeName: "a"}
// 	node2 := Node{NodeName: "b"}
// 	node3 := Node{NodeName: "c"}
// 	edge1 := Edge{From: node1.NodeName, To: node2.NodeName, lineType: "->"}
// 	edge2 := Edge{From: node1.NodeName, To: node3.NodeName, lineType: "->"}
// 	edge3 := Edge{From: node2.NodeName, To: node3.NodeName, lineType: "->"}
// 	graph.AddNode(node1)
// 	graph.AddNode(node2)
// 	graph.AddNode(node3)
// 	graph.AddEdge(edge1)
// 	graph.AddEdge(edge2)
// 	graph.AddEdge(edge3)
// 	//fmt.Print(graph.toDOT())
// 	graph.Draw(path, fileName)
// }

// func tableTest() string {
// 	table := Table{}
// 	table.AddAttr("border", "0")
// 	table.AddAttr("cellborder", "0")
// 	table.AddAttr("cellpadding", "3")
// 	table.AddAttr("bgcolor", "white")
// 	row := TableRow{}
// 	cell := TableCell{}
// 	cell.AddAttr("bgcolor", "black")
// 	cell.AddAttr("align", "center")
// 	cell.AddAttr("colspan", "2")
// 	cell.AddFontAttr("color", "white")
// 	cell.Content = "Test Test"
// 	row.AddCell(cell)
// 	table.AddRow(row)
// 	row = TableRow{}
// 	cell = TableCell{}
// 	cell.AddAttr("bgcolor", "red")
// 	cell.AddAttr("align", "center")
// 	cell.AddAttr("colspan", "2")
// 	cell.AddFontAttr("color", "black")
// 	cell.Content = "Test Test"
// 	cell2 := TableCell{}
// 	cell2.AddAttr("bgcolor", "black")
// 	cell2.AddAttr("align", "center")
// 	cell2.AddAttr("colspan", "2")
// 	cell2.AddFontAttr("color", "white")
// 	cell2.Content = "Test Test"
// 	row.AddCell(cell2)
// 	row.AddCell(cell)
// 	row2 := TableRow{}
// 	row2.AddCell(cell2)
// 	row2.AddCell(cell)
// 	table.AddRow(row)
// 	table.AddRow(row2)
// 	s := table.toString()

// 	print(s)
// 	return s
// }
