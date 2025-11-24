# Backend API Specification
## Спецификация API для системы школьного расписания

**Базовый URL**: `/api`

**Авторизация**: JWT токен в **httpOnly cookie**: `access-token=<jwt_token>`

**Важно**: 
- Токены через cookies (не Authorization header)
- Access токен: 10 минут
- Refresh токен: 7 дней
- Бэкенд читает: `req.cookies['access-token']`

---

## 📋 Таблица всех endpoints

| Endpoint | Метод | Описание | Auth |
|----------|-------|----------|------|
| `/auth/login` | POST | Вход в систему | ❌ |
| `/auth/refresh` | POST | Обновление токена | ✅ Refresh |
| `/schedule` | GET | Получить расписание | ✅ |
| `/schedule` | PUT | Сохранить расписание | ✅ |
| `/schedule/generate` | POST | Сгенерировать расписание | ✅ |
| `/schedule/:id` | GET | Расписание по ID | ✅ |
| `/schedule` | POST | Создать именованное расписание | ✅ |
| `/schedule/:id` | DELETE | Удалить расписание | ✅ |
| `/classes` | GET | Получить классы | ✅ |
| `/classes` | POST | Создать класс | ✅ |
| `/classes/:id` | DELETE | Удалить класс | ✅ |
| `/classes/bulk` | PUT | Массовое обновление классов | ✅ |
| `/users/Teachers` | GET | Получить учителей (полные) | ✅ |
| `/users/LightTeachers` | GET | Получить учителей (ФИО) | ✅ |
| `/users/Teachers` | POST | Создать учителя | ✅ |
| `/users/Teachers/:id` | DELETE | Удалить учителя | ✅ |
| `/users/Teachers/bulk` | PATCH | Массовое обновление учителей | ✅ |
| `/subjects` | GET | Получить предметы | ✅ |
| `/subjects` | POST | Создать предмет | ✅ |
| `/subjects/:id` | DELETE | Удалить предмет | ✅ |
| `/classrooms` | GET | Получить кабинеты | ✅ |
| `/classrooms` | POST | Создать кабинет | ✅ |
| `/classrooms/:id` | DELETE | Удалить кабинет | ✅ |

---

## 📚 Оглавление
1. [Авторизация](#авторизация)
2. [Расписание](#расписание)
3. [Классы](#классы)
4. [Учителя](#учителя)
5. [Предметы](#предметы)
6. [Кабинеты](#кабинеты)
7. [Типы данных](#типы-данных)
8. [Примечания для бэкенда](#примечания-для-бэкенда)

---

## Авторизация

### POST /auth/login

| Параметр | Значение |
|----------|----------|
| **Endpoint** | `/auth/login` |
| **Метод** | POST |
| **Auth** | Не требуется |

**Что отправляем (Request Body)**:
```typescript
{
  email: string,      // "teacher@school.com"
  password: string    // "password123"
}
```

**Что получаем (Response 200)**:
```typescript
{
  accessToken: string,   // JWT токен (10 минут)
  refreshToken: string,  // JWT токен (7 дней)
  user: {
    id: string,          // "teacher-1"
    email: string,       // "teacher@school.com"
    name: string         // "Анна Иванова"
  }
}
```

**JWT Payload (accessToken)**:
```typescript
{
  sub: string,      // userId: "teacher-1"
  email: string,    // "teacher@school.com"
  role: string,     // "teacher" | "admin"
  iat: number,      // timestamp создания
  exp: number       // timestamp истечения
}
```

**Ошибки**:
- `401` - Неверный email или пароль

---

### POST /auth/refresh

| Параметр | Значение |
|----------|----------|
| **Endpoint** | `/auth/refresh` |
| **Метод** | POST |
| **Auth** | Refresh токен в header |

**Что отправляем (Headers)**:
```
Authorization: Bearer <refresh_token>
```

**Что получаем (Response 200)**:
```typescript
{
  accessToken: string,    // Новый JWT токен (10 минут)
  refreshToken?: string   // Опционально новый refresh токен
}
```

**Ошибки**:
- `401` - Невалидный refresh токен

---

## Расписание

### GET /schedule

| Параметр | Значение |
|----------|----------|
| **Endpoint** | `/schedule` |
| **Метод** | GET |
| **Auth** | Access токен (cookie) |
| **Фильтрация** | По userId из JWT |

**Что отправляем**: Ничего (токен в cookie)

**Что получаем (Response 200)**:
```typescript
{
  data: [
    {
      dayOfWeek: "monday" | "tuesday" | "wednesday" | "thursday" | "friday" | "saturday" | "sunday",
      lessonNumber: number,  // 1-8
      lessons: [
      {
        id?: string,
        subject: {
          id: string,
          name: string
        },
        teachers: [
          {
            id: string,
            firstName: string,
            lastName: string,
            patronymic: string | null
          }
        ],
        rooms: [
          {
            id: string,
            name: string
          }
        ],
        participants: [
          {
            class: {
              id: string,
              name: string,
              classTeacher: {...},
              subjects: [...],
              groups: [...]
            },
            groupIds?: string[]  // Если не указано - весь класс
          }
        ]
      }
    ]
  }
]
```

**Логика фильтрации**:
1. Извлечь `userId` из JWT токена
2. Найти `teacher_id` по `userId`
3. Вернуть только уроки из таблицы `lesson_teachers` где `teacher_id = ...`

---

### PUT /schedule

| Параметр | Значение |
|----------|----------|
| **Endpoint** | `/schedule` |
| **Метод** | PUT |
| **Auth** | Access токен (cookie) |
| **Роль** | Admin/Scheduler |

**Что отправляем (Request Body)**:
```typescript
{
  data: [
    {
      dayOfWeek: "monday" | "tuesday" | ...,
      lessonNumber: number,
      lessons: [
        {
          subject: { id: string, name: string },
          teachers: [{ id: string, firstName: string, lastName: string, patronymic: string | null }],
          rooms: [{ id: string, name: string }],
          participants: [
            {
              class: { id: string, name: string },
              groupIds?: string[]
            }
          ]
        }
      ]
    }
  ]
}
```

**Что получаем (Response 200)**:
```typescript
{
  message: string  // "Расписание успешно сохранено"
}
```

**Ошибки (Response 400)**:
```typescript
{
  error: string,  // "Конфликт расписания"
  details: [
    {
      type: "teacher_conflict" | "classroom_conflict" | "class_conflict",
      message: string,
      dayOfWeek: string,
      lessonNumber: number
    }
  ]
}
```

**Валидация**:
- ❌ Учитель не может вести два урока одновременно
- ❌ Кабинет не может быть занят дважды
- ❌ Класс/группа не может иметь два урока в одно время

---

### POST /schedule/generate

| Параметр | Значение |
|----------|----------|
| **Endpoint** | `/schedule/generate` |
| **Метод** | POST |
| **Auth** | Access токен (cookie) |
| **Роль** | Admin |

**Что отправляем (Request Body, опционально)**:
```typescript
{
  algorithm?: "greedy",
  maxLessonsPerDay?: number,
  priorities?: {
    balanceWorkload?: boolean,
    minimizeGaps?: boolean
  }
}
```

**Что получаем (Response 200)**:
```typescript
{
  data: [
    {
      dayOfWeek: string,
      lessonNumber: number,
      lessons: [...]
    }
  ]
}
```

**Ошибки (Response 400)**:
```typescript
{
  error: string,  // "Невозможно сгенерировать расписание"
  reason: string  // "Недостаточно учителей для покрытия всех уроков"
}
```

---

### GET /schedule/:id

| Параметр | Значение |
|----------|----------|
| **Endpoint** | `/schedule/:id` |
| **Метод** | GET |
| **Auth** | Access токен (cookie) |

**Что отправляем**: `id` в URL

**Что получаем (Response 200)**:
```typescript
{
  data: {
    id: string,
    name: string,  // "Расписание на 1 семестр 2024"
    scheduleSlots: [
      // Тот же формат что GET /schedule
    ]
  }
}
```

---

### POST /schedule

| Параметр | Значение |
|----------|----------|
| **Endpoint** | `/schedule` |
| **Метод** | POST |
| **Auth** | Access токен (cookie) |
| **Роль** | Admin |

**Что отправляем (Request Body)**:
```typescript
{
  name: string,  // "Расписание на 2 семестр 2024"
  scheduleSlots: [
    // Тот же формат что PUT /schedule
  ]
}
```

**Что получаем (Response 201)**:
```typescript
{
  data: {
    id: string,
    name: string,
    scheduleSlots: [...]
  }
}
```

---

### DELETE /schedule/:id

| Параметр | Значение |
|----------|----------|
| **Endpoint** | `/schedule/:id` |
| **Метод** | DELETE |
| **Auth** | Access токен (cookie) |
| **Роль** | Admin |

**Что отправляем**: `id` в URL

**Что получаем (Response 204)**: Пустой ответ

---

## Классы

### GET /classes

| Параметр | Значение |
|----------|----------|
| **Endpoint** | `/classes` |
| **Метод** | GET |
| **Auth** | Access токен (cookie) |

**Что отправляем**: Ничего

**Что получаем (Response 200)**:
```typescript
{
  data: [
    {
      id: string,
      name: string,  // "5А"
      classTeacher: {
        id: string,
        firstName: string,
        lastName: string,
        patronymic: string | null
      } | null,
      subjects: [
        {
          subject: {
            id: string,
            name: string
          },
          hoursPerWeek: number,
          split?: {
            groupsCount: number,
            crossClassAllowed?: boolean
          }
        }
      ],
      groups?: [
        {
          id: string,
          name: string,  // "Группа 1"
          size?: number
        }
      ]
    }
  ]
}
```

---

### POST /classes

| Параметр | Значение |
|----------|----------|
| **Endpoint** | `/classes` |
| **Метод** | POST |
| **Auth** | Access токен (cookie) |

**Что отправляем (Request Body)**:
```typescript
{
  name: string  // "5А"
}
```

**Что получаем (Response 201)**:
```typescript
{
  data: {
    id: string,
    name: string,
    classTeacher: null,
    subjects: [],
    groups: []
  }
}
```

---

### DELETE /classes/:id

| Параметр | Значение |
|----------|----------|
| **Endpoint** | `/classes/:id` |
| **Метод** | DELETE |
| **Auth** | Access токен (cookie) |

**Что отправляем**: `id` в URL

**Что получаем (Response 204)**: Пустой ответ

---

### PUT /classes/bulk

| Параметр | Значение |
|----------|----------|
| **Endpoint** | `/classes/bulk` |
| **Метод** | PUT |
| **Auth** | Access токен (cookie) |

**Что отправляем (Request Body)**:
```typescript
{
  data: [
    {
      id: string,
      name: string,
      classTeacher: { id: string },
      subjects: [
        {
          subject: { id: string, name: string },
          hoursPerWeek: number,
          split?: {
            groupsCount: number,
            crossClassAllowed?: boolean
          }
        }
      ],
      groups: [
        {
          id: string,
          name: string,
          size?: number
        }
      ]
    }
  ]
}
```

**Что получаем (Response 200)**:
```typescript
{
  message: string,  // "Классы успешно обновлены"
  updated: number
}
```

---

## Учителя

### GET /users/Teachers

| Параметр | Значение |
|----------|----------|
| **Endpoint** | `/users/Teachers` |
| **Метод** | GET |
| **Auth** | Access токен (cookie) |

**Что отправляем**: Ничего

**Что получаем (Response 200)**:
```typescript
{
  data: [
    {
      id: string,
      firstName: string,
      lastName: string,
      patronymic: string | null,
      subjects: [
      {
        subject: { id: string, name: string },
        hoursPerWeek: number | null
      }
    ],
    classRoom: {
      id: string,
      name: string
    },
    class: {
      id: string,
      name: string
    },
        workloadHoursPerWeek: number,
        classHours: [
          {
            class: { id: string, name: string },
            subject: { id: string, name: string },
            groupId?: string,
            hours: number
          }
        ]
      }
    ]
  }
}
```

---

### GET /users/LightTeachers

| Параметр | Значение |
|----------|----------|
| **Endpoint** | `/users/LightTeachers` |
| **Метод** | GET |
| **Auth** | Access токен (cookie) |

**Что отправляем**: Ничего

**Что получаем (Response 200)**:
```typescript
{
  data: [
    {
      id: string,
      firstName: string,
      lastName: string,
      patronymic: string | null
    }
  ]
}
```

---

### POST /users/Teachers

| Параметр | Значение |
|----------|----------|
| **Endpoint** | `/users/Teachers` |
| **Метод** | POST |
| **Auth** | Access токен (cookie) |

**Что отправляем (Request Body)**:
```typescript
{
  firstName: string,
  lastName: string,
  patronymic: string | null
}
```

**Что получаем (Response 201)**:
```typescript
{
  data: {
    id: string,
    firstName: string,
    lastName: string,
    patronymic: string | null,
    subjects: [],
    classRoom: null,
    class: null,
    workloadHoursPerWeek: 0,
    classHours: []
  }
}
```

---

### DELETE /users/Teachers/:id

| Параметр | Значение |
|----------|----------|
| **Endpoint** | `/users/Teachers/:id` |
| **Метод** | DELETE |
| **Auth** | Access токен (cookie) |

**Что отправляем**: `id` в URL

**Что получаем (Response 204)**: Пустой ответ

---

### PATCH /users/Teachers/bulk

| Параметр | Значение |
|----------|----------|
| **Endpoint** | `/users/Teachers/bulk` |
| **Метод** | PATCH |
| **Auth** | Access токен (cookie) |

**Что отправляем (Request Body)**:
```typescript
{
  data: [
    {
      id: string,
      firstName: string,
      lastName: string,
      patronymic: string | null,
      subjects: [...],
      classRoom: {...},
      class: {...},
      workloadHoursPerWeek: number,
      classHours: [...]
    }
  ]
}
```

**Что получаем (Response 200)**:
```typescript
{
  message: string,  // "Учителя успешно обновлены"
  updated: number
}
```

---

## Предметы

### GET /subjects

| Параметр | Значение |
|----------|----------|
| **Endpoint** | `/subjects` |
| **Метод** | GET |
| **Auth** | Access токен (cookie) |

**Что отправляем**: Ничего

**Что получаем (Response 200)**:
```typescript
{
  data: [
    {
      id: string,
      name: string  // "Математика"
    }
  ]
}
```

---

### POST /subjects

| Параметр | Значение |
|----------|----------|
| **Endpoint** | `/subjects` |
| **Метод** | POST |
| **Auth** | Access токен (cookie) |

**Что отправляем (Request Body)**:
```typescript
{
  name: string  // "Математика"
}
```

**Что получаем (Response 201)**:
```typescript
{
  id: string,
  name: string
}
```

---

### DELETE /subjects/:id

| Параметр | Значение |
|----------|----------|
| **Endpoint** | `/subjects/:id` |
| **Метод** | DELETE |
| **Auth** | Access токен (cookie) |

**Что отправляем**: `id` в URL

**Что получаем (Response 204)**: Пустой ответ

---

## Кабинеты

### GET /classrooms

| Параметр | Значение |
|----------|----------|
| **Endpoint** | `/classrooms` |
| **Метод** | GET |
| **Auth** | Access токен (cookie) |

**Что отправляем**: Ничего

**Что получаем (Response 200)**:
```typescript
{
  data: [
    {
      id: string,
      name: string  // "101"
    }
  ]
}
```

---

### POST /classrooms

| Параметр | Значение |
|----------|----------|
| **Endpoint** | `/classrooms` |
| **Метод** | POST |
| **Auth** | Access токен (cookie) |

**Что отправляем (Request Body)**:
```typescript
{
  name: string  // "101"
}
```

**Что получаем (Response 201)**:
```typescript
{
  id: string,
  name: string
}
```

---

### DELETE /classrooms/:id

| Параметр | Значение |
|----------|----------|
| **Endpoint** | `/classrooms/:id` |
| **Метод** | DELETE |
| **Auth** | Access токен (cookie) |

**Что отправляем**: `id` в URL

**Что получаем (Response 204)**: Пустой ответ
{
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "teacher-1",
    "email": "teacher@school.com",
    "name": "Анна Иванова"
  }
}
```

**Что должен вернуть бэкенд**:
1. `accessToken` - JWT токен со сроком жизни 10 минут
2. `refreshToken` - JWT токен со сроком жизни 7 дней
3. `user` - информация о пользователе

**JWT Payload для access токена**:
```json
{
  "sub": "teacher-1",
  "email": "teacher@school.com",
  "role": "teacher",
  "iat": 1700000000,
  "exp": 1700000600
}
```

**JWT Payload для refresh токена**:
```json
{
  "sub": "teacher-1",
  "type": "refresh",
  "iat": 1700000000,
  "exp": 1700604800
}
```

**Response** `401 Unauthorized`:
```json
{
  "error": "Invalid credentials"
}
```

---

### `POST /auth/refresh`
**Описание**: Обновление access токена

**Headers**:
```
Authorization: Bearer <refresh_token>
```

**Response** `200 OK`:
```json
{
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response** `401 Unauthorized`:
```json
{
  "error": "Invalid refresh token"
}
```

---

### Как работает авторизация на фронте

#### 1. **Login Flow**
```
User → LoginForm → POST /api/auth/login (Next.js API Route)
                    ↓
                    POST /auth/login (Backend)
                    ↓
                    Получить токены
                    ↓
                    Сохранить в httpOnly cookies
                    ↓
                    Redirect → /panel/schedule
```

#### 2. **Token Refresh Flow**
```
User → Protected Route → Middleware проверяет access токен
                          ↓
                          Токен истек?
                          ↓ Да
                          POST /api/auth/refresh (Next.js)
                          ↓
                          POST /auth/refresh (Backend)
                          ↓
                          Новый access токен → cookie
                          ↓
                          Пропустить пользователя
```

#### 3. **API Request Flow**
```
Component → React Query → Service.proxyFetch()
                          ↓
                          fetch(url, { credentials: 'include' })
                          ↓
                          Браузер автоматически добавляет cookie
                          ↓
                          Backend получает access-token из cookies
```

#### 4. **Middleware Logic**
```typescript
// src/middleware.ts
export async function middleware(request: NextRequest) {
  // 1. Проверить наличие refresh токена
  const refreshToken = request.cookies.get('refresh-token')?.value;
  if (!refreshToken) {
    return redirect('/login');
  }

  // 2. Проверить access токен
  const accessToken = request.cookies.get('access-token')?.value;
  if (!accessToken || isTokenExpired(accessToken)) {
    // 3. Обновить через /api/auth/refresh
    const refreshResponse = await fetch('/api/auth/refresh', {
      method: 'POST',
      headers: { Cookie: request.headers.get('cookie') || '' }
    });
    
    if (!refreshResponse.ok) {
      return redirect('/login');
    }
  }

  return NextResponse.next();
}
```

---

### Безопасность токенов

#### HttpOnly Cookies
```typescript
// Фронтенд устанавливает cookies после получения от бэкенда
response.cookies.set('access-token', data.accessToken, {
  httpOnly: true,        // ❌ Недоступен через JavaScript
  secure: true,          // ✅ Только HTTPS (на production)
  sameSite: 'lax',       // ✅ Защита от CSRF
  maxAge: 10 * 60,       // ⏰ 10 минут
  path: '/',             // 🌐 Доступен для всех путей
});
```

#### Почему httpOnly?
- ❌ **XSS защита**: JavaScript не может прочитать токен
- ✅ **Автоматическая отправка**: Браузер сам добавляет cookie к каждому запросу
- ✅ **Безопасное хранение**: Токен не попадает в localStorage/sessionStorage

---

## Расписание (Schedule)

### `GET /schedule`
**Описание**: Получить расписание для текущего пользователя (фильтруется по JWT токену)

**Headers**:
```
Cookie: access-token=<jwt_token>
```

**Как бэкенд читает токен**:
```javascript
// Express.js пример
app.get('/schedule', (req, res) => {
  const token = req.cookies['access-token'];
  const decoded = jwt.verify(token, SECRET_KEY);
  const userId = decoded.sub;
  // Фильтровать расписание по userId
});
```

**Query Parameters**: Нет

**Response** `200 OK`:
```json
[
  {
    "dayOfWeek": "monday",
    "lessonNumber": 1,
    "lessons": [
      {
        "id": "lesson-1",
        "subject": {
          "id": "subject-1",
          "name": "Математика"
        },
        "teachers": [
          {
            "id": "teacher-1",
            "firstName": "Анна",
            "lastName": "Иванова",
            "patronymic": "Петровна"
          }
        ],
        "rooms": [
          {
            "id": "room-1",
            "name": "101"
          }
        ],
        "participants": [
          {
            "class": {
              "id": "class-1",
              "name": "5А",
              "classTeacher": {
                "id": "teacher-2",
                "firstName": "Мария",
                "lastName": "Петрова",
                "patronymic": "Александровна"
              },
              "subjects": [],
              "groups": []
            },
            "groupIds": ["group-1"]
          }
        ]
      }
    ]
  }
]
```

**Логика фильтрации**:
1. Извлечь `userId` из JWT токена
2. Определить `teacher_id` по `userId` из таблицы `teachers`
3. Вернуть только уроки, где `teacher_id` присутствует в таблице `lesson_teachers`

**SQL Example**:
```sql
SELECT 
  ss.day_of_week,
  ss.lesson_number,
  sl.id as lesson_id,
  s.id as subject_id,
  s.name as subject_name,
  -- учителя
  t.id as teacher_id,
  t.first_name,
  t.last_name,
  t.patronymic,
  -- кабинеты
  cr.id as classroom_id,
  cr.name as classroom_name,
  -- классы
  c.id as class_id,
  c.name as class_name
FROM schedule_slots ss
JOIN schedule_lessons sl ON sl.schedule_slot_id = ss.id
JOIN lesson_teachers lt ON lt.lesson_id = sl.id
JOIN teachers t ON t.id = lt.teacher_id
JOIN subjects s ON s.id = sl.subject_id
JOIN lesson_participants lp ON lp.lesson_id = sl.id
JOIN classes c ON c.id = lp.class_id
LEFT JOIN lesson_rooms lr ON lr.lesson_id = sl.id
LEFT JOIN classrooms cr ON cr.id = lr.classroom_id
WHERE lt.teacher_id = (
  SELECT id FROM teachers WHERE user_id = $userId
)
ORDER BY ss.day_of_week, ss.lesson_number;
```

---

### `PUT /schedule`
**Описание**: Сохранить/обновить расписание (только для администратора)

**Headers**:
```
Cookie: access-token=<jwt_token>
Content-Type: application/json
```

**Body**:
```json
[
  {
    "dayOfWeek": "monday",
    "lessonNumber": 1,
    "lessons": [
      {
        "subject": {
          "id": "subject-1",
          "name": "Математика"
        },
        "teachers": [
          {
            "id": "teacher-1",
            "firstName": "Анна",
            "lastName": "Иванова",
            "patronymic": "Петровна"
          }
        ],
        "rooms": [
          {
            "id": "room-1",
            "name": "101"
          }
        ],
        "participants": [
          {
            "class": {
              "id": "class-1",
              "name": "5А"
            },
            "groupIds": ["group-1"]
          }
        ]
      }
    ]
  }
]
```

**Response** `200 OK`:
```json
{
  "message": "Расписание успешно сохранено"
}
```

**Response** `400 Bad Request` (при конфликтах):
```json
{
  "error": "Конфликт расписания",
  "details": [
    {
      "type": "teacher_conflict",
      "message": "Учитель Иванова А.П. уже занят в это время",
      "dayOfWeek": "monday",
      "lessonNumber": 1
    }
  ]
}
```

**Валидация**:
1. Проверить роль пользователя (только `admin` или `scheduler`)
2. Валидация конфликтов:
   - ❌ Учитель не может вести два урока одновременно
   - ❌ Кабинет не может быть занят дважды
   - ❌ Класс/группа не может иметь два урока в одно время
3. Транзакция:
   - Удалить все старые уроки текущего расписания
   - Вставить новые уроки
   - Обновить связи (lesson_teachers, lesson_rooms, lesson_participants)

---

### `POST /schedule/generate`
**Описание**: Автоматическая генерация расписания (только для администратора)

**Headers**:
```
Cookie: access-token=<jwt_token>
Content-Type: application/json
```

**Body** (опционально):
```json
{
  "algorithm": "greedy",
  "maxLessonsPerDay": 7,
  "priorities": {
    "balanceWorkload": true,
    "minimizeGaps": true
  }
}
```

**Response** `200 OK`:
```json
[
  {
    "dayOfWeek": "monday",
    "lessonNumber": 1,
    "lessons": [...]
  }
]
```

**Response** `400 Bad Request`:
```json
{
  "error": "Невозможно сгенерировать расписание",
  "reason": "Недостаточно учителей для покрытия всех уроков"
}
```

**Логика генерации**:
1. Проверить роль пользователя (только `admin`)
2. Загрузить:
   - Учебный план (study_plans)
   - Нагрузку учителей (teacher_workload)
   - Доступные кабинеты (classrooms)
3. Запустить алгоритм генерации:
   - Распределить уроки по дням недели
   - Учесть ограничения (максимум уроков в день)
   - Избежать конфликтов
   - Оптимизировать нагрузку учителей
4. Вернуть сгенерированное расписание (не сохранять автоматически)

---

### `GET /schedule/:id`
**Описание**: Получить конкретное расписание по ID (для версионирования/архива)

**Response** `200 OK`:
```json
{
  "id": "schedule-1",
  "name": "Расписание на 1 семестр 2024",
  "scheduleSlots": [...]
}
```

---

### `POST /schedule`
**Описание**: Создать новое именованное расписание

**Body**:
```json
{
  "name": "Расписание на 2 семестр 2024",
  "scheduleSlots": [...]
}
```

**Response** `201 Created`:
```json
{
  "id": "schedule-2",
  "name": "Расписание на 2 семестр 2024",
  "scheduleSlots": [...]
}
```

---

### `DELETE /schedule/:id`
**Описание**: Удалить расписание по ID

**Response** `204 No Content`

---

## Классы (Classes)

### `GET /classes`
**Описание**: Получить все классы

**Response** `200 OK`:
```json
[
  {
    "id": "class-1",
    "name": "5А",
    "classTeacher": {
      "id": "teacher-1",
      "firstName": "Анна",
      "lastName": "Иванова",
      "patronymic": "Петровна"
    },
    "subjects": [
      {
        "subject": {
          "id": "subject-1",
          "name": "Математика"
        },
        "hoursPerWeek": 5,
        "split": {
          "groupsCount": 2,
          "crossClassAllowed": false
        }
      }
    ],
    "groups": [
      {
        "id": "group-1",
        "name": "Группа 1",
        "size": 15
      }
    ]
  }
]
```

---

### `POST /classes`
**Описание**: Создать новый класс

**Body**:
```json
{
  "name": "5А"
}
```

**Response** `201 Created`:
```json
{
  "id": "class-1",
  "name": "5А",
  "classTeacher": null,
  "subjects": [],
  "groups": []
}
```

---

### `DELETE /classes/:id`
**Описание**: Удалить класс

**Response** `204 No Content`

---

### `PUT /classes/bulk`
**Описание**: Массовое обновление классов (учебный план, группы)

**Body**:
```json
[
  {
    "id": "class-1",
    "name": "5А",
    "classTeacher": {
      "id": "teacher-1"
    },
    "subjects": [
      {
        "subject": {
          "id": "subject-1",
          "name": "Математика"
        },
        "hoursPerWeek": 5,
        "split": {
          "groupsCount": 2,
          "crossClassAllowed": false
        }
      }
    ],
    "groups": [
      {
        "id": "group-1",
        "name": "Группа 1",
        "size": 15
      }
    ]
  }
]
```

**Response** `200 OK`:
```json
{
  "message": "Классы успешно обновлены",
  "updated": 5
}
```

---

## Учителя (Teachers)

### `GET /users/Teachers`
**Описание**: Получить всех учителей (полная информация)

**Response** `200 OK`:
```json
[
  {
    "id": "teacher-1",
    "firstName": "Анна",
    "lastName": "Иванова",
    "patronymic": "Петровна",
    "subjects": [
      {
        "subject": {
          "id": "subject-1",
          "name": "Математика"
        },
        "hoursPerWeek": 20
      }
    ],
    "classRoom": {
      "id": "room-1",
      "name": "101"
    },
    "class": {
      "id": "class-1",
      "name": "5А"
    },
    "workloadHoursPerWeek": 18,
    "classHours": [
      {
        "class": {
          "id": "class-1",
          "name": "5А"
        },
        "subject": {
          "id": "subject-1",
          "name": "Математика"
        },
        "groupId": "group-1",
        "hours": 3
      }
    ]
  }
]
```

---

### `GET /users/LightTeachers`
**Описание**: Получить список учителей (облегченная версия, только ФИО)

**Response** `200 OK`:
```json
[
  {
    "id": "teacher-1",
    "firstName": "Анна",
    "lastName": "Иванова",
    "patronymic": "Петровна"
  }
]
```

---

### `POST /users/Teachers`
**Описание**: Создать нового учителя

**Body**:
```json
{
  "firstName": "Анна",
  "lastName": "Иванова",
  "patronymic": "Петровна"
}
```

**Response** `201 Created`:
```json
{
  "id": "teacher-1",
  "firstName": "Анна",
  "lastName": "Иванова",
  "patronymic": "Петровна",
  "subjects": [],
  "classRoom": null,
  "class": null,
  "workloadHoursPerWeek": 0,
  "classHours": []
}
```

---

### `DELETE /users/Teachers/:id`
**Описание**: Удалить учителя

**Response** `204 No Content`

---

### `PATCH /users/Teachers/bulk`
**Описание**: Массовое обновление учителей (нагрузка, классы)

**Body**:
```json
[
  {
    "id": "teacher-1",
    "firstName": "Анна",
    "lastName": "Иванова",
    "patronymic": "Петровна",
    "subjects": [
      {
        "subject": {
          "id": "subject-1",
          "name": "Математика"
        },
        "hoursPerWeek": 20
      }
    ],
    "classRoom": {
      "id": "room-1",
      "name": "101"
    },
    "class": {
      "id": "class-1",
      "name": "5А"
    },
    "workloadHoursPerWeek": 18,
    "classHours": [
      {
        "class": {
          "id": "class-1",
          "name": "5А"
        },
        "subject": {
          "id": "subject-1",
          "name": "Математика"
        },
        "groupId": "group-1",
        "hours": 3
      }
    ]
  }
]
```

**Response** `200 OK`:
```json
{
  "message": "Учителя успешно обновлены",
  "updated": 3
}
```

---

## Предметы (Subjects)

### `GET /subjects`
**Описание**: Получить все предметы

**Response** `200 OK`:
```json
[
  {
    "id": "subject-1",
    "name": "Математика"
  }
]
```

---

### `POST /subjects`
**Описание**: Создать новый предмет

**Body**:
```json
{
  "name": "Математика"
}
```

**Response** `201 Created`:
```json
{
  "id": "subject-1",
  "name": "Математика"
}
```

---

### `DELETE /subjects/:id`
**Описание**: Удалить предмет

**Response** `204 No Content`

---

## Кабинеты (Classrooms)

### `GET /classrooms`
**Описание**: Получить все кабинеты

**Response** `200 OK`:
```json
[
  {
    "id": "room-1",
    "name": "101"
  }
]
```

---

### `POST /classrooms`
**Описание**: Создать новый кабинет

**Body**:
```json
{
  "name": "101"
}
```

**Response** `201 Created`:
```json
{
  "id": "room-1",
  "name": "101"
}
```

---

### `DELETE /classrooms/:id`
**Описание**: Удалить кабинет

**Response** `204 No Content`

---

## Типы данных

### WeekDaysCode (enum)
```typescript
"monday" | "tuesday" | "wednesday" | "thursday" | "friday" | "saturday" | "sunday"
```

---

## Примечания для бэкенда

### 1. Авторизация через httpOnly cookies
```javascript
// Express.js пример
const cookieParser = require('cookie-parser');
app.use(cookieParser());

app.get('/schedule', (req, res) => {
  const token = req.cookies['access-token'];
  if (!token) return res.status(401).json({ error: 'No token' });
  
  const decoded = jwt.verify(token, process.env.JWT_SECRET);
  const userId = decoded.sub;
  const userRole = decoded.role;
  
  // Фильтровать данные по userId и role
});
```

### 2. CORS настройки
```javascript
app.use(cors({
  origin: 'http://localhost:3000',
  credentials: true  // ✅ ВАЖНО для cookies
}));
```

### 3. Фильтрация расписания
```sql
-- Для учителей: только их уроки
SELECT * FROM schedule_lessons sl
JOIN lesson_teachers lt ON lt.lesson_id = sl.id
WHERE lt.teacher_id = (SELECT id FROM teachers WHERE user_id = $userId)

-- Для админов: все расписание
SELECT * FROM schedule_lessons
```

### 4. Валидация конфликтов
```sql
-- Конфликт учителя
SELECT COUNT(*) FROM schedule_lessons sl1
JOIN lesson_teachers lt1 ON lt1.lesson_id = sl1.id
JOIN schedule_slots ss1 ON ss1.id = sl1.schedule_slot_id
WHERE lt1.teacher_id = $teacherId
  AND ss1.day_of_week = $dayOfWeek
  AND ss1.lesson_number = $lessonNumber
-- Если COUNT > 0 → конфликт
```

### 5. Коды ошибок

| Код | Описание |
|-----|----------|
| `200` | Успех |
| `201` | Создано |
| `204` | Удалено (нет контента) |
| `400` | Неверный запрос |
| `401` | Не авторизован |
| `403` | Нет прав |
| `404` | Не найдено |
| `409` | Конфликт |
| `500` | Ошибка сервера |

---

## Примеры curl

### Login
```bash
curl -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "teacher@school.com", "password": "pass123"}' \
  -c cookies.txt
```

### Получить расписание
```bash
curl -X GET http://localhost:3000/api/schedule \
  -b cookies.txt
```

### Сохранить расписание
```bash
curl -X PUT http://localhost:3000/api/schedule \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -d @schedule.json
```

---

**Дата**: 24.11.2025 | **Версия**: 1.0
