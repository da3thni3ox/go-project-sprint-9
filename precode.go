package main

import (
	"context"
	"fmt"
	"log"
	"sync"
)

func Generator(ctx context.Context, ch chan<- int64, fn func(int64)) {
	for i := int64(1); ; i++ {
		select {
		case <-ctx.Done(): // Проверка, не отменен ли контекст
			return
		case ch <- i: // Отправляем число в канал
			fn(i) // Вызываем функцию для подсчета
		}
	}
}

func Worker(in <-chan int64, out chan<- int64) {
	defer close(out)
	for num := range in {
		out <- num
	}
}

func main() {
	chIn := make(chan int64)

	// 3. Создание контекста
	// ...
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// для проверки будем считать количество и сумму отправленных чисел
	var inputSum int64   // сумма сгенерированных чисел
	var inputCount int64 // количество сгенерированных чисел

	// генерируем числа, считая параллельно их количество и сумму
	go Generator(ctx, chIn, func(i int64) {
		inputSum += i
		inputCount++
	})

	const NumOut = 5 // количество обрабатывающих горутин и каналов
	// outs — слайс каналов, куда будут записываться числа из chIn
	outs := make([]chan int64, NumOut)
	for i := 0; i < NumOut; i++ {
		// создаём каналы и для каждого из них вызываем горутину Worker
		outs[i] = make(chan int64)
		go Worker(chIn, outs[i])
	}

	// amounts — слайс, в который собирается статистика по горутинам
	amounts := make([]int64, NumOut)
	// chOut — канал, в который будут отправляться числа из горутин `outs[i]`
	chOut := make(chan int64, NumOut)

	var wg sync.WaitGroup

	// 4. Собираем числа из каналов outs
	// ...

	for i := 0; i < NumOut; i++ {
		wg.Add(1) // Увеличиваем счетчик WaitGroup для каждой горутины
		go func(index int, out chan int64) {
			defer wg.Done() // Уменьшаем счетчик при завершении горутины
			for num := range out {
				chOut <- num     // Отправляем число в результирующий канал
				amounts[index]++ // Увеличиваем статистику для соответствующей горутины
			}
		}(i, outs[i])
	}

	go func() {
		// ждём завершения работы всех горутин для outs
		wg.Wait()
		// закрываем результирующий канал
		close(chOut)
	}()

	var count int64 // количество чисел результирующего канала
	var sum int64   // сумма чисел результирующего канала

	// 5. Читаем числа из результирующего канала
	// ...
	for num := range chOut {
		count++    // Увеличиваем количество
		sum += num // Увеличиваем сумму
	}

	fmt.Println("Количество чисел", inputCount, count)
	fmt.Println("Сумма чисел", inputSum, sum)
	fmt.Println("Разбивка по каналам", amounts)

	// проверка результатов
	if inputSum != sum {
		log.Fatalf("Ошибка: суммы чисел не равны: %d != %d\n", inputSum, sum)
	}
	if inputCount != count {
		log.Fatalf("Ошибка: количество чисел не равно: %d != %d\n", inputCount, count)
	}
	for _, v := range amounts {
		inputCount -= v
	}
	if inputCount != 0 {
		log.Fatalf("Ошибка: разделение чисел по каналам неверное\n")
	}

}
