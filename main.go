package main

import (
	"context"
	"fmt"
	"github.com/tattoo1880/myants/myants"
	"log"
	"math/rand"
	"strconv"
	"time"
)

type InFunc struct {
	//todo 传入参数是一个map 的 切片
	DictName []map[int]string
}

func (inf InFunc) Len() int {
	return len(inf.DictName)
}

type OutFunc struct {
	//返回值是一个string
	string
}

func main() {

	timeout, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()

	a, err := myants.NewMyAnts[InFunc, OutFunc](5, 500, timeout)
	if err != nil {
		panic(err)
	}

	taskfn := myants.TaskFunc[InFunc, OutFunc](func(ctx context.Context, param InFunc, page int) (OutFunc, error) {
		//todo index 为page ,找到下表的value
		fmt.Println("正在获取的下标为:", page)
		select {
		case <-ctx.Done():
			return OutFunc{}, ctx.Err()
		default:
			//todo 将下标的value 进行处理,并返回结果
			key := param.DictName[page]
			value := param.DictName[page]

			return OutFunc{string: "处理结果: " + strconv.Itoa(page) + " - " + fmt.Sprintf("%v", key) + " - " + fmt.Sprintf("%v", value) + " - " + time.Now().Format(time.RFC3339) + " - 随机数: " + strconv.Itoa(rand.Intn(100))}, nil
		}
	})

	//todo example map

	inmap := new(InFunc)

	inmap.DictName = []map[int]string{
		{1: "张飞"},
		{2: "李逵"},
		{3: "驴x x"},
		{4: "地雷"},
		{5: "吕布"},
		{5: "吕布"},
		{5: "吕布"},
		{5: "吕布"},
		{5: "吕布"},
		{5: "吕布"},
	}

	result, err := a.UseMyAnts(taskfn, *inmap)

	if err != nil {
		log.Println(err)
		panic(err)
	}

	fmt.Println("All tasks submitted, waiting for results...", &result)

	for res := range result {
		log.Println(res)
	}

}
