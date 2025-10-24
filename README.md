# Flare Chat Server

Полноценный сервер чатов с поддержкой личных сообщений и групповых чатов, построенный на Go с использованием Firebase Firestore.

## Возможности

- 🔐 **Аутентификация** - JWT токены с безопасным хешированием паролей
- 💬 **Личные чаты** - Приватные сообщения между двумя пользователями
- 👥 **Групповые чаты** - Чаты с неограниченным количеством участников
- ⚡ **Real-time сообщения** - WebSocket поддержка для мгновенных сообщений
- 📱 **REST API** - Полноценное API для всех операций
- 🔒 **Контроль доступа** - Роли администраторов и участников
- 📄 **Пагинация** - Эффективная загрузка истории сообщений
- 🔔 **Системные уведомления** - Автоматические сообщения о событиях в чате

## Технологии

- **Backend**: Go 1.25+
- **База данных**: Firebase Firestore
- **Аутентификация**: JWT токены
- **Real-time**: WebSocket (gorilla/websocket)
- **Шифрование**: bcrypt для паролей

## Быстрый старт

### Предварительные требования

1. Go 1.25 или выше
2. Firebase проект с настроенным Firestore
3. Service Account Key от Firebase

### Установка

1. Клонируйте репозиторий:
```bash
git clone <repository-url>
cd Flare-Server
```

2. Установите зависимости:
```bash
go mod download
```

3. Настройте Firebase:
   - Создайте проект в [Firebase Console](https://console.firebase.google.com/)
   - Включите Firestore Database
   - Создайте Service Account и скачайте JSON ключ
   - Поместите файл ключа в корень проекта как `serviceAccountKey.json`

4. Настройте переменные окружения:
```bash
# Windows
set PORT=8080
set JWT_SECRET=your-super-secret-jwt-key
set FIREBASE_KEY=serviceAccountKey.json
set COLLECTION=messages

# Linux/Mac
export PORT=8080
export JWT_SECRET=your-super-secret-jwt-key
export FIREBASE_KEY=serviceAccountKey.json
export COLLECTION=messages
```

5. Запустите сервер:
```bash
go run main.go
```

Сервер будет доступен по адресу `http://localhost:8080`

## Структура проекта

```
Flare-Server/
├── internal/
│   ├── config/          # Конфигурация приложения
│   ├── handler/         # HTTP и WebSocket хендлеры
│   ├── middleware/      # Middleware (CORS, аутентификация)
│   ├── models/          # Модели данных
│   ├── repository/      # Слой доступа к данным
│   └── service/         # Бизнес-логика
├── main.go              # Точка входа приложения
├── go.mod               # Go модули
├── serviceAccountKey.json # Firebase ключ (не в git)
├── API_DOCUMENTATION.md # Документация API
└── README.md           # Этот файл
```

## API Endpoints

### Аутентификация
- `POST /api/register` - Регистрация пользователя
- `POST /api/login` - Вход в систему
- `POST /api/logout` - Выход из системы
- `GET /api/profile` - Профиль пользователя

### Чаты
- `GET /api/chats` - Список чатов пользователя
- `POST /api/chats` - Создать новый чат
- `GET /api/chats/{id}` - Информация о чате
- `PUT /api/chats/{id}` - Обновить чат
- `DELETE /api/chats/{id}` - Удалить чат

### Сообщения
- `GET /api/chats/{id}/messages` - Получить сообщения
- `POST /api/chats/{id}/messages` - Отправить сообщение

### Участники
- `POST /api/chats/{id}/members` - Добавить участника
- `DELETE /api/chats/{id}/members` - Удалить участника
- `POST /api/chats/{id}/leave` - Покинуть чат

### WebSocket
- `WS /api/ws` - WebSocket соединение для real-time сообщений

Подробная документация API доступна в [API_DOCUMENTATION.md](API_DOCUMENTATION.md)

## Примеры использования

### Регистрация и вход
```bash
# Регистрация
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "password": "password123"}'

# Вход
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "password": "password123"}'
```

### Создание чата
```bash
# Приватный чат
curl -X POST http://localhost:8080/api/chats \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"type": "private", "members": ["otheruser"]}'

# Групповой чат
curl -X POST http://localhost:8080/api/chats \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name": "My Group", "type": "group", "members": ["user1", "user2"]}'
```

### WebSocket подключение (JavaScript)
```javascript
const token = 'your-jwt-token';
const ws = new WebSocket(`ws://localhost:8080/api/ws?token=${token}`);

ws.onopen = () => {
  // Присоединиться к чату
  ws.send(JSON.stringify({
    type: 'join_chat',
    data: 'chat-id'
  }));
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('Received:', message);
};

// Отправить сообщение
ws.send(JSON.stringify({
  type: 'send_message',
  data: {
    chatId: 'chat-id',
    text: 'Hello, world!'
  }
}));
```

## Конфигурация

Сервер настраивается через переменные окружения:

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `PORT` | Порт сервера | `8080` |
| `JWT_SECRET` | Секретный ключ для JWT | `secret` |
| `FIREBASE_KEY` | Путь к Firebase ключу | `serviceAccountKey.json` |
| `COLLECTION` | Коллекция для старых сообщений | `messages` |

## Безопасность

- Пароли хешируются с использованием bcrypt
- JWT токены с истечением срока действия (24 часа)
- Blacklist для отозванных токенов
- Проверка прав доступа на уровне чатов
- CORS middleware для веб-безопасности

## Разработка

### Запуск в режиме разработки
```bash
go run main.go
```

### Сборка для продакшена
```bash
go build -o flare-server main.go
```

### Тестирование
```bash
go test ./...
```

## Развертывание

### Docker (рекомендуется)
```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/serviceAccountKey.json .
CMD ["./main"]
```

### Переменные окружения для продакшена
```bash
PORT=8080
JWT_SECRET=your-very-secure-secret-key
FIREBASE_KEY=serviceAccountKey.json
COLLECTION=messages
```

## Лицензия

MIT License - см. файл LICENSE для подробностей.
