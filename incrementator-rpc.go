package main

// 2020 Sergey Sidorenko.
// Пакет с реализацией RPC сервера работы со счетчиком
// Сведения о лицензии отсутствуют

import (
	_ "github.com/mattn/go-sqlite3"
)

// Settings желаемые настройки счетчика, передаваемые клиентами по RPC протоколу
type Settings struct {
	Step     *int // шаг инкрементации
	MaxValue *int // максимальное значение счетчика, по превышении которого счетчику присваивается нулевое значение
}

// OnUpdateIncrementor функция обработчик события изменения состояния счетчика
type OnUpdateIncrementor func() error

// RPCIncrementator объект-обертка, позволяющая вести подсчет
// возникновений определенного события, ресурсов и.т.д
// Используется для регистрации RPC сервера
type RPCIncrementator struct {
	IObj     *Incrementator
	OnUpdate OnUpdateIncrementor
}

// CreateRPCIncrementator функция создает новый объет типа RPCIncrementator и возвращает указатель на него.
func CreateRPCIncrementator() *RPCIncrementator {
	return &RPCIncrementator{CreateIncrementator(), nil}
}

// GetNumber метод возвращает текущее значение счетчика
// req - запрос от клиента
// resp - ответ клиенту
// Вызов метода потокобезопасен
func (i *RPCIncrementator) GetNumber(req int, resp *int) error {
	*resp = i.IObj.GetNumber()
	return nil
}

// IncrementNumber метод увеличивает значение счетчика
// req - запрос от клиента
// resp - ответ клиенту
// Вызов метода потокобезопасен
func (i *RPCIncrementator) IncrementNumber(req int, resp *int) (err error) {
	i.IObj.IncrementNumber()
	if i.OnUpdate != nil {
		err = i.OnUpdate()
	}
	return
}

// SetSettings метод принимает новые настройки счетчика
// В случае, если новые значения настроек меньше нуля, - возвращает ошибку
// req - запрос от клиента
// resp - ответ клиенту
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
