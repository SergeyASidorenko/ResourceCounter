package main

// 2020 Sergey Sidorenko.
// Пакет с реализацией веб-сервиса работы со счетчиком
// Сведения о лицензии отсутствуют

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"time"
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
	// инициазизируем состояние счетчика
	inc, err := initIncrementator(db)
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
	go http.Serve(listener, nil)
	time.Sleep(1 * time.Second)
	client()
}
