package incrementator

// 2020 Sergey Sidorenko.
// Пакет с реализацией веб-сервиса работы со счетчиком
// Сведения о лицензии отсутствуют

import (
	"errors"
	"sync"
)

// Settings настройки счетчика
type Settings struct {
	Step     *int // шаг инкрементации
	MaxValue *int // максимальное значение счетчика, по превышении которого счетчику присваивается нулевое значение
}

// Incrementor тип, позволяющий вести подсчет
// возникновений определенного события, ресурсов и.т.д
// Изменение счетчика реализовано через мьютекс, потому что
// атомарная команда (sync/atomic) изменения переменной не гарантирует правильный
// порядок доступа к участку памяти, операция над которым производится атомарно,
// например, если одновременно с началом выполнения атомарной операции изменения участка памяти произоошло чтение этого участка памяти
// из другого потока, нет гарантии что этот поток в итоге считает измененные атомарной операцией данные
type Incrementor struct {
	step        int          // шаг инкрементации
	counter     int          // внутренний счетчик
	maxValue    int          // максимальное значение счетчика, по превышении которого счетчику присваивается нулевое значение
	mtxCounter  sync.RWMutex // мьютекс чтения/записи для блокировки одновременного доступа к значению счетчика
	mtxMaxValue sync.RWMutex // мьютекс чтения/записи для блокировки одновременного доступа к максимальному значению счетчика
	mtxStep     sync.RWMutex // мьютекс чтения/записи для блокировки одновременного доступа к значению шага счетчика
}

// CreateIncrementor функция создает новый объет типа Incrementor и возвращает указатель на него.
// Инициализирует счетчик нулевым значением явно и максимальное значение -
// максимальным значеним для типа Integer
func CreateIncrementor() *Incrementor {
	i := new(Incrementor)
	i.counter = 0
	i.maxValue = int(^uint(0) >> 1)
	i.step = 1
	return i
}

// GetNumber метод возвращает текущее значение счетчика
// Вызов метода потокобезопасен
func (i *Incrementor) GetNumber(req int, resp *int) error {
	i.mtxCounter.RLock()
	defer i.mtxCounter.RUnlock()
	*resp = i.counter
	return nil
}

// IncrementNumber метод увеличивает значение счетчика
// Вызов метода потокобезопасен
func (i *Incrementor) IncrementNumber(req int, resp *int) error {
	// блокируем доступ с возможность чтения
	// к полю максимального значения счетчика
	i.mtxMaxValue.RLock()
	maxCounterValue := i.maxValue
	i.mtxMaxValue.RUnlock()
	i.mtxCounter.Lock()
	defer i.mtxCounter.Unlock()
	i.counter++
	if i.counter > maxCounterValue {
		i.counter = 1
	}
	return nil
}

// SetSettings метод принимает новое максимальное значения счетчика
// В случае, если новое значение меньше нуля, - возвращает ошибку
// Вызов метода потокобезопасен
func (i *Incrementor) SetSettings(req *Settings, resp *int) error {
	// блокируем доступ к полю максимального значения счетчика
	if req.MaxValue != nil {
		maxValue := *(req.MaxValue)
		if maxValue < 0 {
			return errors.New("недопустимое значение максимального значения")
		}
		i.mtxMaxValue.Lock()
		i.maxValue = maxValue
		i.mtxMaxValue.Unlock()
		i.mtxCounter.Lock()
		defer i.mtxCounter.Unlock()
		if i.counter > i.maxValue {
			i.counter = 0
		}
	}
	if req.Step != nil {
		step := *(req.Step)
		if step < 0 {
			return errors.New("недопустимое значение шага счетчика")
		}
		i.mtxStep.Lock()
		i.step = step
		i.mtxStep.Unlock()
	}
	return nil
}
