# Flare Chat API Documentation

## Обзор

Flare Chat API предоставляет полноценную систему чатов с поддержкой личных сообщений и групповых чатов, аналогичную Telegram.

## Аутентификация

Все защищенные эндпоинты требуют JWT токен в заголовке Authorization:
```
Authorization: Bearer <your_jwt_token>
```

## Базовые эндпоинты

### Аутентификация

#### Регистрация пользователя
```http
POST /api/register
Content-Type: application/json

{
  "username": "string",
  "password": "string"
}
```

**Ответ:**
```json
{
  "message": "User registered",
  "user": {
    "id": "string",
    "username": "string"
  }
}
```

#### Вход в систему
```http
POST /api/login
Content-Type: application/json

{
  "username": "string",
  "password": "string"
}
```

**Ответ:**
```json
{
  "token": "jwt_token_string"
}
```

#### Выход из системы
```http
POST /api/logout
Authorization: Bearer <token>
```

#### Профиль пользователя
```http
GET /api/profile
Authorization: Bearer <token>
```

**Ответ:**
```json
{
  "userID": "string",
  "username": "string",
  "message": "Profile retrieved"
}
```

## Чаты

### Получить список чатов пользователя
```http
GET /api/chats
Authorization: Bearer <token>
```

**Ответ:**
```json
{
  "chats": [
    {
      "id": "string",
      "name": "string",
      "type": "private|group",
      "createdBy": "string",
      "createdAt": "2023-01-01T00:00:00Z",
      "updatedAt": "2023-01-01T00:00:00Z",
      "memberCount": 2,
      "avatar": "string",
      "description": "string",
      "lastMessage": {
        "id": "string",
        "chatId": "string",
        "senderId": "string",
        "username": "string",
        "text": "string",
        "type": "text|system|image|file",
        "timestamp": "2023-01-01T00:00:00Z"
      }
    }
  ]
}
```

### Создать новый чат
```http
POST /api/chats
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "string",           // Обязательно для групповых чатов
  "type": "private|group",    // Обязательно
  "members": ["username1"],   // Для private - 1 пользователь, для group - список
  "description": "string"     // Опционально
}
```

**Ответ:**
```json
{
  "id": "string",
  "name": "string",
  "type": "private|group",
  "createdBy": "string",
  "createdAt": "2023-01-01T00:00:00Z",
  "updatedAt": "2023-01-01T00:00:00Z",
  "memberCount": 2,
  "description": "string"
}
```

### Получить информацию о чате
```http
GET /api/chats/{chatId}
Authorization: Bearer <token>
```

**Ответ:**
```json
{
  "chat": {
    "id": "string",
    "name": "string",
    "type": "private|group",
    "createdBy": "string",
    "createdAt": "2023-01-01T00:00:00Z",
    "updatedAt": "2023-01-01T00:00:00Z",
    "memberCount": 2,
    "description": "string"
  },
  "members": [
    {
      "id": "string",
      "chatId": "string",
      "userId": "string",
      "username": "string",
      "role": "admin|member",
      "joinedAt": "2023-01-01T00:00:00Z"
    }
  ]
}
```

### Обновить чат
```http
PUT /api/chats/{chatId}
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "string",        // Опционально
  "description": "string", // Опционально
  "avatar": "string"       // Опционально
}
```

### Удалить чат
```http
DELETE /api/chats/{chatId}
Authorization: Bearer <token>
```

**Примечание:** Только создатель чата может удалить его.

## Сообщения

### Получить сообщения чата
```http
GET /api/chats/{chatId}/messages?limit=50&lastMessageId=string
Authorization: Bearer <token>
```

**Параметры:**
- `limit` (опционально): количество сообщений (по умолчанию 50, максимум 100)
- `lastMessageId` (опционально): ID последнего сообщения для пагинации

**Ответ:**
```json
{
  "messages": [
    {
      "id": "string",
      "chatId": "string",
      "senderId": "string",
      "username": "string",
      "text": "string",
      "type": "text|system|image|file",
      "timestamp": "2023-01-01T00:00:00Z",
      "editedAt": "2023-01-01T00:00:00Z",
      "replyTo": "string"
    }
  ],
  "hasMore": true
}
```

### Отправить сообщение
```http
POST /api/chats/{chatId}/messages
Authorization: Bearer <token>
Content-Type: application/json

{
  "text": "string",      // Обязательно
  "replyTo": "string"   // Опционально - ID сообщения для ответа
}
```

**Ответ:**
```json
{
  "id": "string",
  "chatId": "string",
  "senderId": "string",
  "username": "string",
  "text": "string",
  "type": "text",
  "timestamp": "2023-01-01T00:00:00Z",
  "replyTo": "string"
}
```

## Управление участниками

### Добавить участника в групповой чат
```http
POST /api/chats/{chatId}/members
Authorization: Bearer <token>
Content-Type: application/json

{
  "username": "string"
}
```

**Примечание:** Только администраторы могут добавлять участников.

### Удалить участника из группового чата
```http
DELETE /api/chats/{chatId}/members?userId=string
Authorization: Bearer <token>
```

**Примечание:** Только администраторы могут удалять участников. Создателя чата нельзя удалить.

### Покинуть чат
```http
POST /api/chats/{chatId}/leave
Authorization: Bearer <token>
```

**Примечание:** Создатель группового чата не может покинуть чат.

## WebSocket API

### Подключение
```
ws://localhost:8080/api/ws?token=<jwt_token>
```

### Типы сообщений

#### Присоединиться к чату
```json
{
  "type": "join_chat",
  "data": "chatId"
}
```

#### Покинуть чат
```json
{
  "type": "leave_chat",
  "data": "chatId"
}
```

#### Отправить сообщение
```json
{
  "type": "send_message",
  "data": {
    "chatId": "string",
    "text": "string"
  }
}
```

#### Уведомление о наборе текста
```json
{
  "type": "typing",
  "data": "chatId"
}
```

### Входящие сообщения

#### Новое сообщение
```json
{
  "type": "new_message",
  "chatId": "string",
  "data": {
    "id": "string",
    "chatId": "string",
    "senderId": "string",
    "username": "string",
    "text": "string",
    "type": "text",
    "timestamp": "2023-01-01T00:00:00Z"
  }
}
```

#### Пользователь набирает текст
```json
{
  "type": "user_typing",
  "chatId": "string",
  "data": {
    "userId": "string",
    "username": "string"
  }
}
```

#### Присоединение к чату
```json
{
  "type": "joined_chat",
  "chatId": "string"
}
```

#### Покидание чата
```json
{
  "type": "left_chat",
  "chatId": "string"
}
```

#### Ошибка
```json
{
  "type": "error",
  "error": "string"
}
```

## Коды ошибок

- `400 Bad Request` - Неверные данные запроса
- `401 Unauthorized` - Требуется аутентификация или неверный токен
- `403 Forbidden` - Доступ запрещен (не участник чата, не администратор и т.д.)
- `404 Not Found` - Ресурс не найден
- `405 Method Not Allowed` - Неподдерживаемый HTTP метод
- `500 Internal Server Error` - Внутренняя ошибка сервера

## Примеры использования

### Создание приватного чата
```bash
curl -X POST http://localhost:8080/api/chats \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "private",
    "members": ["username2"]
  }'
```

### Создание группового чата
```bash
curl -X POST http://localhost:8080/api/chats \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Мой групповой чат",
    "type": "group",
    "members": ["user1", "user2", "user3"],
    "description": "Описание группового чата"
  }'
```

### Отправка сообщения
```bash
curl -X POST http://localhost:8080/api/chats/{chatId}/messages \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "text": "Привет всем!"
  }'
```

## Структура базы данных Firestore

### Коллекции:
- `users` - пользователи
- `chats` - чаты
- `chat_members` - участники чатов
- `messages` - сообщения
- `blacklisted_tokens` - заблокированные токены

### Индексы (рекомендуемые):
- `chat_members`: `userId` + `chatId`
- `messages`: `chatId` + `timestamp`
- `chats`: `createdBy`
- `users`: `username`

## Особенности реализации

1. **Приватные чаты** автоматически создаются между двумя пользователями и не могут содержать больше участников.

2. **Групповые чаты** могут содержать неограниченное количество участников с ролями администратора и обычного участника.

3. **Real-time сообщения** поддерживаются через WebSocket соединения.

4. **Пагинация сообщений** реализована для эффективной загрузки истории чата.

5. **Системные сообщения** автоматически создаются при добавлении/удалении участников.

6. **Безопасность**: все операции проверяют права доступа пользователя к чату.
