Получить все команды
```shell
curl -X GET http://localhost:8080/
```

Добавить и запустить команду 
```shell
curl -X POST -H "Content-Type: application/json" -d '{"command":"echo Hello, World!"}' http://localhost:8080/
```

Получить команду по `id`
```shell
curl -X GET http://localhost:8080/10
```

Удалить команду по `id`
```shell
curl -X DELETE http://localhost:8080/1
```

Остановить команду по PID
```shell
curl -X POST http://localhost:8080/stop/35789
```