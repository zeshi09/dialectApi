#!/usr/bin/env python3
"""
Скрипт для обновления PostgreSQL из CSV файлов
Добавляет только новые записи, не дублирует существующие
"""

import csv
import psycopg2
from psycopg2.extras import execute_values
import sys

# Параметры подключения к PostgreSQL
DB_PARAMS = {
    'host': 'localhost',
    'port': 5432,
    'database': 'location',
    'user': 'postgres',
    'password': 'postgres'
}

def load_csv_data():
    """Загружает данные из CSV файлов"""
    print("Загрузка данных из CSV...")

    # Загружаем хрононимы
    data = []
    with open('data/chrononyms_fin.csv', mode='r', encoding="utf-8-sig") as file:
        reader = csv.DictReader(file)
        for row in reader:
            data.append(dict(row))

    # Загружаем координаты
    coords = []
    with open('data/districts_coords.csv', mode='r', encoding="utf-8-sig") as f:
        reader_coord = csv.DictReader(f)
        for row in reader_coord:
            coords.append(dict(row))

    # Сопоставляем координаты с хрононимами
    for item in data:
        for coord_item in coords:
            if item["SS"] == coord_item["SS"] and item["District"] == coord_item["Short_dis"]:
                item["Latitude"] = coord_item["Lat"]
                item["Longitude"] = coord_item["Long"]
                break

    print(f"Загружено {len(data)} записей из CSV")
    return data

def get_existing_chrononyms(conn):
    """Получает список существующих хрононимов из БД"""
    with conn.cursor() as cur:
        cur.execute("""
            SELECT chrononym, district, selsovet, definition
            FROM locations
        """)
        existing = set()
        for row in cur.fetchall():
            # Создаем уникальный ключ из основных полей
            key = (row[0], row[1], row[2], row[3])
            existing.add(key)
        return existing

def insert_new_records(conn, data):
    """Вставляет только новые записи"""
    existing = get_existing_chrononyms(conn)
    print(f"В базе уже есть {len(existing)} записей")

    new_records = []
    for item in data:
        # Создаем ключ для проверки
        key = (
            item["Chrononym"],
            item["District"],
            item["SS"],
            item["Def"]
        )

        # Если записи нет в базе, добавляем
        if key not in existing:
            lat = float(item.get("Latitude", 0)) if item.get("Latitude") else 0.0
            lon = float(item.get("Longitude", 0)) if item.get("Longitude") else 0.0
            diss = f"{item['District']}, {item['SS']}"

            new_records.append((
                item["Chrononym"],
                item["Def"],
                item.get("Context", ""),
                item["District"],
                item["SS"],
                lat,
                lon,
                item.get("Comment", ""),
                item.get("Year", ""),
                diss
            ))

    if new_records:
        print(f"Добавление {len(new_records)} новых записей...")
        with conn.cursor() as cur:
            execute_values(
                cur,
                """
                INSERT INTO locations
                (chrononym, definition, context, district, selsovet, latitude, longitude, comment, year, district_ss)
                VALUES %s
                """,
                new_records
            )
        conn.commit()
        print(f"✓ Успешно добавлено {len(new_records)} новых записей")
    else:
        print("Новых записей для добавления нет")

def main():
    try:
        # Подключаемся к PostgreSQL
        print("Подключение к PostgreSQL...")
        conn = psycopg2.connect(**DB_PARAMS)
        print("✓ Подключение успешно")

        # Загружаем данные из CSV
        csv_data = load_csv_data()

        # Вставляем новые записи
        insert_new_records(conn, csv_data)

        # Закрываем соединение
        conn.close()
        print("✓ Обновление завершено успешно")

    except Exception as e:
        print(f"✗ Ошибка: {e}", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()
