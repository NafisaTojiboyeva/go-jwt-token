package main

import (
	"fmt"
)

var (
	dbHost = "localhost"
	dbUser = "nafisa"
	dbPassword = "1517"
	dbName = "app1"
	dbPort = 5432
)

var DB_CONFIG = fmt.Sprintf(
	"host=%s user=%s password=%s dbname=%s port=%d",
	dbHost, dbUser, dbPassword, dbName, dbPort,
)

