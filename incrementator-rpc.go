package main

// 2020 Sergey Sidorenko.
// Пакет с реализацией веб-сервиса работы со счетчиком
// Сведения о лицензии отсутствуют

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// Settings настройки счетчика
type Settings struct {
	Step     *int // шаг инкрементации
	MaxValue *int // максимальное значение счетчика, по превышении которого счетчику присваивается нулевое значение
}

// RPCIncrementator тип, позволяющий вести подсчет
// возникновений определенного события, ресурсов и.т.д
type RPCIncrementator struct {
	IObj *Incrementator
}

// updateIncrementator метод обновления
// записи о состоянии счетчика во внешнем хранилище
// на основе текущего значения счетчика
// Вызов метода потокобезопасен
func updateIncrementator(db *sql.DB, i *RPCIncrementator) (err error) {
	_, err = db.Exec(fmt.Sprintf("UPDATE %s SET value = ?, step = ?, max_value = ?", settings.TableName), i.IObj.counter, i.IObj.step, i.IObj.maxValue)
	return
}

// connectToDB метод подключения к БД
// Вызов метода потокобезопасен
func connectToDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// initIncrementator инициализация состояния счетчика
// Если во внешнем хранилище нет никаких сведений о прежних состояниях -
// вносим запись в хранилище
// Вызов метода потокобезопасен
func initIncrementator(db *sql.DB) (i *RPCIncrementator, err error) {
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS incrementor
	(
		id    INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE, 
		value INTEGER,
		step  INTEGER,
		max_value INTEGER
	)`)
	if err != nil {
		return
	}
	i = new(RPCIncrementator)
	IObj := new(Incrementator)
	row := db.QueryRow(fmt.Sprintf("SELECT value, step, max_value FROM %s WHERE id = (SELECT MAX(id) AS id FROM %s)", settings.TableName, settings.TableName))
	err = row.Scan(&IObj.counter, &IObj.step, &IObj.maxValue)
	// Если записей о текущем прогнозе еще нет - добавляем
	if err == sql.ErrNoRows {
		IObj = CreateIncrementator()
		_, err = db.Exec(fmt.Sprintf("INSERT INTO %s(value,step,max_value) VALUES(?,?,?)", settings.TableName), IObj.counter, IObj.step, IObj.maxValue)
	}
	i.IObj = IObj
	return
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
	number := i.IObj.GetNumber()
	*resp = number
	return nil
}

// IncrementNumber метод увеличивает значение счетчика
// Вызов метода потокобезопасен
func (i *RPCIncrementator) IncrementNumber(req int, resp *int) error {
	i.IObj.IncrementNumber()
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
		err = i.IObj.SetMaximumValue(maxValue)
		if err != nil {
			return err
		}
	}
	if req.Step != nil {
		step := *(req.Step)
		if step < 0 {
			return errors.New("недопустимое значение шага счетчика")
		}
		err = i.IObj.SetStep(step)
		if err != nil {
			return err
		}
	}
	return nil
}
