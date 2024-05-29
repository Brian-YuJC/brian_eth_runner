package main

import (
	"fmt"

	"github.com/ethereum/go-ethereum/parallel"
)

//--------------------------------------------------------------------------------------
//本文件使用parallel/get_relationship.go返回的区块内部交易与账号之间的读写关系图来进行绘图
//运行完 DoProcess、parallel.OutputGraph 后调用GetGraphFromRelationship
//--------------------------------------------------------------------------------------

// 操作和边颜色的 Map
var colorMap map[string]string = map[string]string{
	"Read":         "green",
	"Write":        "cyan",
	"Create":       "pink",
	"Read & Write": "blue",
	"SelfDestruct": "red",
	"Transfer":     "black",
}

// 往图里添加边的封装函数
func addEdge2Graph(from string, to string, lineType string, label string, color string, graph *Graph) {
	edge := Edge{From: from, To: to, lineType: lineType}
	if len(label) > 0 {
		edge.AddAttr("label", label)
	}
	edge.AddAttr("color", color)
	graph.AddEdge(edge)
}

func GetGraphFromRelationship(g *parallel.Graph, path string, fileName string) {
	//获取图的节点和边的信息
	txList := g.TxNodeList
	accountNodeList := g.AccountNodeList
	edgeList := g.EdgeList

	//新建图
	graph := Graph{GraphName: "G"}
	graph.GraphAttr = append(graph.GraphAttr, "fontsize=30 labelloc=\"t\" label=\"\" splines=true overlap=false rankdir = \"LR\" ordering=\"in\"")

	//添加交易节点
	for _, tx := range txList {

		//设置交易的图像节点
		node := Node{NodeName: "port_tx" + fmt.Sprintf("%d", tx.ID)}
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
		txName := fmt.Sprintf("Tx_%d", tx.ID)
		txNameCell := TableCell{Content: txName}
		txNameCell.AddAttr("bgcolor", "black")
		txNameCell.AddAttr("colspan", "2")
		txNameCell.AddFontAttr("color", "white")
		row1.AddCell(txNameCell)

		//设置 From To Value的cell
		row2 := TableRow{}
		from := tx.From
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

	}

	//添加账户节点
	for _, account := range accountNodeList {
		node := Node{NodeName: "port_account" + account.Address}
		node.AddAttr("style", "filled")
		node.AddAttr("shape", "Mrecord")
		node.AddAttr("penwidth", 1)
		node.AddAttr("fillcolor", "grey")
		node.AddAttr("fontname", "Courier New")
		node.AddAttr("label", "Account Address: "+account.Address)
		graph.AddNode(node)
	}

	//添加边
	for _, edge := range edgeList {
		addEdge2Graph("port_tx"+edge.From, "port_account"+edge.To, "->", edge.Op, colorMap[edge.Op], &graph)
	}

	//print(graph.toDOT())
	graph.Draw(path, fileName)

}

// 只画出关联两个以上 Transaction 的 Account
func GetGraph_RelatedAccount(g *parallel.Graph, path string, fileName string) {

	//获取图的节点和边的信息
	txList := g.TxNodeList
	accountNodeList := g.AccountNodeList
	edgeList := g.EdgeList

	//维护一个Account 和 Transaction 关连的 Map
	accountTxMap := make(map[string]map[string]bool)

	//初始化 Map
	for _, account := range accountNodeList {
		accountTxMap[account.Address] = make(map[string]bool)
	}

	//如果一个 Transaction 指向一个 Account，那么这个 Transaction 与这个 Account 有关连，加入 Map
	for _, edge := range edgeList {
		from := edge.From
		to := edge.To
		accountTxMap[to][from] = true
	}

	//根据 Transaction 和 Account 的关连 Map，只保留与多个 Transaction 相连的 Account
	newAccountNodeList := []parallel.AccountNode{}
	for _, account := range accountNodeList {
		if len(accountTxMap[account.Address]) > 1 { //说明 Account 与多个 Transaction 有关连
			newAccountNodeList = append(newAccountNodeList, account)
		}
	}

	//根据 Transaction 和 Account 的关连 Map，只保留与多个 Transaction 相连的 Account 有关的边
	newEdgeList := []parallel.Edge{}
	for _, edge := range edgeList {
		if len(accountTxMap[edge.To]) > 1 { //说明Edge的 Account 与多个 Transaction 有关连
			newEdgeList = append(newEdgeList, edge)
		}
	}

	//构成新的关系图
	newGraph := parallel.Graph{}
	newGraph.TxNodeList = txList
	newGraph.AccountNodeList = newAccountNodeList
	newGraph.EdgeList = newEdgeList

	//根据新关系图绘制图片
	GetGraphFromRelationship(&newGraph, path, fileName)
}
