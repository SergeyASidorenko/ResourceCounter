package main

// Package main Пакет с реализацией тестового задания
// Реализован тип потокобезопасного счетчика с интерфейсом использования

import (
	"sync"
	"testing"
)

// Тестирование функции создания счетчика
func TestCreateIncrementator(t *testing.T) {
	incObj := CreateIncrementator()
	expectedValue := 0
	expectedMaxValue := 1000
	expectedStepValue := 1
	counter := incObj.counter
	if counter != expectedValue {
		t.Fatalf(`функция CreateIncrementator создает объект типа Incrementator с некорректным значением счетчика.\n
		Ожидалось: %d, получено: %d`, expectedValue, counter)
	}
	if incObj.maxValue != expectedMaxValue {
		t.Fatalf(`функция CreateIncrementator создает объект типа Incrementator с некорректным максимальным значением.\n
		Ожидалось: %d, получено: %d`, expectedMaxValue, incObj.maxValue)
	}
	if incObj.step != expectedStepValue {
		t.Fatalf(`функция CreateIncrementator создает объект типа Incrementator с некорректным шагом инкрементации.\n
		Ожидалось: %d, получено: %d`, expectedStepValue, incObj.step)
	}
}

// Тестирование метода установки значения шага счетчика
func TestStepValue(t *testing.T) {
	expectedStepValue := 5
	incObj := CreateIncrementator()
	incObj.SetStep(expectedStepValue)
	if incObj.step != expectedStepValue {
		t.Fatalf(`функция SetStep отработала некорректно.\n
		Ожидалось значения шага счетчика: %d, получено: %d`, expectedStepValue, incObj.step)
	}
}

// Тестирование метода увеличения значения счетчика
func TestIncrementNumber(t *testing.T) {
	incObj := CreateIncrementator()
	expectedCounterValue := 2
	for i := 0; i < expectedCounterValue; i++ {
		incObj.IncrementNumber()
	}
	if incObj.counter != expectedCounterValue {
		t.Fatalf(`функция IncrementNumber после %d вызовов отработала некорректно.\n
		Ожидалось значение счетчика: %d, получено: %d`, expectedCounterValue, expectedCounterValue, incObj.counter)
	}
	incObj.maxValue = expectedCounterValue
	incObj.IncrementNumber()
	if incObj.counter != 1 {
		t.Fatalf(`функция IncrementNumber после перехода счетчика через максимальное значение отработала некорректно.\n
		Ожидалось значение счетчика: %d, получено: %d`, 1, incObj.counter)
	}
}

// Тестирование метода получения значения счетчика
func TestGetNumber(t *testing.T) {
	incObj := CreateIncrementator()
	expectedCounterValue := 5
	// вместо простого присваивания полю counter значения expectedCounterValue
	// использую средства интерфейса данного типа, дабы симитиривать вызовы стороннего кода
	for i := 0; i < expectedCounterValue; i++ {
		incObj.IncrementNumber()
	}
	if incObj.GetNumber() != expectedCounterValue {
		t.Fatalf(`функция GetNumber отработала некорректно.\n
		Ожидалось значение счетчика: %d, получено: %d`, expectedCounterValue, incObj.counter)
	}
	incObj.maxValue = expectedCounterValue
	incObj.IncrementNumber()
	if incObj.GetNumber() != 1 {
		t.Fatalf(`функция GetNumber после перехода счетчика через максимальное значение отработала некорректно.\n
		Ожидалось значение счетчика: %d, получено: %d`, 1, incObj.counter)
	}
}

// Тестирование метода установки максимального значения счетчика
func TestSetMaximumValue(t *testing.T) {
	maxCounterValue := 7
	incObj := CreateIncrementator()
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
		incObj.IncrementNumber()
	}
	err = incObj.SetMaximumValue(maxCounterValue - 1)
	if err != nil {
		t.Fatal(`функция SetMaximumValue отработала некорректно.\n
		После установки максимального значения меньшего текущего значения
		счетчика функция вернула ошибку, ожидался сброс счетчика в нуль`)
	}
	counter := incObj.GetNumber()
	if counter != 0 {
		t.Fatal(`функция SetMaximumValue отработала некорректно.\n
		После установки максимального значения меньшего текущего значения
		счетчика функция вернула ошибку, ожидался сброс счетчика в нуль`)
	}
}

// Тестирование работы счетчика в многопоточном режиме
func TestIncrementInParalell(t *testing.T) {
	incObj := CreateIncrementator()
	goroutineAmount := 10
	var w sync.WaitGroup
	w.Add(goroutineAmount)
	// Запускаем определнное количество горутин, каждая из которых вызывает метод увеличения значения счетчика
	for i := 0; i < goroutineAmount; i++ {
		go func(w *sync.WaitGroup) {
			incObj.IncrementNumber()
			w.Done()
		}(&w)
	}
	w.Wait()
	// ...проверям, каждой ли горутине удалось успешно вызвать метод
	counter := incObj.GetNumber()
	if counter != goroutineAmount {
		t.Fatalf(`функция IncrementNumber отработала некорректно в конкурентом режиме.\n
		Ожидалось значения счетчика: %d, получено: %d`, goroutineAmount, counter)
	}
}

// Тестирование многопоточной установки максимального значения
func TestSetMaximumValueInParalell(t *testing.T) {
	var w sync.WaitGroup
	var mtx = sync.Mutex{}
	ch := make(chan bool)
	incObj := CreateIncrementator()
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
			incObj.IncrementNumber()
			if incObj.GetNumber() == maxCounterValue {
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
	counter := incObj.GetNumber()
	if counter != expectedCounterValue {
		t.Fatalf(`функция SetMaximumValue отработала некорректно в конкурентом режиме.\n
		Ожидалось значения счетчика: %d, получено: %d`, expectedCounterValue, counter)
	}
}
