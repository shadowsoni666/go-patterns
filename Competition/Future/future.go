package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Future interface {
	Result() (string, error)
}

type InnerFuture struct {
	once  sync.Once
	wg    sync.WaitGroup
	res   string
	err   error
	resCh <-chan string
	errCh <-chan error
}

func (f *InnerFuture) Result() (string, error) {
	f.once.Do(func() {
		f.wg.Add(1)
		defer f.wg.Done()
		f.res = <-f.resCh
		f.err = <-f.errCh
	})
	f.wg.Wait()
	return f.res, f.err
}

// SlowFunction – это обертка для указанной функции, которую требуется выполнить асинхронно.
// Ее задача – создать каналы результатов, запустить за данную функцию в сопрограмме,
// а также создать и вернуть реализацию Future (InnerFuture в этом примере
func SlowFunction(ctx context.Context) Future {
	resCh := make(chan string)
	errCh := make(chan error)
	go func() {
		select {
		case <-time.After(time.Second * 2):
			resCh <- "I slept for 2 seconds"
			errCh <- nil
		case <-ctx.Done():
			resCh <- ""
			errCh <- ctx.Err()
		}
	}()
	return &InnerFuture{resCh: resCh, errCh: errCh}
}
func main() {
	ctx := context.Background()
	future := SlowFunction(ctx)
	res, err := future.Result()
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(res)
}

//Такой подход обеспечивает простой и понятный интерфейс.
//Программист
//может создать экземпляр Future, обращаться к нему и даже применять таймауты или крайние сроки с помощью контекста Context
