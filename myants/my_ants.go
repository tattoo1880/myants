package myants

import (
	"context"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"golang.org/x/time/rate"
	"log"
)

type SizeAble interface {
	Len() int
}

type TaskFunc[T SizeAble, R any] func(ctx context.Context, param T, page int) (R, error)

type MyAnts[T SizeAble, R any] interface {
	UseMyAnts(task TaskFunc[T, R], param T) (chan R, error)
}

type MyAntsImpl[T SizeAble, R any] struct {
	MyPool      *ants.Pool
	RateLimiter *rate.Limiter
	Ctx         context.Context
}

func NewMyAnts[T SizeAble, R any](size int, hz int, ctx context.Context) (MyAnts[T, R], error) {

	mypool, err := ants.NewPool(size)
	if err != nil {
		log.Println("创建pool失败:", err.Error())
		return nil, err

	}

	limiter := rate.NewLimiter(rate.Limit(hz), 1)
	//todo return

	return &MyAntsImpl[T, R]{
		MyPool:      mypool,
		RateLimiter: limiter,
		Ctx:         ctx,
	}, nil

}

func (m *MyAntsImpl[T, R]) UseMyAnts(task TaskFunc[T, R], param T) (chan R, error) {

	//todo 声明result 为结果的 泛型

	result := make(chan R, 10)

	//todo 判断T的长度,如果长度为0,则直接返回空的结果
	if param.Len() == 0 {
		close(result)
		return result, nil
	}

	//var wg sync.WaitGroup
	fmt.Println("开始执行任务,总页数:", param.Len())
	for i := 0; i < param.Len(); i++ {

		page := i

		//todo 间隔
		err := m.RateLimiter.Wait(m.Ctx)
		if err != nil {
			log.Printf("Rate limiter error: %v", err)
			continue
		}

		//wg.Add(1)
		runerror := m.MyPool.Submit(func() {

			//defer wg.Done()
			r, err1 := task(m.Ctx, param, page)
			if err1 != nil {
				log.Println(err)
				return
			}
			fmt.Println("page", page, "is done with result:", r)

			select {
			case <-m.Ctx.Done():
				log.Println("上下文已取消，停止发送结果")
				return
			case result <- r:
				log.Println("结果已发送到通道:", r)

			}

		})

		if runerror != nil {
			//wg.Done()
			return nil, runerror
		}
	}

	go func() {
		err := m.MyPool.ReleaseContext(m.Ctx)
		if err != nil {
			log.Println("等待多有完成发生错误", err)
			return
		} // 等所有 worker 执行完毕
		close(result)
	}()

	return result, nil

}
