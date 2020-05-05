## Счетчик ресурсов (событий)

![Счетчик|100x200](https://img.favpng.com/25/0/11/number-counting-apng-icon-png-favpng-uiz0XDNqasHk1Xr1Ajwx1Gx8U.jpg)
----
Простое приложение, предоставляющее работу по RPC протоколу с потокобезопасным счетчиком.
Сериализация абстрактных данных при передачи происходит в формате gob.
Счетчик может быть использован для подсчета ресурсов, возникновения определенных событий и так далее.
Сервис хранит состояние счетчика в постоянном хранилище, поэтому остановка сервера не приводит к удалению этих сведений.