package main

import (
	"booker/config"
	"booker/storage/database"
	"fmt"
)

func main() {
	cfg, err := config.LoadConfig(".env")
	if err != nil {
		panic(err)
	}
	fmt.Println(cfg)
	dbConfig := database.DbConfig{
		DbName:     cfg.DbName,
		DbUser:     cfg.DbUser,
		DbPassword: cfg.DbPassword,
		DbHost:     cfg.DbHost,
		DbPort:     cfg.DbPort,
	}
	db, err := database.DbInit(&dbConfig)
	if err != nil {
		panic(err)
	}
	err = database.CreateTables(db)
	if err != nil {
		panic(err)
	}
}
