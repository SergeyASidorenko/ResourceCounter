package main

// 2020 Sergey Sidorenko.
// Пакет с реализацией веб-сервиса работы со счетчиком
// Сведения о лицензии отсутствуют

import (
	"errors"
	"sync"
)

// Incrementator тип, позволяющий вести подсчет
// возникновений определенного события, ресурсов и.т.д
// Изменение счетчика реализовано через мьютекс, потому что
// атомарная команда (sync/atomic) изменения переменной не гарантирует правильный
// порядок доступа к участку памяти, операция над которым производится атомарно,
// например, если одновременно с началом выполнения атомарной операции изменения участка памяти произоошло чтение этого участка памяти
// из другого потока, нет гарантии что этот поток в итоге считает измененные атомарной операцией данные
type Incrementator struct {
	step        int          // шаг инкрементации
	counter     int          // внутренний счетчик
	maxValue    int          // максимальное значение счетчика, по превышении которого счетчику присваивается нулевое значение
	mtxCounter  sync.RWMutex // мьютекс чтения/записи для блокировки одновременного доступа к значению счетчика
	mtxMaxValue sync.RWMutex // мьютекс чтения/записи для блокировки одновременного доступа к максимальному значению счетчика
	mtxStep     sync.RWMutex // мьютекс чтения/записи для блокировки одновременного доступа к значению шага счетчика
}

// CreateIncrementator функция создает новый объет типа Incrementator и возвращает указатель на него.
// Инициализирует счетчик нулевым значением явно и максимальное значение -
// максимальным значеним для типа Integer
func CreateIncrementator() *Incrementator {
	i := new(Incrementator)
	i.counter = 0
	i.maxValue = 1000
	i.step = 1
	return i
}

// GetNumber метод возвращает текущее значение счетчика
// Вызов метода потокобезопасен
func (i *Incrementator) GetNumber() int {
	i.mtxCounter.RLock()
	defer i.mtxCounter.RUnlock()
	counter := i.counter
	return counter
}

// IncrementNumber метод увеличивает значение счетчика
// Вызов метода потокобезопасен
func (i *Incrementator) IncrementNumber() {
	// блокируем доступ с возможность чтения
	// к полю максимального значения счетчика
	i.mtxMaxValue.RLock()
	maxCounterValue := i.maxValue
	i.mtxMaxValue.RUnlock()
	i.mtxCounter.Lock()
	defer i.mtxCounter.Unlock()
	i.counter += i.step
	if i.counter > maxCounterValue {
		i.counter = 1
	}
}

// SetMaximumValue метод принимает новое максимальное значения счетчика
// В случае, если новое значение меньше нуля, - возвращает ошибку
// Вызов метода потокобезопасен
func (i *Incrementator) SetMaximumValue(maximumValue int) error {
	// блокируем доступ к полю максимального значения счетчика
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

// SetStep метод принимает новое значения шага приращения счетчика
// В случае, если новое значение меньше нуля, - возвращает ошибку
// Вызов метода потокобезопасен
func (i *Incrementator) SetStep(step int) error {
	// блокируем доступ к полю максимального значения счетчика
	i.mtxStep.Lock()
	defer i.mtxCounter.Lock()
	if step < 0 {
		return errors.New("недопустимое значение шага счетчика")
	}
	i.step = step
	return nil
}
