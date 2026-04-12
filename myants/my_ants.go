package myants

import (
	"context"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"golang.org/x/time/rate"
	"log"
	"sync"
)

type TaskFunc[T any, R any] func(ctx context.Context, param T, page int) (R, error)

type MyAnts[T any, R any] interface {
	UseMyAnts(task TaskFunc[T, R], param T) (chan R, error)
}

type MyAntsImpl[T any, R any] struct {
	MyPool      *ants.Pool
	RateLimiter *rate.Limiter
	Ctx         context.Context
}

func NewMyAnts[T any, R any](size int, hz int, ctx context.Context) (MyAnts[T, R], error) {

	mypool, err := ants.NewPool(size)
	if err != nil {
		log.Println("创建pool失败:", err.Error())
		return nil, err

	}

	limiter := rate.NewLimiter(rate.Limit(hz), hz)
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

	defer m.MyPool.Release()

	//jobdone := make(chan struct{}, 10)

	var wg sync.WaitGroup
	var mu sync.Mutex
	for i := 0; i < 10; i++ {

		page := i

		//todo 间隔
		err := m.RateLimiter.Wait(m.Ctx)
		if err != nil {
			log.Printf("Rate limiter error: %v", err)
			continue
		}

		runerror := m.MyPool.Submit(func() {

			wg.Go(func() {
				r, err2 := task(m.Ctx, param, page)
				if err2 != nil {
					log.Println(err2)
				}
				mu.Lock()
				defer mu.Unlock()
				fmt.Println("Task", param, "page", page, "is done with result:", r)
				result <- r

			})
		})

		if runerror != nil {
			return nil, runerror
		}
	}

	wg.Wait()
	close(result)

	return result, nil

}
