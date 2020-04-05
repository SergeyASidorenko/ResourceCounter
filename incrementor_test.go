// Package test Пакет с реализацией тестового задания
// Реализован тип потокобезопасного счетчика с интерфейсом использования
package test

import (
	"sync"
	"testing"
)

// Тестирование функции создания счетчика
func TestCreateIncrementor(t *testing.T) {
	incObj := CreateIncrementor()
	maxCounterValue := int(^uint(0) >> 1)
	expectedCounterValue := 0
	counter := incObj.counter
	if counter != 0 {
		t.Fatalf(`функция CreateIncrementor создает объект типа Incrementor с некорректным значением счетчика.\n
		Ожидалось: %d, получено: %d`, expectedCounterValue, counter)
	}
	if incObj.maxValue != maxCounterValue {
		t.Fatalf(`функция CreateIncrementor создает объект типа Incrementor с некорректным максимальным значением.\n
		Ожидалось: %d, получено: %d`, maxCounterValue, incObj.maxValue)
	}
}

// Тестирование метода увеличения значения счетчика
func TestIncrementNumber(t *testing.T) {
	incObj := CreateIncrementor()
	expectedCounterValue := 2
	for i := 0; i < expectedCounterValue; i++ {
		incObj.Incrementcounter()
	}
	if incObj.counter != expectedCounterValue {
		t.Fatalf(`функция Incrementcounter после %d вызовов отработала некорректно.\n
		Ожидалось значение счетчика: %d, получено: %d`, expectedCounterValue, expectedCounterValue, incObj.counter)
	}
	incObj.maxValue = expectedCounterValue
	incObj.Incrementcounter()
	if incObj.counter != 1 {
		t.Fatalf(`функция Incrementcounter после перехода счетчика через максимальное значение отработала некорректно.\n
		Ожидалось значение счетчика: %d, получено: %d`, 1, incObj.counter)
	}
}

// Тестирование метода получения значения счетчика
func TestGetNumber(t *testing.T) {
	incObj := CreateIncrementor()
	expectedCounterValue := 5
	// вместо простого присваивания полю counter значения expectedCounterValue
	// использую средства интерфейса данного типа, дабы симитиривать вызовы стороннего кода
	for i := 0; i < expectedCounterValue; i++ {
		incObj.Incrementcounter()
	}
	if incObj.Getcounter() != expectedCounterValue {
		t.Fatalf(`функция Getcounter отработала некорректно.\n
		Ожидалось значение счетчика: %d, получено: %d`, expectedCounterValue, incObj.counter)
	}
	incObj.maxValue = expectedCounterValue
	incObj.Incrementcounter()
	if incObj.Getcounter() != 1 {
		t.Fatalf(`функция Getcounter после перехода счетчика через максимальное значение отработала некорректно.\n
		Ожидалось значение счетчика: %d, получено: %d`, 1, incObj.counter)
	}
}

// Тестирование метода установки максимального значения счетчика
func TestSetMaximumValue(t *testing.T) {
	maxCounterValue := 7
	incObj := CreateIncrementor()
	err := incObj.SetMaximumValue(-7)
	if err == nil {
		t.Fatal(`функция SetMaximumValue отработала некорректно.\n
		Передано некорректное значение максимальной величины счетчика, однако функция не вернула ошибку`)
	}
	err = incObj.SetMaximumValue(maxCounterValue)
	if err != nil {
		t.Fatal(`функция SetMaximumValue отработала некорректно.\n
		Передано корректное значение максимальной величины счетчика, однако функция вернула ошибку`)
	}
	if incObj.maxValue != maxCounterValue {
		t.Fatalf(`функция SetMaximumValue отработала некорректно.\n
			Ожидалось максимальное значения счетчика: %d, получено: %d`, maxCounterValue, incObj.maxValue)
	}
	// вместо простого присваивания полю counter значения maxCounterValue+1
	// использую средства интерфейса данного типа, дабы симитиривать вызовы стороннего кода
	for i := 0; i < maxCounterValue; i++ {
		incObj.Incrementcounter()
	}
	err = incObj.SetMaximumValue(maxCounterValue - 1)
	if err != nil {
		t.Fatal(`функция SetMaximumValue отработала некорректно.\n
		После установки максимального значения меньшего текущего значения 
		счетчика функция вернула ошибку, ожидался сброс счетчика в нуль`)
	}
	counter := incObj.Getcounter()
	if counter != 0 {
		t.Fatal(`функция SetMaximumValue отработала некорректно.\n
		После установки максимального значения меньшего текущего значения 
		счетчика функция вернула ошибку, ожидался сброс счетчика в нуль`)
	}
}

// Тестирование работы счетчика в многопоточном режиме
func TestIncrementInParalell(t *testing.T) {
	incObj := CreateIncrementor()
	goroutineAmount := 10
	var w sync.WaitGroup
	w.Add(goroutineAmount)
	// Запускаем определнное количество горутин, каждая из которых вызывает метод увеличения значения счетчика
	for i := 0; i < goroutineAmount; i++ {
		go func(w *sync.WaitGroup) {
			incObj.Incrementcounter()
			w.Done()
		}(&w)
	}
	w.Wait()
	// ...проверям, каждой ли горутине удалось успешно вызвать метод
	counter := incObj.Getcounter()
	if counter != goroutineAmount {
		t.Fatalf(`функция Incrementcounter отработала некорректно в конкурентом режиме.\n
		Ожидалось значения счетчика: %d, получено: %d`, goroutineAmount, counter)
	}
}

// Тестирование многопоточной установки максимального значения
func TestSetMaximumValueInParalell(t *testing.T) {
	var w sync.WaitGroup
	var mtx = sync.Mutex{}
	ch := make(chan bool)
	incObj := CreateIncrementor()
	goroutineAmount := 10
	// задаем максимальное число счетчика, ожидая что
	// в многопоточном режиме последующие инкременты начнут отсчет заново
	maxCounterValue := 8
	// А здесь - ожидаемое значение счетчика в итоге
	expectedCounterValue := goroutineAmount - maxCounterValue
	w.Add(goroutineAmount)
	// Запускаем определенное количество горутин, каждая из которых вызывает метод увеличения значения счетчика
	for i := 0; i < goroutineAmount; i++ {
		go func() {
			mtx.Lock()
			incObj.Incrementcounter()
			if incObj.Getcounter() == maxCounterValue {
				ch <- true
				<-ch
			}
			mtx.Unlock()
			w.Done()
		}()
	}
	go func() {
		<-ch
		incObj.SetMaximumValue(maxCounterValue)
		ch <- true
	}()
	w.Wait()
	counter := incObj.Getcounter()
	if counter != expectedCounterValue {
		t.Fatalf(`функция SetMaximumValue отработала некорректно в конкурентом режиме.\n
		Ожидалось значения счетчика: %d, получено: %d`, expectedCounterValue, counter)
	}
}
