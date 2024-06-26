package db

import (
    "database/sql"
    _ "github.com/lib/pq"
    "log"
)

var DB *sql.DB

func InitDB(dataSourceName string) {
    log.Printf("Connecting to database with DSN: %s", dataSourceName)
    
    var err error
    DB, err = sql.Open("postgres", dataSourceName)
    if err != nil {
        log.Fatalf("The error is: %v", err)
    }

    if err = DB.Ping(); err != nil {
        log.Fatalf("Ping error: %v", err)
    }

    log.Println("Connected to the database successfully.")
}

func GetDB() *sql.DB {
    return DB
}

