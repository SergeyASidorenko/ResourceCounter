// Package test Пакет с реализацией тестового задания
// Реализован тип потокобезопасного счетчика с интерфейсом использования
package test

import (
	"errors"
	"sync"
)

// Incrementor тип, позволяющий вести подсчет
// возникновений определенного события, ресурсов и.т.д
// Изменение счетчика реализовано через мьютекс, потому что
// атомарная команда (sync/atomic) изменения переменной не гарантирует правильный
// порядок доступа к участку памяти, операция над которым производится атомарно,
// например, если одновременно с началом выполнения атомарной операции изменения участка памяти произоошло чтение этого участка памяти
// из другого потока, нет гарантии что этот поток в итоге считает измененные атомарной операцией данные
type Incrementor struct {
	counter     int          // внутренний счетчик
	maxValue    int          // максимальное значение счетчика, по превышении которого счетчику присваивается нулевое значение
	mtxCounter  sync.RWMutex // мьютекс чтения/записи для блокировки одновременного доступа к счетчику
	mtxMaxValue sync.Mutex   // мьютекс для блокировки одновременного доступа к максимальному значени счетчика
}

// CreateIncrementor функция создает новый объет типа Incrementor и возвращает указатель на него.
// Инициализирует счетчик нулевым значением явно и максимальное значение -
// максимальным значеним для типа Integer
func CreateIncrementor() *Incrementor {
	i := new(Incrementor)
	i.counter = 0
	i.maxValue = int(^uint(0) >> 1)
	return i
}

// Getcounter метод возвращает текущее значение счетчика
// Вызов метода потокобезопасен
func (i *Incrementor) Getcounter() int {
	i.mtxCounter.RLock()
	defer i.mtxCounter.RUnlock()
	counter := i.counter
	return counter
}

// Incrementcounter метод увеличивает значение счетчика
// Вызов метода потокобезопасен
func (i *Incrementor) Incrementcounter() {
	i.mtxCounter.Lock()
	defer i.mtxCounter.Unlock()
	i.counter++
	if i.counter > i.maxValue {
		i.counter = 1
	}
}

// SetMaximumValue метод принимает новое максимальное значения счетчика
// В случае, если новое значение меньше нуля, - возвращает ошибку
// Вызов метода потокобезопасен
func (i *Incrementor) SetMaximumValue(maximumValue int) error {
	i.mtxMaxValue.Lock()
	if maximumValue < 0 {
		return errors.New("недопустимое значение максимального значения")
	}
	i.maxValue = maximumValue
	i.mtxMaxValue.Unlock()
	i.mtxCounter.Lock()
	defer i.mtxCounter.Unlock()
	if i.counter > i.maxValue {
		i.counter = 0
	}
	return nil
}
