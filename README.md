# Booker 

# Документация по запуску приложения

## 1. Требования
- **Go 1.23+**
- **PostgreSQL 12+**
- Доступ к конфигурации (`config.yaml`, `.env` или переменные окружения)
- **Golang-migrate** (для работы с миграциями)

## 2. Конфигурация

Поддерживаемые источники конфигурации:
- ENV-переменные
- `.env` файл
- `.yaml` файл

### Обязательные параметры:
```yaml
ENV: "local" , "dev" , "prod"  # Окружение
address: string                # Адрес приложения
db_host: string                # Хост БД
db_port: string                # Порт БД
db_name: string                # Имя БД
db_user: string                # Пользователь БД
db_password: string            # Пароль БД
jwt_secret_key: string         # Ключ для подписи JWT токена 
```
### Все параметры:
```dotenv
ENV: local , dev" , "prod"
ADDRESS: хост и порт который будет слушать приложение. (в контейнере лучше прописать :порт)
DB_HOST: Хост БД
DB_PORT: Порт БД
DB_NAME: Имя БД
DB_USER: Пользователь БД
DB_PASSWORD: Пароль БД
DB_MAX_CONNECTIONS: Максимальное количество доступных подключений к бд (пишем просто число. например 10)
DB_MIN_CONNECTIONS: Минимальное количество подключений к бд которое приложение будет поддерживать постоянно (пишем просто число. например 10)
DB_MAX_CONN_LIFETIME: Максимальное время жизни конекшена (принимает формат времени 1h, 1m, 1s)
DB_MAX_CONN_IDLE_TIME: Максимальное время бездействия конекшена (принимает формат времени 1h, 1m, 1s)
DB_HEALTH_CHECK_PERIOD: Периодичность с которой пул будет проверять состояние соединения с БД (принимает формат времени 1h, 1m, 1s)
JWT_SECRET_KEY: Ключ для подписи JWT токена (для работы с телепортом) Нужен ключ котороым телепорт подписывает свои JWT токены
JWT_DURATION: Время жизни JWT (если приложение само будет выдавать JWT токены)
SERVER_TIMEOUT: Таймаут при котором будет завершаться запрос на сервер (принимает формат времени 1h, 1m, 1s)
```
Конфиги при запуске считываются в 3 этапа:

* Считывается файл config.yaml в корне репозитория
* Считывается файл с секретами который можно добавить в корень репозитория (По умолчанию ищет файл "secret_config.yaml")
* Считываются переменные окружения
Каждое последующее считывание, перетирает предыдущее (если есть что перетереть)
## 3. Миграции

### Структура:
```
internal/storage/database/migration/
  |- 000001_<name>_up.sql
  |- 000001_<name>_down.sql
  ...
```

### Команды:
- **Создать новую миграцию**:
  ```bash
  migrate register -ext sql -dir internal/storage/database/migration -seq <название_миграции>
  ```

- **Применить миграции (up)**:
  ```bash
  migrate -path internal/storage/database/migration \
          -database "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable" \
          -verbose up
  ```

- **Откатить миграции (down)**:
  ```bash
  migrate -path internal/storage/database/migration \
          -database "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable" \
          -verbose down
  ```

## 4. Запуск приложения

1. Установите зависимости:
   ```bash
   go mod tidy
   ```

2. Запустите приложение:
   ```bash
   go run .cmd/auth-JWT-microservice/main.go
   ```

## 5. Тестирование

Для запуска тестов выполните:
```bash
go test -v ./...
```