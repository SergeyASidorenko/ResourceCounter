package main

// 2020 Sergey Sidorenko.
// Пакет с реализацией веб-сервиса работы со счетчиком
// Сведения о лицензии отсутствуют

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
)

// переменная, хранящая настройки веб-сервиса
var settings *AppSettings

// AppSettings структура хранения настроек веб-сервиса
type AppSettings struct {
	DB        string `json:"db"`
	TableName string `json:"table_name"`
}

// Load загрузка настроек веб-сервиса
// settingsPath - путь к файлу настроек в формате JSON
// Возвращает ошибку, если не удалось завершить работу
func (s *AppSettings) Load(settingsPath string) (err error) {
	fSet, err := os.Open(settingsPath)
	if err != nil {
		return
	}
	err = json.NewDecoder(fSet).Decode(s)
	if err != nil {
		return
	}
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
func initIncrementator(db *sql.DB, s *AppSettings) (i *RPCIncrementator, err error) {
	// Создаем таблицу, где будет храниться состояние счетчика
	_, err = db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s
	(
		id    INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE, 
		value INTEGER,
		step  INTEGER,
		max_value INTEGER
	)`, s.TableName))
	if err != nil {
		return
	}
	i = new(RPCIncrementator)
	IObj := new(Incrementator)
	row := db.QueryRow(fmt.Sprintf("SELECT value, step, max_value FROM %s WHERE id = (SELECT MAX(id) AS id FROM %s)", s.TableName, s.TableName))
	err = row.Scan(&IObj.counter, &IObj.step, &IObj.maxValue)
	// Если записей о текущем прогнозе еще нет - добавляем
	if err == sql.ErrNoRows {
		IObj = CreateIncrementator()
		_, err = db.Exec(fmt.Sprintf("INSERT INTO %s(value,step,max_value) VALUES(?,?,?)", s.TableName), IObj.counter, IObj.step, IObj.maxValue)
	}
	i.IObj = IObj
	// Устанавливаем функцию обратного вызова,
	// которая будет вызываться при каждом изменении состояния счетчика
	// Так как обработчик не принимает параметров,
	// то для использования объекта подключения к БД -
	// используем замыкание
	i.OnUpdate = func() error {
		_, err := db.Exec(fmt.Sprintf("UPDATE %s SET value = ?, step = ?, max_value = ?", s.TableName), i.IObj.counter, i.IObj.step, i.IObj.maxValue)
		return err
	}
	return
}

func main() {
	db, err := connectToDB("incrementator.db")
	if err != nil {
		log.Fatalln(err)
	}
	settings = new(AppSettings)
	// Читаем настройки
	err = settings.Load("config/settings.json")
	if err != nil {
		log.Fatalf("Ошибка инициализации сервера: %s", err.Error())
		return
	}
	// инициализируем состояние счетчика
	inc, err := initIncrementator(db, settings)
	if err != nil {
		log.Fatalln("Ошибка инициализации сервера: ", err)
	}
	err = rpc.Register(inc)
	if err != nil {
		log.Fatalln("Ошибка инициализации сервера: ", err)
	}
	rpc.HandleHTTP()
	listener, err := net.Listen("tcp", ":8716")
	if err != nil {
		log.Fatalln("Ошибка инициализации сервера: ", err)
	}
	http.Serve(listener, nil)
}
