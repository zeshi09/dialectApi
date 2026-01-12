package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/zeshi09/dialectApi/ent"
	_ "github.com/lib/pq"
)

func connectDB(ctx context.Context) *ent.Client {
	// Получаем параметры подключения из переменных окружения или используем значения по умолчанию
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "location")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	client, err := ent.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("failed opening connection to postgres: %v", err)
	}

	// Проверка подключения
	if err := client.Schema.Create(ctx); err != nil {
		log.Printf("warning: schema creation: %v", err)
	}

	fmt.Println("successful db connecting")
	return client
}

// getEnv возвращает значение переменной окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
