# stan-service

## Instructions for installing and running the application:

### 1 Method:

1) Up docker, thereby creating a local database(postgres), redis and nats-streaming:

```
docker-compose up
```

2) Launch the application using the command:

```
go run go-service/cmd/main.go
```

3) Send messages to the broker using the following command:
```
go run pusher/main.go
```

3) Get data using id by handler:

https://localhost:8000/data/{id}