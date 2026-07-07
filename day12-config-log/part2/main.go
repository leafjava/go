package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// 读取环境变量
	port := os.Getenv("SERVER_PORT")
	dbHost := os.Getenv("DB_HOST")
	jwtSecret := os.Getenv("JWT_SECRET")
	_ = jwtSecret // 预留，后续使用

	log.Printf("端口: %s, 数据库: %s, JWT密钥: %s\n", port, dbHost, jwtSecret)
}
