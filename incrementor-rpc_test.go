package main

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http/httptest"
	"net/rpc"
	"os"
	"strings"
	"sync"
	"testing"
)

// Package main Пакет с реализацией тестового задания
// Реализован тип потокобезопасного счетчика с интерфейсом использования
var (
	newServer                                                 *rpc.Server
	defaultServerAddr, serverAddr, serverWithDBAddr           string
	httpServerAddr                                            string
	defaultServerOnce, serverWithDBOnce, serverOnce, httpOnce sync.Once
	tempDBName                                                string
)

const (
	RPCHTTPPath               = "/testRPC"
	RPCDebugHTTPPath          = "/debugRPC"
	RPCHTTPPathIntegrate      = "/testRPCIntegrate"
	RPCDebugHTTPPathIntegrate = "/debugRPCIntegrate"
)

// Создание тестового сервера
func listenTCP() (net.Listener, string) {
	l, e := net.Listen("tcp", "127.0.0.1:0")
	if e != nil {
		log.Fatalf("ошибка при создании сервера: %v", e)
	}
	return l, l.Addr().String()
}

// Запуск тестового сервера
func createDefaultServer() {
	var l net.Listener
	rpc.Register(CreateRPCIncrementator())
	l, defaultServerAddr = listenTCP()
	go rpc.DefaultServer.Accept(l)
	rpc.HandleHTTP()
	// Запуск HTTP сервера по локальному адресу и случайному порту
	httpOnce.Do(startHTTPServer)
}

// Запуск дополнительного сервера для проверки
// работы с несколькими экземплярами одновременно
func createServer() {
	newServer = rpc.NewServer()
	newServer.Register(CreateRPCIncrementator())
	var l net.Listener
	l, serverAddr = listenTCP()
	go newServer.Accept(l)
	newServer.HandleHTTP(RPCHTTPPath, RPCDebugHTTPPath)
	httpOnce.Do(startHTTPServer)
}

// Запуск дополнительного сервера для проверки
// работы с несколькими экземплярами одновременно
func createIntegratingServer(i *RPCIncrementator) {
	newServer = rpc.NewServer()
	newServer.Register(i)
	var l net.Listener
	l, serverWithDBAddr = listenTCP()
	go newServer.Accept(l)
	newServer.HandleHTTP(RPCHTTPPathIntegrate, RPCDebugHTTPPathIntegrate)
	httpOnce.Do(startHTTPServer)
}

// Запуск HTTP сервера по локальному адресу и случайному порту
func startHTTPServer() {
	server := httptest.NewServer(nil)
	httpServerAddr = server.Listener.Addr().String()
}

func TestRPC(t *testing.T) {
	defaultServerOnce.Do(createDefaultServer)
	testBasicRPCClient(t, defaultServerAddr)
	testRPCClient(t, defaultServerAddr)
	serverOnce.Do(createServer)
	testBasicRPCClient(t, serverAddr)
	testRPCClient(t, serverAddr)
	// Тестирование сервиса с интеграцией с БД
	i, clean := initRPCIntegration(t)
	defer clean()
	createServerWithDB := func() {
		createIntegratingServer(i)
	}
	serverWithDBOnce.Do(createServerWithDB)
	testBasicRPCClient(t, serverWithDBAddr)
	testRPCClient(t, serverWithDBAddr)
}
func testBasicRPCClient(t *testing.T, addr string) {
	client, err := rpc.Dial("tcp", addr)
	if err != nil {
		t.Fatal("ошибка создания клиента для RPC сервера: ", err)
	}
	defer client.Close()
	var reply int
	var s = new(Settings)
	var step int = InitValue
	var maxValue int = InitMaxValue
	s.Step = &step
	s.MaxValue = &maxValue
	// Проверяем метод инкрементации на успешный исход работы
	err = client.Call("RPCIncrementator.IncrementNumber", 0, nil)
	if err != nil {
		t.Fatalf("IncrementNumber: метод возвратил ошибку: %q", err.Error())
	}
	// Проверяем метод получения значения счетчика на успешный исход работы
	err = client.Call("RPCIncrementator.GetNumber", 0, &reply)
	if err != nil {
		t.Fatalf("GetNumber: метод возвратил ошибку: %q", err.Error())
	}
	// Проверяем метод установки новых параметровсчетчика на успешный исход работы
	err = client.Call("RPCIncrementator.SetSettings", s, nil)
	if err != nil {
		t.Fatalf("SetSettings: метод возвратил ошибку: %q", err.Error())
	}
	// Проверяем как отреагируемт сервис на вызов несуществующего метода в нашем RPC сервисе
	err = client.Call("RPCIncrementator.NoMethod", 0, &reply)
	if err == nil {
		t.Error("NoMethod: ожидаемая ошибка")
	} else if !strings.HasPrefix(err.Error(), "rpc: can't find method ") {
		t.Errorf("NoMethod: ожидалось получение ошибки вызова несуществующего метода; получена ошибка: %q", err)
	}
}
func testRPCClient(t *testing.T, addr string) {
	// Создаем клиента RPC сервиса
	client, err := rpc.Dial("tcp", addr)
	if err != nil {
		t.Fatal("ошибка создания клиента для RPC сервера: ", err)
	}
	defer client.Close()
	var reply int
	var step int = 2
	var maxValue int
	var expectedCounterValue int = 5
	var startCounterValue int
	var s = new(Settings)
	s.Step = &step
	// Проверяем изменение счетчика при шаге инкрементации и максимальном значении по умолчанию
	// Сначала получаем текущее значение счетчика
	client.Call("RPCIncrementator.GetNumber", 0, &startCounterValue)
	for i := 0; i < expectedCounterValue; i++ {
		client.Call("RPCIncrementator.IncrementNumber", 0, nil)
	}
	client.Call("RPCIncrementator.GetNumber", 0, &reply)
	// Находим значение счетчика, которое ожидаем получить, исходя из начального его значения
	expectedCounterValue += startCounterValue
	if reply != expectedCounterValue {
		t.Fatalf("Неверное значение счетчика, ожидалось: %d, получено: %d", expectedCounterValue, reply)
	}
	// Проверяем изменение счетчика при изменении шага инкрементации
	startCounterValue = reply
	client.Call("RPCIncrementator.SetSettings", s, nil)
	client.Call("RPCIncrementator.IncrementNumber", 0, nil)
	client.Call("RPCIncrementator.GetNumber", 0, &reply)
	expectedCounterValue = startCounterValue + *s.Step
	if reply != expectedCounterValue {
		t.Fatalf("Неверное значение счетчика после изменения шага инкрементации, ожидалось: %d, получено: %d", expectedCounterValue, reply)
	}
	// Проверяем сброс счетчика в 1 при превышении максимального значения
	maxValue = expectedCounterValue
	expectedCounterValue = 1
	s.MaxValue = &maxValue
	client.Call("RPCIncrementator.SetSettings", s, nil)
	client.Call("RPCIncrementator.IncrementNumber", 0, nil)
	client.Call("RPCIncrementator.GetNumber", 0, &reply)
	if reply != expectedCounterValue {
		t.Fatalf("Неверное значение счетчика после изменения максимального значения, ожидалось: %d, получено: %d", expectedCounterValue, reply)
	}
	step = -3
	err = client.Call("RPCIncrementator.SetSettings", s, nil)
	if err == nil {
		t.Fatal("SetSettings не вернул ошибку, при установке отрицательного шага инкрементации")
	}
	step = 2
	maxValue = -2
	err = client.Call("RPCIncrementator.SetSettings", s, nil)
	if err == nil {
		t.Fatal("SetSettings не вернул ошибку, при установке отрицательного максимального значения")
	}

}
func TestHTTP(t *testing.T) {
	defaultServerOnce.Do(createDefaultServer)
	testHTTPRPC(t, "")
	serverOnce.Do(createServer)
	testHTTPRPC(t, RPCHTTPPath)
	i, clean := initRPCIntegration(t)
	defer clean()
	createServerWithDB := func() {
		createIntegratingServer(i)
	}
	serverWithDBOnce.Do(createServerWithDB)
	testHTTPRPC(t, RPCHTTPPathIntegrate)
}

func testHTTPRPC(t *testing.T, path string) {
	var client *rpc.Client
	var err error
	if path == "" {
		client, err = rpc.DialHTTP("tcp", httpServerAddr)
	} else {
		client, err = rpc.DialHTTPPath("tcp", httpServerAddr, path)
	}
	if err != nil {
		t.Fatal("ошибка создания клиента для RPC сервера", err)
	}
	defer client.Close()
	var reply int
	err = client.Call("RPCIncrementator.IncrementNumber", 0, &reply)
	if err != nil {
		t.Fatalf("IncrementNumber: метод возвратил ошибку: %q", err.Error())
	}
}

// Метод подключения к БД и инициализации RPC объекта с интеграцией с БД
func initRPCIntegration(t *testing.T) (*RPCIncrementator, func()) {
	tempDBName = "temp.db"
	db, err := connectToDB(tempDBName)
	if err != nil {
		t.Fatal(err)
	}
	if db == nil {
		t.Fatal("метод connectToDB вернул нулевой указатель соединения с БД")
	}
	settings = new(AppSettings)
	// Читаем настройки
	tempConfigFile := "test_config.json"
	file, err := os.OpenFile(tempConfigFile, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatalf("Ошибка при создании временного файла настроек: %q", err.Error())
	}
	settings.DB = tempDBName
	settings.TableName = "incrementator"
	data, err := json.Marshal(settings)
	if err != nil {
		log.Fatalf("Ошибка сериализации временного объекта настроек приложения: %q", err.Error())
	}
	_, err = io.WriteString(file, string(data))
	if err != nil {
		log.Fatalf("Ошибка записи тестовых настроек приложения в тестовый файл: %q", err.Error())
	}
	err = settings.Load(tempConfigFile)
	if err != nil {
		t.Fatalf("Метод загрузки настроек приложения Load вернул ошибку: %q", err.Error())
	}
	if settings.DB == "" {
		t.Fatal("Метод загрузки настроек приложения Load неверно считал имя базы данных")
	}
	if settings.TableName == "" {
		t.Fatal("Метод загрузки настроек приложения Load  неверно считал имя таблицы")
	}
	i, err := initIncrementator(db, settings)
	if err != nil {
		t.Fatalf("метод initIncrementator вернул ошибку: %q", err.Error())
	}
	if i == nil {
		t.Fatal("метод initIncrementator вернул нулевой указатель объекта RPCIncrementor")
	}
	if i.IObj == nil {
		t.Fatal("метод initIncrementator вернул нулевой указатель объекта Incrementor")
	}
	if i.OnUpdate == nil {
		t.Fatal("метод initIncrementator вернул нулевой указатель на обработчик события обновления счетчика")
	}
	return i, clean(tempConfigFile, tempDBName)
}

// Очитска файловой системы и освобождение ресурсов после тестирования
func clean(files ...string) func() {
	return func() {
		for _, file := range files {
			err := os.Remove(file)
			if err != nil {
				log.Printf("ошибка при очистке: %q", err.Error())
			}
		}
	}
}
