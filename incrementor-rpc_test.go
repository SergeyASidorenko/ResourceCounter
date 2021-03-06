package main

// 2020 Sergey Sidorenko.
// Пакет с реализацией RPC сервера работы со счетчиком
// Сведения о лицензии отсутствуют
import (
	"database/sql"
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

var (
	// RPC сервер
	newServer *rpc.Server
	// адреса развертываемых тестовых RPC серверов
	defaultServerAddr, serverAddr, serverWithDBAddr, httpServerAddr string
	// объекта для выполнения единоразовой инициализации развертываемых тестовых RPC серверов
	defaultServerOnce, serverWithDBOnce, serverOnce, httpOnce sync.Once
	i                                                         *RPCIncrementator
	// указатель на объект подключения к БД в интеграционных тестах
	db *sql.DB
)

const (
	// путь HTTP обработчика тестового RPC сервера
	RPCHTTPPath = "/testRPC"
	// путь отладочного HTTP обработчика тестового RPC сервера
	RPCDebugHTTPPath = "/debugRPC"
	// путь HTTP обработчика тестового RPC сервера с интеграцией с БД
	RPCHTTPPathIntegrate = "/testRPCIntegrate"
	// путь отладочного HTTP обработчика тестового RPC сервера с интеграцией с БД
	RPCDebugHTTPPathIntegrate = "/debugRPCIntegrate"
	// имя временной (для тестирования) БД
	tempDBName string = "test.db"
	// имя таблицы хранения состояния счетчика во временной (для тестирования) БД
	tableName string = "incrementator"
)

// Создание тестового сервера на случайном порту
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

// Запуск дополнительного сервера с интеграцией с БД для проверки
// работы с несколькими экземплярами одновременно
func createIntegratingServer() {
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

// Сводный метод тестирования RPC сервера
func TestRPC(t *testing.T) {
	defer clean(tempDBName)
	defaultServerOnce.Do(createDefaultServer)
	testBasicRPCClient(t, defaultServerAddr)
	testRPCClient(t, defaultServerAddr)
	serverOnce.Do(createServer)
	testBasicRPCClient(t, serverAddr)
	testRPCClient(t, serverAddr)
	// Тестирование сервиса с интеграцией с БД
	initRPCWithDBIntegration(t)
	serverWithDBOnce.Do(createIntegratingServer)
	testBasicRPCClient(t, serverWithDBAddr)
	testRPCClient(t, serverWithDBAddr)
}

// Тестирование общего обслуживания RPC запросов
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

// Расширенное тестирование обслуживания RPC запросов
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
	// Проверяем обработку запроса на установку шага
	// инкрементации с неверным его значеним (отрицательным)
	step = -3
	err = client.Call("RPCIncrementator.SetSettings", s, nil)
	if err == nil {
		t.Fatal("SetSettings не вернул ошибку, при установке отрицательного шага инкрементации")
	}
	// Проверяем обработку запроса на установку максимального значения счетчика
	// с неверным его значеним (отрицательным)
	step = 2
	maxValue = -2
	err = client.Call("RPCIncrementator.SetSettings", s, nil)
	if err == nil {
		t.Fatal("SetSettings не вернул ошибку, при установке отрицательного максимального значения")
	}
}

// Тестирование HTTP обработчиков
func TestHTTP(t *testing.T) {
	defaultServerOnce.Do(createDefaultServer)
	testHTTPRPC(t, "")
	serverOnce.Do(createServer)
	testHTTPRPC(t, RPCHTTPPath)
	initRPCWithDBIntegration(t)
	serverWithDBOnce.Do(createIntegratingServer)
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
func initRPCWithDBIntegration(t *testing.T) {
	db, err := connectToDB(tempDBName)
	if err != nil {
		t.Fatal(err)
	}
	if db == nil {
		t.Fatal("метод connectToDB вернул нулевой указатель соединения с БД")
	}
	i, err = initIncrementator(db, tableName)
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
}

// Тестирование создания/загрузки файла логирования приложения
func TestInitLog(t *testing.T) {
	logFilePath := "test_errors.log"
	defer clean(logFilePath)
	// Тестирование создания файла логирования приложения
	err := initLog(logFilePath)
	if err != nil {
		t.Fatalf("Функция создания/загрузки файла логирования вернула ошибку: %q", err.Error())
	}
	// Тестирование загрузки файла логирования приложения
	err = initLog(logFilePath)
	if err != nil {
		t.Fatalf("Функции создания/загрузки файла логирования не удалось использовать уже ранее созданный файл: %q", err.Error())
	}
}

// Тестирование загрузки настроек приложения
func TestLoadSettings(t *testing.T) {
	tempConfigFile := "test_config.json"
	defer clean(tempConfigFile)
	settings := new(AppSettings)
	// Читаем настройки
	file, err := os.OpenFile(tempConfigFile, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatalf("Ошибка при создании временного файла настроек: %q", err.Error())
	}
	settings.DB = tempDBName
	settings.TableName = tableName
	settings.LogFilePath = tempConfigFile
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
}

// Очитска файловой системы и освобождение ресурсов после тестирования
func clean(files ...string) {
	for _, file := range files {
		err := os.Remove(file)
		if err != nil {
			log.Printf("ошибка при очистке: %q", err.Error())
		}
	}
}
