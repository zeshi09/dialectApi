# Инструкция по работе с базой данных

## Проблемы которые были исправлены

1. **file.py** - исправлен порядок полей `year` и `comment` при вставке в SQLite
2. Создана система автоматического обновления PostgreSQL из CSV файлов

## Структура проекта

- `chrononyms_fin.csv` - основные данные (хрононимы)
- `districts_coords.csv` - координаты для пар район-сельсовет
- `file.py` - создание SQLite базы (теперь исправлен)
- `migrate.py` - миграция из SQLite в PostgreSQL (старый способ)
- `update_db.py` - обновление PostgreSQL из CSV (новый способ)
- `init_db_docker.py` - инициализация PostgreSQL в Docker
- `Dockerfile.init` - Docker образ для инициализации
- `docker-compose.yml` - конфигурация Docker

## Первый запуск с Docker

```bash
# 1. Остановить и удалить старые контейнеры и данные
docker compose down -v

# 2. Запустить PostgreSQL и автоматическую инициализацию
docker compose up -d

# База данных автоматически заполнится из CSV файлов
```

## Добавление новых слов

### Способ 1: Добавить в CSV и пересобрать Docker

1. Добавьте новые строки в `chrononyms_fin.csv`
2. Если нужно, обновите координаты в `districts_coords.csv`
3. Пересоздайте базу:

```bash
# Удаляем старые данные
docker compose down -v

# Запускаем заново - база заполнится из обновленных CSV
docker compose up -d
```

### Способ 2: Обновить существующую базу (без пересоздания)

1. Добавьте новые строки в `chrononyms_fin.csv`
2. Запустите скрипт обновления:

```bash
# Установите psycopg2 если еще не установлен
pip install psycopg2-binary

# Запустите скрипт обновления
python3 update_db.py
```

Скрипт автоматически:
- Подключится к PostgreSQL
- Проверит существующие записи
- Добавит только новые записи (избегая дублирования)

## Проверка данных

```bash
# Подключиться к PostgreSQL
docker exec -it dialect_postgres psql -U postgres -d location

# Проверить количество записей
SELECT COUNT(*) FROM locations;

# Проверить последние добавленные записи
SELECT id, chrononym, district, selsovet
FROM locations
ORDER BY id DESC
LIMIT 10;

# Выйти
\q
```

## Порядок карточек на сайте

Карточки отображаются в порядке `ORDER BY id ASC`, то есть в порядке добавления в базу.

## Проверка порядка Year и Comment

Выполните в psql:

```sql
SELECT id, chrononym, comment, year
FROM locations
WHERE comment IS NOT NULL AND comment != ''
LIMIT 5;
```

Если поля перепутаны, исправьте через:

```sql
UPDATE locations
SET comment = year, year = comment
WHERE id > 0;
```

## Очистка базы данных

```bash
# Полная очистка
docker compose down -v

# Удаление только данных из таблицы (сохраняет структуру)
docker exec -it dialect_postgres psql -U postgres -d location -c "TRUNCATE TABLE locations RESTART IDENTITY;"
```

## Резервное копирование

```bash
# Создать backup
docker exec dialect_postgres pg_dump -U postgres location > backup_$(date +%Y%m%d).sql

# Восстановить из backup
cat backup_YYYYMMDD.sql | docker exec -i dialect_postgres psql -U postgres location
```
