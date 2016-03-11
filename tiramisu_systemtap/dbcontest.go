package main

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func main() {
	db, err := gorm.Open("postgres", "user=postgres password=12344321 dbname=tiramisu sslmode=disable")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", db.HasTable("tiramisu_state"))
	db.AutoMigrate(&VMstate{})
}
