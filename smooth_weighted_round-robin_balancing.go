//smooth weighted round-robin balancing
//phusion /nginx 
//https://github.com/phusion/nginx/commit/27e94984486058d73157038f7950a0a36ecc6e35

package main

import (
	"fmt"
)


type WeightNode struct{
	Name string
	ConstWeight int//配置的不变的权重
	CurrWeight int //当前权重,每轮选择会发生变化
}

func SmoothWeightRound(nodes []WeightNode) (index int){
	var totalWeight = 0			//权重之和（就是初始值所有权重之和）
	var maxCurrWeight = 0		//最大权重
	var maxCurrWeightIndex=0	//最大权重对应的索引
	//每次选出权值最大的那个
	for  i:=0;i<len(nodes);i++{
		curNode := &nodes[i]
		curNode.CurrWeight +=curNode.ConstWeight//加上自己的恒定权重
		if curNode.CurrWeight >= maxCurrWeight {
			maxCurrWeight = curNode.CurrWeight
			maxCurrWeightIndex = i
		}
		totalWeight += curNode.ConstWeight
	}
	//最大值节点要减掉所有权重之和
	nodes[maxCurrWeightIndex].CurrWeight -= totalWeight
	index = maxCurrWeightIndex
	return
}
func testit(){
	var nodes []WeightNode= []WeightNode{
		{Name:"aaa",ConstWeight:4,CurrWeight:0},
		{Name:"bbb",ConstWeight:3,CurrWeight:0},
		{Name:"ccc",ConstWeight:2,CurrWeight:0},
		{Name:"ddd",ConstWeight:1,CurrWeight:0},
		//{Name:"eee",ConstWeight:11,CurrWeight:0},
		}
	for i:=0;i<15;i++{
		var index = SmoothWeightRound(nodes)
		var curNode = nodes[index]
		fmt.Println("当前:",index,curNode.Name)
	}
}

