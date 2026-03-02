# SHIBA

Веб-платформа для взаимодействия моделей и агентств (работодателей).

---

## Быстрый старт

> Требования: Docker Desktop запущен

**Шаг 1 — Склонировать и настроить**
```bash
cp .env.example .env
# Открыть .env и заменить два обязательных JWT-секрета:
#   JWT_ACCESS_SECRET=минимум-32-символа-любой-текст
#   JWT_REFRESH_SECRET=другой-секрет-минимум-32-символа
```

**Шаг 2 — Собрать и запустить**
```bash
docker compose up -d --build
# Первый запуск: ~3-5 минут (скачивает образы, собирает Flutter web)
```

**Шаг 3 — Проверить что всё работает**
```bash
docker compose ps          # все сервисы должны быть healthy/running
curl http://localhost/health   # → {"data":{"status":"ok",...}}
```

**Шаг 4 — Открыть в браузере**
- http://localhost — Flutter Web (страница логина)
- http://localhost:9001 — MinIO Console (minioadmin / minioadmin_secret)

**Шаг 5 — Создать администратора**
```bash
# Зарегистрировать пользователя (role=model, т.к. admin недоступна через API):
curl -X POST http://localhost/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@shiba.app","password":"Admin1234!","role":"model"}'

# Повысить роль до admin и активировать аккаунт в БД:
docker compose exec postgres psql -U shiba -d shiba \
  -c "UPDATE users SET role='admin', status='active' WHERE email='admin@shiba.app';"
```

---

## Содержание

1. [Описание проекта](#1-описание-проекта)
2. [Технологический стек](#2-технологический-стек)
3. [Архитектура](#3-архитектура)
4. [Схема базы данных](#4-схема-базы-данных)
5. [API — полный список эндпоинтов](#5-api--полный-список-эндпоинтов)
6. [Email-уведомления](#6-email-уведомления)
7. [Что реализовано (V1 MVP)](#7-что-реализовано-v1-mvp)
8. [Что предстоит сделать](#8-что-предстоит-сделать)
9. [Запуск проекта локально](#9-запуск-проекта-локально)
10. [Конфигурация](#10-конфигурация)
11. [Тест-сценарий (smoke test)](#11-тест-сценарий-smoke-test)

---

## 1. Описание проекта

**SHIBA** — платформа для взаимодействия моделей и модельных агентств.

**Ключевые возможности:**

| Роль | Возможности |
|------|------------|
| **Модель** | Создать профиль с параметрами тела; загрузить портфолио фотографий; искать кастинги и откликаться на них; получать прямые приглашения от агентств |
| **Агентство** | Зарегистрироваться (с ручным подтверждением admin); публиковать кастинги; искать моделей по фильтрам; отправлять прямые приглашения; управлять откликами |
| **Администратор** | Модерировать фотографии; одобрять/отклонять агентства; управлять пользователями; обрабатывать жалобы |

**Роли:** `admin`, `agency`, `model`

---

## 2. Технологический стек

### Backend

| Компонент | Технология | Версия |
|-----------|-----------|--------|
| Язык | Go | 1.22 |
| HTTP-фреймворк | Gin | v1.9.1 |
| База данных | PostgreSQL | 16 (alpine) |
| Драйвер БД | pgx/v5 | v5.6.0 |
| Миграции | golang-migrate/v4 | v4.17.1 |
| JWT | golang-jwt/v5 | v5.2.1 |
| Хранилище файлов | MinIO (S3-совместимый) | latest |
| MinIO-клиент | minio-go/v7 | v7.0.70 |
| Email | gomail.v2 | v2.0.0 |
| Хэширование | bcrypt (golang.org/x/crypto) | v0.24.0 |
| Rate limiting | golang.org/x/time/rate | v0.5.0 |
| CORS | gin-contrib/cors | v1.7.2 |
| UUID | google/uuid | v1.6.0 |

### Frontend

| Компонент | Технология | Версия |
|-----------|-----------|--------|
| Фреймворк | Flutter Web | >=3.19.0 |
| State management | flutter_bloc / Cubit | ^8.1.5 |
| HTTP-клиент | Dio | ^5.4.3+1 |
| Навигация | GoRouter | ^13.2.0 |
| DI-контейнер | GetIt | ^7.7.0 |
| Хранение токенов | flutter_secure_storage | ^9.0.0 |
| Изображения | cached_network_image | ^3.3.1 |
| Загрузка фото | image_picker | ^1.0.7 |
| Локализация | intl | ^0.19.0 |

### Инфраструктура

| Сервис | Image | Порт |
|--------|-------|------|
| API (Go) | multi-stage Dockerfile | 8080 |
| PostgreSQL | postgres:16-alpine | 5432 |
| MinIO | minio/minio:latest | 9000 (API), 9001 (Console) |
| Nginx | nginx:alpine | 80 |

---

## 3. Архитектура

### Структура репозитория

```
SHIBA/
├── cmd/api/main.go                     # Точка входа, DI, graceful shutdown
├── internal/
│   ├── config/config.go                # Конфигурация из env-переменных
│   ├── domain/                         # Доменные структуры (без зависимостей)
│   │   ├── user.go                     # User, EmailVerification, RefreshToken
│   │   ├── profile.go                  # ModelProfile, AgencyProfile, ModelPhoto
│   │   ├── casting.go                  # Casting, Application, Invitation
│   │   └── complaint.go                # Complaint
│   ├── repository/                     # Слой данных (SQL через pgx)
│   │   ├── postgres.go                 # Пул соединений pgxpool
│   │   ├── user_repo.go
│   │   ├── model_repo.go
│   │   ├── agency_repo.go
│   │   ├── casting_repo.go
│   │   └── complaint_repo.go
│   ├── service/                        # Бизнес-логика
│   │   ├── auth_service.go
│   │   ├── model_service.go
│   │   ├── agency_service.go
│   │   ├── casting_service.go
│   │   ├── admin_service.go
│   │   └── complaint_service.go
│   ├── handler/                        # HTTP-обработчики
│   │   ├── router.go                   # Маршруты + middleware-цепочки
│   │   ├── auth_handler.go
│   │   ├── model_handler.go
│   │   ├── agency_handler.go
│   │   ├── casting_handler.go          # + ApplicationHandler, InvitationHandler
│   │   ├── admin_handler.go
│   │   ├── complaint_handler.go
│   │   └── ref_handler.go
│   ├── dto/                            # Request/Response структуры
│   │   ├── auth.go
│   │   ├── model.go
│   │   ├── agency.go
│   │   ├── casting.go
│   │   └── admin.go
│   ├── middleware/
│   │   ├── auth.go                     # JWT-проверка + извлечение claims
│   │   ├── cors.go                     # CORS-политика
│   │   └── ratelimit.go                # Token bucket: 10 RPS / burst 20
│   └── storage/minio.go                # MinIO-клиент (upload, delete, URL)
├── pkg/
│   ├── hash/hash.go                    # bcrypt (пароли) + SHA-256 (токены)
│   ├── jwt/jwt.go                      # Access + Refresh token manager
│   ├── mail/mail.go                    # SMTP sender + HTML-шаблоны писем
│   └── response/response.go            # Gin helpers: OK, BadRequest, NotFound…
├── migrations/
│   ├── 000001_init.up.sql
│   └── 000001_init.down.sql
├── frontend/
│   └── lib/
│       ├── main.dart
│       ├── core/
│       │   ├── api/api_client.dart     # Dio + auth interceptor (auto-refresh)
│       │   ├── di/injection.dart       # GetIt регистрация
│       │   ├── router/app_router.dart  # GoRouter + MainShell с Drawer
│       │   └── theme/app_theme.dart    # Dark navy + red accent
│       └── features/
│           ├── auth/                   # BLoC + datasource + 3 страницы
│           ├── model_profile/          # Cubit + страница
│           ├── agency_profile/         # Cubit + страница
│           ├── search/                 # BLoC + страница
│           ├── castings/               # Cubit + 3 страницы
│           ├── invitations/            # Cubit + страница
│           └── admin/                  # Cubit + dashboard-страница
├── docker-compose.yml
├── Dockerfile                          # Multi-stage: builder (Go) → alpine
├── nginx.conf                          # /api/ → Go :8080, / → Flutter Web
├── Makefile
└── .env.example
```

### Слои backend

```
HTTP Request
    ↓
Middleware (CORS → RateLimit → Auth → RequireRoles)
    ↓
Handler  (парсинг, валидация, ответ)
    ↓
Service  (бизнес-логика, авторизация владельца)
    ↓
Repository  (SQL через pgx)  ←→  PostgreSQL
    ↓ (параллельно)
Storage (MinIO)  /  MailSender (SMTP)
```

### Паттерн frontend

```
Page (Widget)
    → BlocProvider / CubitProvider
    → BLoC / Cubit
    → ApiClient (Dio + auth interceptor)
    → Backend API
    ← State обновляет UI через BlocBuilder
```

### Ключевые архитектурные решения

| Решение | Обоснование |
|---------|-------------|
| Одна роль `agency` для агентств | Упрощение MVP; одинаковые права |
| Ручная активация агентств (admin) | Защита от ботов |
| Фото: max 10 шт., max 10 MB, JPEG/PNG | Ограничения хранилища |
| Soft-delete (`status = blocked`) | Возможность разблокировки |
| JWT access (15 мин) + refresh (720 ч) с ротацией | Безопасность + UX |
| Refresh token: SHA-256 для хранения хэша | Детерминированный поиск по DB |
| MinIO публичная read-политика бакета | Прямая раздача фото через URL |
| Rate limit: 10 RPS / burst 20 per IP | Защита от DoS |
| Миграции при старте сервера | Auto-migration в Docker Compose |
| `go func()` для email-уведомлений | Не блокирует HTTP-ответ |

---

## 4. Схема базы данных

### Перечисления (Enums)

```sql
user_role:             admin | agency | model
user_status:           pending_email | active | blocked
agency_status:         pending_approval | active | blocked
moderation_status:     pending | approved | rejected
casting_status:        active | closed | cancelled
application_status:    pending | viewed | accepted | rejected
complaint_target_type: casting | invitation | agency | model
complaint_reason:      inappropriate_content | harassment | fraud | other
complaint_status:      open | reviewed | resolved | dismissed
```

### Таблицы

```
users               id, email, password_hash, role, status, created_at, updated_at

email_verifications id, user_id→users, token, expires_at, used_at

refresh_tokens      id, user_id→users, token_hash (SHA-256), expires_at, revoked_at

categories          id (SMALLSERIAL), name                           [справочник]
photo_tags          id (SMALLSERIAL), name                           [справочник]

model_profiles      id, user_id→users (UNIQUE), first_name, last_name,
                    birth_date, city, country, willing_to_relocate,
                    height_cm, weight_kg, chest_cm, waist_cm, hips_cm,
                    shoe_size, clothing_size, phone, bio, shoot_restrictions[]

model_categories    model_profile_id ↔ category_id                  [M2M]

agency_profiles     id, user_id→users (UNIQUE), company_name, description,
                    phone, logo_url, status (agency_status),
                    approved_by→users, approved_at

model_photos        id, model_profile_id→model_profiles, storage_key,
                    original_name, size_bytes, mime_type, width_px, height_px,
                    moderation_status, moderated_by→users, moderated_at,
                    rejection_reason, is_cover, sort_order

model_photo_tags    photo_id ↔ photo_tag_id                          [M2M]

castings            id, agency_profile_id→agency_profiles, title, description,
                    city, casting_date, category_id→categories, status

applications        id, casting_id→castings, model_profile_id→model_profiles,
                    status (application_status), model_message, agency_response
                    UNIQUE(casting_id, model_profile_id)

invitations         id, agency_profile_id→agency_profiles,
                    model_profile_id→model_profiles,
                    casting_id→castings (nullable),
                    message, status (application_status)

complaints          id, reporter_id→users, target_type, target_id,
                    reason_category, description, status,
                    reviewed_by→users, resolution_note, resolved_at
```

---

## 5. API — полный список эндпоинтов

Base path: `/api/v1`

### System

| Метод | Путь | Доступ | Описание |
|-------|------|--------|----------|
| GET | `/health` | public | Health check |

### Auth `/auth`

| Метод | Путь | Доступ | Описание |
|-------|------|--------|----------|
| POST | `/auth/register` | public | Регистрация (`email`, `password`, `role`) |
| GET | `/auth/verify-email?token=` | public | Подтверждение email |
| POST | `/auth/login` | public | Вход → `access_token` + `refresh_token` |
| POST | `/auth/refresh` | public | Ротация токенов |
| POST | `/auth/logout` | any auth | Отзыв всех refresh-токенов пользователя |

### Reference `/ref`

| Метод | Путь | Доступ | Описание |
|-------|------|--------|----------|
| GET | `/ref/categories` | public | Справочник категорий моделей |
| GET | `/ref/photo-tags` | public | Справочник тегов фотографий |

### Models `/models`

| Метод | Путь | Доступ | Описание |
|-------|------|--------|----------|
| GET | `/models` | agency, admin | Поиск (`city`, `min_age`, `max_age`, `category_id`, `willing_to_relocate`, `limit`, `offset`) |
| GET | `/models/me` | model | Свой профиль (создаётся автоматически при первом запросе) |
| PUT | `/models/me` | model | Обновить профиль |
| POST | `/models/me/categories` | model | Добавить категорию |
| DELETE | `/models/me/categories/:id` | model | Удалить категорию |
| POST | `/models/me/photos` | model | Загрузить фото (multipart `photo`, max 10 MB, JPEG/PNG) |
| PUT | `/models/me/photos/:id` | model | Обновить фото (теги, порядок, is_cover) |
| DELETE | `/models/me/photos/:id` | model | Удалить фото |
| GET | `/models/:id` | any auth | Профиль модели по ID |
| GET | `/models/:id/photos` | any auth | Фото модели (агентство видит только `approved`) |

### Agencies `/agencies`

| Метод | Путь | Доступ | Описание |
|-------|------|--------|----------|
| GET | `/agencies/me` | agency | Свой профиль агентства |
| PUT | `/agencies/me` | agency | Обновить профиль |
| POST | `/agencies/me/logo` | agency | Загрузить логотип (multipart `logo`) |
| GET | `/agencies/me/invitations` | agency | Исходящие приглашения |
| GET | `/agencies/:id` | any auth | Профиль агентства по ID |

### Castings `/castings`

| Метод | Путь | Доступ | Описание |
|-------|------|--------|----------|
| POST | `/castings` | agency | Создать кастинг (требует `status=active` агентства) |
| GET | `/castings` | model, agency, admin | Список (`status`, `city`, `limit`, `offset`) |
| GET | `/castings/:id` | any auth | Детали кастинга |
| PUT | `/castings/:id` | agency (owner) | Обновить кастинг |
| DELETE | `/castings/:id` | agency (owner) | Удалить кастинг |
| GET | `/castings/:id/applications` | agency (owner) | Отклики на кастинг |
| PUT | `/castings/:id/applications/:app_id` | agency (owner) | Изменить статус отклика + ответ |

### Applications `/applications`

| Метод | Путь | Доступ | Описание |
|-------|------|--------|----------|
| POST | `/applications` | model | Откликнуться на кастинг (`casting_id`, `message`) |
| GET | `/applications/me` | model | Мои отклики |
| DELETE | `/applications/:id` | model (owner) | Отозвать отклик (только `pending`) |

### Invitations `/invitations`

| Метод | Путь | Доступ | Описание |
|-------|------|--------|----------|
| POST | `/invitations` | agency | Пригласить модель (`model_profile_id`, `message`, `casting_id?`) |
| GET | `/invitations/me` | model | Мои входящие приглашения |
| PUT | `/invitations/:id` | model (owner) | Ответить (`accepted` / `rejected`) |

### Complaints

| Метод | Путь | Доступ | Описание |
|-------|------|--------|----------|
| POST | `/complaints` | any auth | Подать жалобу (`target_type`, `target_id`, `reason_category`, `description`) |

### Admin `/admin`

| Метод | Путь | Доступ | Описание |
|-------|------|--------|----------|
| GET | `/admin/users` | admin | Список пользователей (`role`, `status`, `limit`, `offset`) |
| PUT | `/admin/users/:id/status` | admin | Изменить статус пользователя (`active` / `blocked`) |
| GET | `/admin/agencies/pending` | admin | Очередь агентств на одобрение |
| PUT | `/admin/agencies/:id/approve` | admin | Одобрить агентство |
| PUT | `/admin/agencies/:id/reject` | admin | Отклонить агентство |
| GET | `/admin/photos/pending` | admin | Очередь фото на модерацию |
| PUT | `/admin/photos/:id/moderate` | admin | Модерировать фото (`approved` / `rejected` + `rejection_reason`) |
| GET | `/admin/complaints` | admin | Список жалоб (`status`) |
| PUT | `/admin/complaints/:id` | admin | Изменить статус жалобы + `resolution_note` |

---

## 6. Email-уведомления

| Событие | Получатель | Шаблон |
|---------|-----------|--------|
| Регистрация | Новый пользователь | Ссылка верификации (expires 24 ч) |
| Новый отклик на кастинг | Агентство | Имя модели + название кастинга |
| Новое прямое приглашение | Модель | Название агентства + текст сообщения |
| Статус отклика изменён | Модель | Название кастинга + новый статус |
| Фото отклонено | Модель | Причина отклонения |

Письма отправляются **асинхронно** (`go func()`) — не блокируют HTTP-ответ.

---

## 7. Что реализовано (V1 MVP)

### Backend

- [x] Конфигурация из env-переменных (`internal/config`)
- [x] PostgreSQL pool через pgx/v5 (`internal/repository/postgres.go`)
- [x] Миграции при старте сервера (`golang-migrate`)
- [x] Доменные структуры (`internal/domain`)
- [x] Все репозитории с SQL-запросами (users, models, agencies, castings, applications, invitations, complaints)
- [x] Auth: регистрация, email-верификация, логин, refresh с ротацией (SHA-256), logout
- [x] JWT access (15 мин) + refresh (720 ч) с хранением хэша в БД
- [x] Model service: CRUD профиля, категории, загрузка/обновление/удаление фото
- [x] Agency service: CRUD профиля, загрузка логотипа
- [x] Casting service: CRUD кастингов, отклики, приглашения
- [x] Admin service: управление пользователями, модерация агентств и фото, жалобы
- [x] MinIO: загрузка фото/логотипов, удаление, публичный URL
- [x] Email: все 5 шаблонов уведомлений
- [x] Middleware: CORS, rate limiting (10 RPS / burst 20), JWT auth, role-based access
- [x] HTTP handlers для всех 45 эндпоинтов
- [x] Unified response envelope `{"data": ..., "error": ..., "message": ...}`
- [x] Graceful shutdown (SIGINT/SIGTERM, 10 сек)
- [x] Dockerfile multi-stage (builder Go 1.22 → alpine:3.19)
- [x] Docker Compose (api + postgres + minio + minio-init + nginx)
- [x] Nginx (proxy `/api/` → Go, `/` → Flutter Web, `client_max_body_size 15M`)
- [x] Makefile с целями: build, run, test, docker-up/down, dev-infra, flutter-web

### Frontend

- [x] Flutter Web (SDK >=3.3.0)
- [x] Dio ApiClient с auth interceptor (автоматический refresh при 401)
- [x] GetIt DI-контейнер (`core/di/injection.dart`)
- [x] GoRouter навигация с MainShell + Drawer (`core/router/app_router.dart`)
- [x] Dark navy + red accent тема (`core/theme/app_theme.dart`)
- [x] Auth BLoC + datasource: логин, регистрация, верификация email
- [x] Model profile Cubit + страница
- [x] Agency profile Cubit + страница
- [x] Search BLoC + страница (фильтры по городу, возрасту, категории)
- [x] Castings Cubit + 3 страницы: список, детали, создание
- [x] Invitations Cubit + страница (входящие + ответ)
- [x] Admin Cubit + dashboard (пользователи, агентства, фото, жалобы)

---

## 8. Что предстоит сделать

### V1.1 — Улучшения core-функций

#### Бэкенд

- [ ] **Видео-снепы** — загрузка видео-портфолио моделей
  - Транскодинг: FFmpeg → HLS (m3u8), max 60 сек / 100 MB
  - Новые таблицы: `model_videos`
  - Endpoints: `POST /models/me/videos`, `GET /models/:id/videos`, `DELETE /models/me/videos/:id`
- [ ] **Компкарта PDF** — генерация на основе профиля модели
  - Endpoint: `GET /models/me/compcard`
- [ ] **2FA (TOTP)** — Google Authenticator (`pquerna/otp`)
  - Endpoints: `POST /auth/2fa/setup`, `POST /auth/2fa/verify`, `DELETE /auth/2fa`
- [ ] **Избранное** — агентство сохраняет модели в shortlist
  - Новая таблица: `favorites(agency_profile_id, model_profile_id)`
  - Endpoints: `POST /favorites/:model_id`, `DELETE /favorites/:model_id`, `GET /favorites`

#### Фронтенд

- [ ] Страница профиля модели: загрузка/удаление фото через API
- [ ] Страница профиля агентства: загрузка логотипа
- [ ] Детальная страница кастинга: кнопка "Откликнуться"
- [ ] Страница поиска: полная интеграция фильтров + пагинация
- [ ] Admin dashboard: полная интеграция всех вкладок
- [ ] Просмотр публичного профиля модели/агентства
- [ ] Разделение Drawer по роли пользователя

### V1.2 — Аналитика и обогащение

- [ ] **История просмотров профиля** — модель видит кто просматривал
  - Endpoint: `GET /models/me/profile-views`
- [ ] **Статистика кастингов** — воронка откликов для агентства
  - Endpoint: `GET /castings/:id/stats`
- [ ] **Audit log** — все действия администраторов
  - Endpoint: `GET /admin/audit-log`
- [ ] **Расширенные фильтры поиска** — тип внешности, размер одежды/обуви

### V2 — Мобильные приложения + монетизация

- [ ] **iOS / Android** — Flutter multi-target build
- [ ] **Firebase Cloud Messaging** — push-уведомления
- [ ] **Подписки** — premium для агентств (больше приглашений, featured)
- [ ] **Рейтинг и отзывы** — модели оценивают агентства
- [ ] **Чат** — внутренний мессенджер (WebSocket)
- [ ] **Верификация личности** — загрузка документов + ручная проверка

### Технический долг

- [ ] Тесты (unit + integration) для сервисов и репозиториев
- [ ] Swagger / OpenAPI документация для API
- [ ] Structured logging (zap или slog вместо `log.Printf`)
- [ ] Метрики (Prometheus) + трейсинг (OpenTelemetry)
- [ ] CI/CD pipeline (GitHub Actions: lint → test → build → push)
- [ ] Helm chart для Kubernetes-деплоя

---

## 9. Запуск проекта локально

### Требования

| Инструмент | Версия |
|-----------|--------|
| Docker | >= 24.0 |
| Docker Compose | >= 2.20 (встроен в Docker Desktop) |
| Go | 1.22+ (только для локальной разработки без Docker) |
| Flutter | >= 3.19.0 (только для frontend-разработки) |

### Способ 1: Полный стек в Docker (рекомендуется)

```bash
# 1. Клонировать репозиторий
git clone <repo-url>
cd SHIBA

# 2. Создать .env из шаблона
cp .env.example .env

# 3. Обязательно задать JWT-секреты в .env
#    JWT_ACCESS_SECRET=<минимум 32 символа>
#    JWT_REFRESH_SECRET=<минимум 32 символа>

# 4. Запустить все сервисы (включая сборку Flutter web)
make docker-up
# или: docker compose up -d --build
# Первый запуск займёт ~3-5 минут: скачивает образы и собирает Flutter web

# 5. Проверить, что всё запустилось
docker compose ps
curl http://localhost/health
```

**Доступные URL:**

| Сервис | URL |
|--------|-----|
| Frontend (Flutter Web) | http://localhost |
| API | http://localhost/api/v1 |
| Health check | http://localhost/health |
| MinIO Console | http://localhost:9001 |
| MinIO API | http://localhost:9000 |
| PostgreSQL | localhost:5432 |

**Учётные данные по умолчанию:**

| Сервис | Логин | Пароль |
|--------|-------|--------|
| PostgreSQL | `shiba` | `shiba_secret` |
| MinIO | `minioadmin` | `minioadmin_secret` |

**Демо-аккаунты (заполняются через `seed.sql`):**

| Роль | Email | Пароль |
|------|-------|--------|
| Admin | `admin@shiba.test` | `Admin1234!` |
| Agency 1 | `agency1@shiba.test` | `Agency1234!` |
| Agency 2 | `agency2@shiba.test` | `Agency1234!` |
| Agency 3 | `agency3@shiba.test` | `Agency1234!` |
| Model 1 | `model1@shiba.test` | `Model1234!` |
| Model 2 | `model2@shiba.test` | `Model1234!` |
| Model 3 | `model3@shiba.test` | `Model1234!` |
| Model 4 | `model4@shiba.test` | `Model1234!` |
| Model 5 | `model5@shiba.test` | `Model1234!` |
| Model 6 | `model6@shiba.test` | `Model1234!` |

---

### Способ 2: Только инфраструктура + локальный Go-сервер

```bash
# 1. Поднять только postgres + minio
make dev-infra
# или: docker compose up -d postgres minio minio-init

# 2. Настроить .env для локального подключения
#    DB_HOST=localhost
#    MINIO_ENDPOINT=localhost:9000

# 3. Запустить сервер
make run
# или: go run ./cmd/api/main.go

# Сервер стартует на :8080
```

---

### Способ 3: Локальная разработка frontend

```bash
# (Предполагает, что backend уже запущен — Способ 1 или 2)

cd frontend

# Установить зависимости
flutter pub get

# Запустить в браузере Chrome
flutter run -d chrome

# Или собрать production-версию
flutter build web --release
```

---

### Создание администратора

В MVP нет endpoint для автоматического создания admin. Алгоритм:

```bash
# 1. Зарегистрировать пользователя (role=model, т.к. admin недоступна через API)
curl -X POST http://localhost/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@shiba.app","password":"AdminPass123!","role":"model"}'

# 2. Повысить роль до admin и активировать аккаунт в БД (минуя email-верификацию)
docker compose exec postgres psql -U shiba -d shiba \
  -c "UPDATE users SET role='admin', status='active' WHERE email='admin@shiba.app';"

# 3. Залогиниться
curl -X POST http://localhost/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@shiba.app","password":"AdminPass123!"}'
```

---

### Полезные команды

```bash
# Посмотреть логи API
make docker-logs

# Перезапустить только API (после изменений в коде)
make docker-restart

# Остановить все сервисы
make docker-down

# Полный сброс (ВНИМАНИЕ: удаляет все данные)
make docker-reset

# Собрать бинарник
make build   # → bin/server

# Запустить тесты
make test

# Форматирование кода
make fmt

# Собрать Flutter web
make flutter-web
```

---

### Сброс пароля пользователя

> **Важно:** пароли хранятся как bcrypt-хеши через Go (`golang.org/x/crypto/bcrypt`, cost 12).
> SQL-функция `pgcrypto crypt()` генерирует **несовместимый** формат хеша (cost 6) —
> Go-backend не сможет его верифицировать и вернёт 401.

Единственный корректный способ сменить пароль вручную — сгенерировать хеш через Go:

```bash
# 1. Создать временный Go-скрипт для генерации хеша
mkdir -p /tmp/genhash
cp go.mod go.sum /tmp/genhash/
sed -i 's/^module .*/module genhash/' /tmp/genhash/go.mod

cat > /tmp/genhash/main.go << 'EOF'
package main
import ("fmt"; "golang.org/x/crypto/bcrypt")
func main() {
    b, _ := bcrypt.GenerateFromPassword([]byte("НовыйПароль123!"), 12)
    fmt.Print(string(b))
}
EOF

# 2. Сгенерировать хеш
HASH=$(cd /tmp/genhash && go run main.go)

# 3. Записать в БД
docker compose exec postgres psql -U shiba -d shiba \
  -c "UPDATE users SET password_hash = '$HASH' WHERE email = 'user@example.com';"
```

---

### Возможные проблемы при запуске

**API не стартует: `required env variable JWT_ACCESS_SECRET is not set`**
```bash
# Добавьте секреты в .env
echo 'JWT_ACCESS_SECRET=your-secret-key-at-least-32-chars' >> .env
echo 'JWT_REFRESH_SECRET=your-other-secret-32-chars-min' >> .env
```

**MinIO недоступен / корзина не создалась**
```bash
# Проверить статус minio-init
docker compose logs minio-init
# Повторно создать корзину вручную
docker compose run --rm minio-init
```

**PostgreSQL не принимает подключения**
```bash
# Проверить health
docker compose ps postgres
docker compose logs postgres
```

**Flutter: `flutter: command not found`**
Установить Flutter SDK: https://docs.flutter.dev/get-started/install

---

## 10. Конфигурация

Все настройки передаются через переменные окружения (файл `.env`).

```bash
# ── Приложение ─────────────────────────────────────────────────────────────
APP_PORT=8080
APP_ENV=development        # production → gin.ReleaseMode
FRONTEND_URL=http://localhost:80

# ── База данных ─────────────────────────────────────────────────────────────
DB_HOST=postgres           # postgres — имя сервиса в Docker Compose
DB_PORT=5432
DB_USER=shiba
DB_PASSWORD=shiba_secret
DB_NAME=shiba
DB_SSLMODE=disable

# ── JWT (ОБЯЗАТЕЛЬНЫЕ) ───────────────────────────────────────────────────────
JWT_ACCESS_SECRET=         # минимум 32 символа, замените перед деплоем
JWT_REFRESH_SECRET=        # минимум 32 символа, замените перед деплоем
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=720h       # 30 дней

# ── MinIO ───────────────────────────────────────────────────────────────────
MINIO_ENDPOINT=minio:9000  # minio — имя сервиса в Docker Compose
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin_secret
MINIO_BUCKET=shiba
MINIO_USE_SSL=false
MINIO_PUBLIC_URL=http://localhost:9000

# ── SMTP ────────────────────────────────────────────────────────────────────
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password    # App Password для Gmail
SMTP_FROM=noreply@shiba.app

# ── Rate Limiting ────────────────────────────────────────────────────────────
RATE_LIMIT_RPS=10          # запросов в секунду с одного IP
RATE_LIMIT_BURST=20        # максимальный burst
```

> **Важно для production**: перед деплоем обязательно замените `JWT_ACCESS_SECRET`, `JWT_REFRESH_SECRET`, пароли БД и MinIO на стойкие значения.

---

## 11. Тест-сценарий (smoke test)

Базовый сценарий для проверки работы всех компонентов после запуска:

```bash
BASE=http://localhost/api/v1

# ── 1. Регистрация модели ────────────────────────────────────────────────────
curl -s -X POST $BASE/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"model@test.com","password":"Test1234!","role":"model"}'

# Подтвердить email вручную в БД:
docker compose exec postgres psql -U shiba -d shiba \
  -c "UPDATE users SET status='active' WHERE email='model@test.com';"

# Залогиниться (сохранить MODEL_TOKEN и MODEL_REFRESH)
curl -s -X POST $BASE/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"model@test.com","password":"Test1234!"}'

MODEL_TOKEN="<access_token из ответа>"

# Обновить профиль модели
curl -s -X PUT $BASE/models/me \
  -H "Authorization: Bearer $MODEL_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"first_name":"Anna","last_name":"Smith","city":"Moscow","height_cm":175}'

# ── 2. Регистрация агентства ─────────────────────────────────────────────────
curl -s -X POST $BASE/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"agency@test.com","password":"Test1234!","role":"agency"}'

docker compose exec postgres psql -U shiba -d shiba \
  -c "UPDATE users SET status='active' WHERE email='agency@test.com';"

# ── 3. Одобрение агентства через admin ───────────────────────────────────────
# Получить токен admin (см. раздел "Создание администратора")
ADMIN_TOKEN="<admin access_token>"

# Получить ID агентства
curl -s $BASE/admin/agencies/pending \
  -H "Authorization: Bearer $ADMIN_TOKEN"

AGENCY_PROFILE_ID="<id из ответа>"

curl -s -X PUT $BASE/admin/agencies/$AGENCY_PROFILE_ID/approve \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# ── 4. Агентство ищет модели и создаёт кастинг ──────────────────────────────
AGENCY_TOKEN="<agency access_token>"

curl -s "$BASE/models?city=Moscow" \
  -H "Authorization: Bearer $AGENCY_TOKEN"

curl -s -X POST $BASE/castings \
  -H "Authorization: Bearer $AGENCY_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Fashion Show","city":"Moscow","description":"Looking for models"}'

CASTING_ID="<id из ответа>"

# ── 5. Модель откликается на кастинг ─────────────────────────────────────────
curl -s -X POST $BASE/applications \
  -H "Authorization: Bearer $MODEL_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"casting_id\":\"$CASTING_ID\",\"message\":\"I am interested\"}"

APP_ID="<id из ответа>"

# ── 6. Агентство принимает отклик ────────────────────────────────────────────
curl -s -X PUT $BASE/castings/$CASTING_ID/applications/$APP_ID \
  -H "Authorization: Bearer $AGENCY_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status":"accepted","response":"We look forward to seeing you!"}'

# ── 7. Проверка безопасности ─────────────────────────────────────────────────
# Запрос admin-ресурса без admin-роли → 403 Forbidden
curl -s $BASE/admin/users \
  -H "Authorization: Bearer $MODEL_TOKEN"
```
