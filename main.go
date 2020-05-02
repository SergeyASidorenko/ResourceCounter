package incrementator

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

	_ "github.com/mattn/go-sqlite3"
)

// переменная, хранящая настройки веб-сервиса
var settings *ServiceSettings

// ServiceSettings структура хранения настроек веб-сервиса
type ServiceSettings struct {
	db        string
	tableName string
}

// Init инициализация запуска веб-сервиса (чтение настроек)
// settingsPath - путь к файлу настроек в формате JSON
// Возвращает ошибку, если не удалось завершить работу
func (s *ServiceSettings) Init(settingsPath string) (err error) {
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

// updateIncrementor метод обновления
// записи о состоянии счетчика во внешнем хранилище
// на основе текущего значения счетчика
// Вызов метода потокобезопасен
func updateIncrementor(db *sql.DB, i *Incrementor) (err error) {
	_, err = db.Exec(fmt.Sprintf("UPDATE %s SET value = ?, step = ?, max_value = ?", settings.tableName), i.counter, i.step, i.maxValue)
	return
}

// initIncrementor инициализация состояния счетчика
// Если во внешнем хранилище нет никаких сведений о прежних состояниях -
// вносим запись в хранилище
// Вызов метода потокобезопасен
func initIncrementor(db *sql.DB) (i *Incrementor, err error) {
	i = CreateIncrementor()
	row := db.QueryRow(fmt.Sprintf("SELECT * FROM %s WHERE id = (SELECT MAX(id) AS id FROM %s)", settings.tableName, settings.tableName))
	err = row.Scan(i)
	// Если записей о текущем прогнозе еще нет - добавляем
	if err == sql.ErrNoRows {
		_, err = db.Exec(fmt.Sprintf("INSERT INTO %s VALUES(?,?,?)"), i.counter, i.step, i.maxValue)
	}
	return
}

// connectToDB метод подключения к БД
// Вызов метода потокобезопасен
func connectToDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?cache=shared&mode=memory", dbPath))
	if err != nil {
		return nil, err
	}
	return db, nil
}
func main() {
	db, err := connectToDB("incrementor.db")
	if err != nil {
		log.Fatalln(err)
	}
	settings = new(ServiceSettings)
	// Читаем настройки
	err = settings.Init("config/settings.json")
	if err != nil {
		log.Fatalf("Ошибка инициализации сервера: %s", err.Error())
		return
	}
	// инициазизируем состояние счетчика
	inc, err := initIncrementor(db)
	if err != nil {
		log.Fatalln("не удалось инициализировать сервер")
	}
	err = rpc.Register(inc)
	if err != nil {
		log.Fatalln("не удалось инициализировать сервер")
	}
	rpc.HandleHTTP()
	listener, err := net.Listen("tcp", ":8717")
	if err != nil {
		log.Fatalln("не удалось инициализировать сервер")
	}
	go http.Serve(listener, nil)
}
