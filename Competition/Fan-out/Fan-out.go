package Fan_out

import (
	"fmt"
	"sync"
)

func Split(source <-chan int, n int) []<-chan int {
	dests := make([]<-chan int, 0) // Создать срез dests
	for i := 0; i < n; i++ {       // Создать n выходных каналов
		ch := make(chan int)
		dests = append(dests, ch)
		go func() { // Каждый выходной канал передается
			defer close(ch) // своей сопрограмме, которая состязается
			// с другими за доступ к source
			for val := range source {
				ch <- val
			}
		}()
	}
	return dests
}

func main() {
	source := make(chan int)  // Входной канал
	dests := Split(source, 5) // Получить 5 выходных каналов
	go func() {               // Передать числа 1..10 в source
		for i := 1; i <= 10; i++ { // и закрыть его по завершении
			source <- i
		}
		close(source)
	}()
	var wg sync.WaitGroup // Использовать WaitGroup для ожидания, пока
	wg.Add(len(dests))    // не закроются выходные каналы
	for i, ch := range dests {
		go func(i int, d <-chan int) {
			defer wg.Done()
			for val := range d {
				fmt.Printf("#%d got %d\n", i, val)
			}
		}(i, ch)
	}
	wg.Wait()
}
