package incrementator

// 2020 Sergey Sidorenko.
// Пакет с реализацией веб-сервиса работы со счетчиком
// Сведения о лицензии отсутствуют

import (
	"errors"
)

// Settings настройки счетчика
type Settings struct {
	Step     *int // шаг инкрементации
	MaxValue *int // максимальное значение счетчика, по превышении которого счетчику присваивается нулевое значение
}

// RPCIncrementator тип, позволяющий вести подсчет
// возникновений определенного события, ресурсов и.т.д
type RPCIncrementator struct {
	counter *Incrementator
}

// CreateRPCIncrementator функция создает новый объет типа RPCIncrementator и возвращает указатель на него.
// Инициализирует счетчик нулевым значением явно и максимальное значение -
// максимальным значеним для типа Integer
func CreateRPCIncrementator() *RPCIncrementator {
	return &RPCIncrementator{CreateIncrementator()}
}

// GetNumber метод возвращает текущее значение счетчика
// Вызов метода потокобезопасен
func (i *RPCIncrementator) GetNumber(req int, resp *int) error {
	number := i.counter.GetNumber()
	*resp = number
	return nil
}

// IncrementNumber метод увеличивает значение счетчика
// Вызов метода потокобезопасен
func (i *RPCIncrementator) IncrementNumber(req int, resp *int) error {
	i.counter.IncrementNumber()
	return nil
}

// SetSettings метод принимает новое максимальное значения счетчика
// В случае, если новое значение меньше нуля, - возвращает ошибку
// Вызов метода потокобезопасен
func (i *RPCIncrementator) SetSettings(req *Settings, resp *int) error {
	// блокируем доступ к полю максимального значения счетчика
	var err error
	if req.MaxValue != nil {
		maxValue := *(req.MaxValue)
		if maxValue < 0 {
			return errors.New("недопустимое значение максимального значения")
		}
		err = i.counter.SetMaximumValue(maxValue)
		if err != nil {
			return err
		}
	}
	if req.Step != nil {
		step := *(req.Step)
		if step < 0 {
			return errors.New("недопустимое значение шага счетчика")
		}
		err = i.counter.SetStep(step)
		if err != nil {
			return err
		}
	}
	return nil
}
