package main

// 2020 Sergey Sidorenko.
// Пакет с реализацией веб-сервиса работы со счетчиком
// Сведения о лицензии отсутствуют

import (
	_ "github.com/mattn/go-sqlite3"
)

// Settings настройки счетчика
type Settings struct {
	Step     *int // шаг инкрементации
	MaxValue *int // максимальное значение счетчика, по превышении которого счетчику присваивается нулевое значение
}

// OnUpdateIncrementor настройки счетчика
type OnUpdateIncrementor func() error

// RPCIncrementator тип, позволяющий вести подсчет
// возникновений определенного события, ресурсов и.т.д
type RPCIncrementator struct {
	IObj     *Incrementator
	OnUpdate OnUpdateIncrementor
}

// CreateRPCIncrementator функция создает новый объет типа RPCIncrementator и возвращает указатель на него.
// Инициализирует счетчик нулевым значением явно и максимальное значение -
// максимальным значеним для типа Integer
func CreateRPCIncrementator() *RPCIncrementator {
	return &RPCIncrementator{CreateIncrementator(), nil}
}

// GetNumber метод возвращает текущее значение счетчика
// Вызов метода потокобезопасен
func (i *RPCIncrementator) GetNumber(req int, resp *int) error {
	*resp = i.IObj.GetNumber()
	return nil
}

// IncrementNumber метод увеличивает значение счетчика
// Вызов метода потокобезопасен
func (i *RPCIncrementator) IncrementNumber(req int, resp *int) (err error) {
	i.IObj.IncrementNumber()
	if i.OnUpdate != nil {
		err = i.OnUpdate()
	}
	return
}

// SetSettings метод принимает новое максимальное значения счетчика
// В случае, если новое значение меньше нуля, - возвращает ошибку
// Вызов метода потокобезопасен
func (i *RPCIncrementator) SetSettings(req *Settings, resp *int) error {
	// блокируем доступ к полю максимального значения счетчика
	var err error
	if req.MaxValue != nil {
		err = i.IObj.SetMaximumValue(*(req.MaxValue))
		if err != nil {
			return err
		}
	}
	if req.Step != nil {
		err = i.IObj.SetStep(*(req.Step))
		if err != nil {
			return err
		}
	}
	if i.OnUpdate != nil {
		err = i.OnUpdate()
	}
	return err
}
