package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"myants/myants"
	"strconv"
	"time"
)

func main() {

	timeout, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()

	a, err := myants.NewMyAnts[int, string](5, 2, timeout)
	if err != nil {
		panic(err)
	}

	taskfn := myants.TaskFunc[int, string](func(ctx context.Context, param int, page int) (string, error) {
		fmt.Println(page)
		//todo 模拟任务
		select {
		case <-ctx.Done():
			return "nil", ctx.Err()
		default:
			fmt.Println("Task", param, "page", page, "is running")
			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
			fmt.Println("Task", param, "page", page, "is done")
			return "Task " + strconv.Itoa(param) + " page " + strconv.Itoa(page) + " is done", nil

		}
	})

	result, err := a.UseMyAnts(taskfn, 1)

	if err != nil {
		log.Println(err)
		panic(err)
	}

	fmt.Println("All tasks submitted, waiting for results...", result)

	for res := range result {
		log.Println(res)
	}

}
