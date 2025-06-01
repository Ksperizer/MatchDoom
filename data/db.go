package data 
import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql" 
)

var DB *sql.DB

func InitDB(){
	dsn := "root:root@tcp(127.0.0.1:3306)/matchdoom"
	var err error 
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	if err = DB.Ping(); err != nil {
		panic(err)
	}

	fmt.Println("Connexion Mysql reussite")
}