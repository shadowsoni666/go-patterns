package Fan_In

import (
	"fmt"
	"sync"
	"time"
)

func Funnel(sources ...<-chan int) <-chan int {
	dest := make(chan int) // Общий выходной канал
	var wg sync.WaitGroup  // Для автоматического закрытия dest,
	// когда закроются все входящие каналы sources
	wg.Add(len(sources))         // Установить размер WaitGroup
	for _, ch := range sources { // Запуск сопрограммы для каждого входного канала
		go func(c <-chan int) {
			defer wg.Done() // Уведомить WaitGroup, когда c закроется
			for n := range c {
				dest <- n
			}
		}(ch)
	}
	go func() { // Запустить сопрограмму, которая закроет dest
		wg.Wait() // после закрытия всех входных каналов
		close(dest)
	}()
	return dest
}

func main() {
	sources := make([]<-chan int, 0) // Создать пустой срез с каналами
	for i := 0; i < 3; i++ {
		ch := make(chan int)
		sources = append(sources, ch) // Создать канал; добавить в срез sources
		go func() {                   // Запустить сопрограмму для каждого
			defer close(ch) // Закрыть канал по завершении сопрограммы
			for i := 1; i <= 5; i++ {
				ch <- i
				time.Sleep(time.Second)
			}
		}()
	}
	dest := Funnel(sources...)
	for d := range dest {
		fmt.Println(d)
	}
}
