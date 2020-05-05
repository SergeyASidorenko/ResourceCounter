## Счетчик ресурсов (событий)

![Счетчик](https://img.icons8.com/plasticine/2x/counter.png)
----
Простое приложение, предоставляющее работу по RPC протоколу с потокобезопасным счетчиком.
Сериализация абстрактных данных при передачи происходит в формате gob.
Счетчик может быть использован для подсчета ресурсов, возникновения определенных событий и так далее.
Сервис хранит состояние счетчика в постоянном хранилище, поэтому остановка сервера не приводит к удалению этих сведений.