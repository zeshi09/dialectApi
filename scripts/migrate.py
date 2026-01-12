#!/usr/bin/env python3
import sqlite3
import psycopg2
from psycopg2.extras import execute_batch
import time

# Параметры подключения
SQLITE_DB = "Locations_fin.db"
PG_HOST = "localhost"
PG_PORT = "5432"
PG_USER = "postgres"
PG_PASSWORD = "postgres"
PG_DATABASE = "location"

def migrate_data():
    print("Начинаем миграцию данных из SQLite в PostgreSQL...")

    # Подключение к SQLite
    sqlite_conn = sqlite3.connect(SQLITE_DB)
    sqlite_cursor = sqlite_conn.cursor()

    # Подключение к PostgreSQL
    pg_conn = psycopg2.connect(
        host=PG_HOST,
        port=PG_PORT,
        user=PG_USER,
        password=PG_PASSWORD,
        database=PG_DATABASE
    )
    pg_cursor = pg_conn.cursor()

    try:
        # Создание таблицы в PostgreSQL
        print("Создание таблицы locations в PostgreSQL...")
        pg_cursor.execute("""
            DROP TABLE IF EXISTS locations;
            CREATE TABLE locations (
                id SERIAL PRIMARY KEY,
                chrononym TEXT NOT NULL,
                definition TEXT NOT NULL,
                context TEXT NOT NULL,
                district TEXT NOT NULL,
                selsovet TEXT NOT NULL,
                latitude REAL,
                longitude REAL,
                comment TEXT,
                year TEXT NOT NULL,
                district_ss TEXT NOT NULL
            );
        """)
        pg_conn.commit()
        print("Таблица создана успешно!")

        # Получение данных из SQLite
        print("Чтение данных из SQLite...")
        sqlite_cursor.execute("SELECT * FROM locations")
        rows = sqlite_cursor.fetchall()
        print(f"Найдено {len(rows)} записей для миграции")

        # Подготовка данных для вставки (пропускаем id, т.к. используем SERIAL)
        insert_query = """
            INSERT INTO locations
            (chrononym, definition, context, district, selsovet, latitude, longitude, comment, year, district_ss)
            VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
        """

        # Преобразуем данные (убираем первый столбец - id)
        # И меняем местами comment (индекс 8) и year (индекс 9), так как в исходной SQLite они перепутаны
        data_to_insert = []
        for row in rows:
            row_list = list(row[1:])  # убираем id
            # Меняем местами comment (индекс 7) и year (индекс 8) в новом списке
            row_list[7], row_list[8] = row_list[8], row_list[7]
            data_to_insert.append(tuple(row_list))

        # Вставка данных пачками для производительности
        print("Вставка данных в PostgreSQL...")
        batch_size = 1000
        for i in range(0, len(data_to_insert), batch_size):
            batch = data_to_insert[i:i + batch_size]
            execute_batch(pg_cursor, insert_query, batch)
            pg_conn.commit()
            print(f"Обработано {min(i + batch_size, len(data_to_insert))} из {len(data_to_insert)} записей")

        # Проверка количества записей
        pg_cursor.execute("SELECT COUNT(*) FROM locations")
        count = pg_cursor.fetchone()[0]
        print(f"\nМиграция завершена! Всего записей в PostgreSQL: {count}")

        # Показать несколько примеров
        print("\nПримеры записей:")
        pg_cursor.execute("SELECT id, chrononym, district, selsovet FROM locations LIMIT 3")
        for row in pg_cursor.fetchall():
            print(f"  ID: {row[0]}, Хрононим: {row[1]}, Район: {row[2]}, Сельсовет: {row[3]}")

    except Exception as e:
        print(f"Ошибка при миграции: {e}")
        pg_conn.rollback()
        raise
    finally:
        sqlite_cursor.close()
        sqlite_conn.close()
        pg_cursor.close()
        pg_conn.close()

if __name__ == "__main__":
    print("Ожидание запуска PostgreSQL...")
    time.sleep(3)
    migrate_data()
