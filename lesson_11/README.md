# Клиника: запись к врачу

Пример к занятию «Слоистая архитектура». Только стандартная библиотека Go 1.22+.

## Запуск

Откройте терминал в папке `code/`, где лежит `go.mod`. Для ДЗ начните с [библиотеки](homework/fat/README.md). Команда ниже запускает пример клиники с занятия, а не библиотеку.

```sh
go run ./cmd/clinic
```

Сервис слушает `:8080` (другой адрес - `ADDR=:9000`). SMS уходят на `SMS_URL`, по умолчанию `http://localhost:9090/send`. Если SMS-сервиса нет, в логе будет `sms not sent: ...`, а запись всё равно создаётся: SMS не ушло - запись не отменяем.

Записаться (201):

```sh
curl -i -X POST localhost:8080/appointments \
  -d '{"doctor_id": 7, "patient": "Анна", "phone": "+79990000000", "time": "2030-01-15T10:00:00Z"}'
```

```
HTTP/1.1 201 Created
{"id":1}
```

Повторить тот же запрос (409):

```
HTTP/1.1 409 Conflict
время занято
```

Ещё маршруты: `GET /appointments/1` - запись (нет такой - 404), `GET /ping` - `pong`.

## Тесты

```sh
go test ./...
go vet ./...
```

## Пакеты

```
cmd/clinic/       main: собирает repository → usecase → delivery и запускает http.Server
entity/           Appointment и ошибки ErrInPast, ErrBusy, ErrNotFound - общий словарь слоёв
usecase/          правила записи; объявляет интерфейсы Repository и Notifier
repository/       хранилища: Memory (map под sync.Mutex) и Postgres (database/sql)
notifier/         клиент SMS-сервиса с таймаутом, подходит под usecase.Notifier
delivery/         HTTP: маршруты, JSON-структуры запросов и ответов, ошибка → код, middleware
before/           плохие примеры со слайдов, каждый - свой пакет
  fathandler/     всё в одном обработчике: HTTP, правила и SQL
  dbusecase/      usecase сам подключается к Postgres
  httpcodes/      usecase возвращает коды HTTP
  norows/         delivery знает про sql.ErrNoRows
homework/         домашнее задание: библиотека
  fat/            всё в одном обработчике, шаги - в homework/fat/README.md
```

Зависимости направлены внутрь: `entity` не импортирует ничего из проекта, `usecase` - только `entity`, `delivery`, `repository` и `notifier` - только `entity`. Всех знает только `main`. Проверить:

```sh
go list -f '{{.ImportPath}}: {{.Imports}}' ./...
```

`repository.Postgres` компилируется без драйвера базы. Чтобы запустить его по-настоящему, подключите драйвер (например `github.com/jackc/pgx/v5/stdlib`) и в `main` замените `repository.NewMemory()` на `repository.NewPostgres(db)`.
