# 13-07-25 тестовое задание

## Go 1.23.3
## Запустить сервер
*предварительно установите в конфиге место сохранения архивов*

go run cmd/main.go

## config/config.yaml 
```
port: 8080
storage: "D:\\Documents"
file_types:
  - "application/pdf"
  - "image/jpeg"
timeout: 10s
max_file_size_mb: 10
max_processing_tasks: 3
```

## API
**Добавление задачи на создание ахрива**

Сервер допускает одновременную обработку max_processing_tasks(3), при наличии 3 задач со статусом "processing", вернется ошибка что сервер занят.
```
POST http://localhost:8080/api/tasks
```
**Получение статуса задачи и ссылки на архив по готовности**
```
GET http://localhost:8080/api/tasks/{id}
```

**Добавление ссылки на файл в задачу**

Если в задачу добавляется 3 файл, то начинается ее обработка. Если в этот момент 3 задачи уже имеют статус "processing", эта задача будет ждать завершения какой-то из-за активных задач.
```
POST http://localhost:8080/api/tasks/{id}
```
*Request body example*

```
{
    "href" : "https://image.fonwall.ru/o/ay/nature-black-and-white-white-relax.jpeg?auto=compress&fit=resize&h=375&w=500&display=thumb&domain=img3.fonwall.ru"
}
```

## A few words
* Проект имеет слоистую (layered) архитектуру
* Dependency Injection через интерфейсы
* В repository слое используется sync.RWMutex для потокобезопасного доступа к данным
* В сервисном слое реализован semaphore для ограничения одновременно обрабатывающихся задач
