# HW
**Домашка:**
- Добавить observability в gRPC cервис
- Логировать все запросы/ответы
- Метрики по длительности/количеству запросов (*)
- Разметить спанами (*)

**В работе использованы:**
- Логгер zap
- Prometheus
- GrayLog
- OpenTracing

Клиент grpc ```./client/client.go```

Запуск сервера ```make run-log```

http://localhost:9090/-/reload