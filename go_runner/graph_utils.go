package main

import (
	"bytes"
	"os"
	"os/exec"
	"reflect"
	"strconv"
)

// 调用系统命令行的方法
func Cmd(s string) {
	cmd := exec.Command(`/bin/sh`, `-c`, s)
	var out bytes.Buffer
	err := cmd.Run()
	if err != nil {
		print("Cmd error", err)
	}
	print("Cmd out: ", out)
}

// Table cell 结构体（row 里面包含许多 cell）（cell	的标签是<td>）
type TableCell struct {
	Content  string
	CellAttr []string
	FontAttr []string
}

// 添加TableCell 的属性
func (cell *TableCell) AddAttr(attr string, value string) {
	cell.CellAttr = append(cell.CellAttr, attr+"=\""+value+"\"")
}

// 添加 TableCell font 的属性
func (cell *TableCell) AddFontAttr(attr string, value string) {
	cell.FontAttr = append(cell.FontAttr, attr+"=\""+value+"\"")
}

// Table row 结构体（row 的标签是<tr>）
type TableRow struct {
	CellList []TableCell
}

// 往 Table 行中添加节点
func (row *TableRow) AddCell(cell TableCell) {
	row.CellList = append(row.CellList, cell)
}

// 节点的 lable 属性可能会用到的表的结构
type Table struct {
	TableAttr []string
	RowList   []TableRow
}

// Table 添加属性的函数
func (table *Table) AddAttr(addr string, value string) {
	table.TableAttr = append(table.TableAttr, addr+"=\""+value+"\"")
}

// Table 添加行的函数
func (table *Table) AddRow(row TableRow) {
	table.RowList = append(table.RowList, row)
}

// 将 Table 转化文文本
func (table *Table) toString() string {
	var ret string = ""
	ret += "<table"

	if len(table.TableAttr) > 0 {
		for _, attr := range table.TableAttr {
			ret += " " + attr
		}
	}
	ret += ">"

	if len(table.RowList) > 0 {
		for _, row := range table.RowList {
			ret += "<tr>"
			if len(row.CellList) > 0 {
				for _, cell := range row.CellList {
					ret += "<td"
					if len(cell.CellAttr) > 0 {
						for _, cellAttr := range cell.CellAttr {
							ret += " " + cellAttr
						}
					}
					ret += "><font"
					if len(cell.FontAttr) > 0 {
						for _, fontAttr := range cell.FontAttr {
							ret += " " + fontAttr
						}
					}
					ret += ">" + cell.Content
					ret += "</font></td>"
				}
			}
			ret += "</tr>"
		}

	}

	ret += "</table>"
	return ret
}

// 图的节点
type Node struct {
	NodeName string
	NodeAttr []string
}

// 添加节点属性函数
func (n *Node) AddAttr(attr string, value interface{}) {
	if reflect.TypeOf(value).String() == "string" {
		if value.(string)[0] == '<' { //判断是否为 HTML 格式标签
			n.NodeAttr = append(n.NodeAttr, attr+" =<"+value.(string)+"> ")
		} else {
			n.NodeAttr = append(n.NodeAttr, attr+" = \""+value.(string)+"\" ")
		}
	} else if reflect.TypeOf(value).String() == "int" {
		n.NodeAttr = append(n.NodeAttr, attr+" = "+strconv.Itoa(value.(int))+" ")
	} else {
		print("Node AddAddr value type error")
	}
}

// 图的边
type Edge struct {
	From     string
	To       string
	lineType string
	EdgeAttr []string
}

// 添加边属性
func (edge *Edge) AddAttr(attr string, value interface{}) {
	edge.EdgeAttr = append(edge.EdgeAttr, attr+" = \""+value.(string)+"\" ")
}

// 图的结构体
type Graph struct {
	GraphName string
	Attribute []string
	NodeAttr  []string
	EdgeAttr  []string
	GraphAttr []string
	NodeList  []Node
	EdgeList  []Edge
}

// 给图添加节点
func (g *Graph) AddNode(node Node) {
	g.NodeList = append(g.NodeList, node)
}

// 给图添加边
func (g *Graph) AddEdge(edge Edge) {
	g.EdgeList = append(g.EdgeList, edge)
}

func (g *Graph) toDOT() string {
	var dot string = ""
	dot = "digraph " + g.GraphName + " {\n"

	if len(g.Attribute) > 0 {
		for _, attribute := range g.Attribute {
			dot += "\t" + attribute + "\n"
		}
	}

	if len(g.NodeAttr) > 0 {
		dot += "\tnode ["
		for _, nodeAttr := range g.NodeAttr {
			dot += nodeAttr + " "
		}
		dot += "];\n"
	}

	if len(g.EdgeAttr) > 0 {
		dot += "\tedge ["
		for _, edgeAttr := range g.EdgeAttr {
			dot += edgeAttr + " "
		}
		dot += "];\n"
	}

	//Add GraphAttr
	if len(g.GraphAttr) > 0 {
		dot += "\tgraph ["
		for _, graphAttr := range g.GraphAttr {
			dot += graphAttr + " "
		}
		dot += "];\n"
	}

	//Add Node
	if len(g.NodeList) > 0 {
		for _, node := range g.NodeList {
			// dot += "\t\"" + node.NodeName + "\" "
			dot += "\t" + node.NodeName + " " //no ""
			if len(node.NodeAttr) == 0 {
				dot += "\n"
			} else {
				dot += "["
				for _, nodeAttr := range node.NodeAttr {
					dot += nodeAttr + " "
				}
				dot += "];\n"
			}
		}
	}

	//Add Edge
	if len(g.EdgeList) > 0 {
		for _, edge := range g.EdgeList {
			// dot += "\t\"" + edge.From + "\" " + edge.lineType + "\"" + edge.To + "\" "
			dot += "\t" + edge.From + " " + edge.lineType + " " + edge.To + " " //no ""
			if len(edge.EdgeAttr) == 0 {
				dot += "\n"
			} else {
				dot += "["
				for _, edgeAttr := range edge.EdgeAttr {
					dot += edgeAttr + " "
				}
				dot += "];\n"
			}
		}
	}

	dot += "}"
	return dot
}

// 绘图方法
func (g *Graph) Draw(path string, fileName string) error {
	//返回 DOT 格式文本
	dot := g.toDOT()
	filePath := path + "/" + fileName + ".gv"
	imagePath := path + "/" + fileName + ".png"
	print("Output File Path: ", filePath)
	print("Output Image Path: ", imagePath)

	//create file and write file
	file, err := os.Create(filePath)
	if err != nil {
		print("Create file error", err)
	}
	defer file.Close()

	//写入 Dot 格式文本
	_, err = file.Write([]byte(dot))
	if err != nil {
		print("Write file error", err)
	}

	//创建图片文件
	_, err = os.Create(imagePath)
	if err != nil {
		print("Create imageFile error", err)
	}
	defer file.Close()
	Cmd("dot " + filePath + " -T png -o " + imagePath) //调用系统程序生成 png
	//print("dot " + filePath + " -T png -o " + imagePath)

	return nil
}
