package main

import (
	"alchemicallabaratory/handlers"
	"alchemicallabaratory/repository"
	"alchemicallabaratory/server"
	"alchemicallabaratory/workers/boiler"
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	db := connectToDB()

	GrimoireRepository := repository.NewGrimoireRepository(db)

	BoiledWorker := boiler.NewBoilerWorker(GrimoireRepository)

	GuildHandler := handlers.NewGuildHandler(GrimoireRepository, BoiledWorker)

	server := server.NewServer(GuildHandler)

	err := server.Run()
	if err != nil {
		log.Fatalf("Server error:%v", err)
	}
}

func connectToDB() *sqlx.DB {

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	fmt.Println(dbHost)

	dns := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sqlx.Open("postgres", dns)

	log.Println("Connect to DB")

	if err != nil {
		log.Fatalf("We can't connect to DB")
	}
	return db
}
