# Тестовое задание

Написать бенчмарк, который ценит RPS запрос postgresql.
Минимальные входные данные (конфиг): оцениваемый sql запрос, dns postgresql, время проведения измерения RPS в миллисекундах.
Нужно использовать конкурентность для максимальной нагрузки на бд. Программа не должна содержать гонки (data race).

# Сборка программы

```bash
go build main.go
```

# Запуск программы

```bash
./main -query="SELECT COUNT(*) FROM test_table;" -dsn="postgres://postgres:root@localhost/testdb?sslmode=disable" -duration=5000 -workers=50
```

# Параметры

| Параметр | Для чего | Пример | 
| ------ | ------ | ------ | 
| -query | Запрос который будем тестировать | SELECT COUNT(*) FROM test_table; |
| -dsn | подключение к базе данных | postgres://postgres:root@localhost/testdb?sslmode=disable |
| -duration | Время проведения измерения RPS в миллисекундах | 5000 (default 1000) |
| -workers | Количество конкурентных воркеров | 50 (default 100) |