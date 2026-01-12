#!/usr/bin/env python3
"""
Скрипт инициализации БД для Docker
Запускается при первом старте контейнера
"""

import csv
import psycopg2
from psycopg2.extras import execute_values
import time
import sys

# Параметры подключения к PostgreSQL в Docker
DB_PARAMS = {
    'host': 'postgres',
    'port': 5432,
    'database': 'location',
    'user': 'postgres',
    'password': 'postgres'
}

def wait_for_db():
    """Ждем пока PostgreSQL будет готов"""
    print("Ожидание готовности PostgreSQL...")
    for i in range(30):
        try:
            conn = psycopg2.connect(**DB_PARAMS)
            conn.close()
            print("✓ PostgreSQL готов!")
            return True
        except psycopg2.OperationalError:
            time.sleep(1)
    print("✗ PostgreSQL не запустился", file=sys.stderr)
    return False

def load_csv_data():
    """Загружает данные из CSV файлов"""
    print("Загрузка данных из CSV...")

    # Загружаем хрононимы
    data = []
    with open('/app/chrononyms_fin.csv', mode='r', encoding="utf-8-sig") as file:
        reader = csv.DictReader(file)
        for row in reader:
            data.append(dict(row))

    # Загружаем координаты
    coords = []
    with open('/app/districts_coords.csv', mode='r', encoding="utf-8-sig") as f:
        reader_coord = csv.DictReader(f)
        for row in reader_coord:
            coords.append(dict(row))

    # Сопоставляем координаты
    for item in data:
        for coord_item in coords:
            if item["SS"] == coord_item["SS"] and item["District"] == coord_item["Short_dis"]:
                item["Latitude"] = coord_item["Lat"]
                item["Longitude"] = coord_item["Long"]
                break

    print(f"Загружено {len(data)} записей из CSV")
    return data

def insert_all_records(conn, data):
    """Вставляет все записи в базу"""
    records = []
    for item in data:
        lat = float(item.get("Latitude", 0)) if item.get("Latitude") else 0.0
        lon = float(item.get("Longitude", 0)) if item.get("Longitude") else 0.0
        diss = f"{item['District']}, {item['SS']}"

        records.append((
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

    if records:
        print(f"Вставка {len(records)} записей в базу...")
        with conn.cursor() as cur:
            execute_values(
                cur,
                """
                INSERT INTO locations
                (chrononym, definition, context, district, selsovet, latitude, longitude, comment, year, district_ss)
                VALUES %s
                """,
                records
            )
        conn.commit()
        print(f"✓ Успешно добавлено {len(records)} записей")

def main():
    try:
        # Ждем готовности БД
        if not wait_for_db():
            sys.exit(1)

        # Подключаемся
        conn = psycopg2.connect(**DB_PARAMS)

        # Проверяем, есть ли уже данные
        with conn.cursor() as cur:
            cur.execute("SELECT COUNT(*) FROM locations")
            count = cur.fetchone()[0]

        if count > 0:
            print(f"База уже содержит {count} записей, пропускаем инициализацию")
            return

        # Загружаем и вставляем данные
        csv_data = load_csv_data()
        insert_all_records(conn, csv_data)

        conn.close()
        print("✓ Инициализация базы данных завершена")

    except Exception as e:
        print(f"✗ Ошибка: {e}", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()
