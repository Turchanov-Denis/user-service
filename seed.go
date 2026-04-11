package main

import (
	"database/sql"
	"fmt"
	"log"
)

func isUsersTableEmpty(db *sql.DB) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func seedFakeUsers(db *sql.DB, stmtInsert *sql.Stmt, n int) {
	log.Printf("inserting %d fake users\n", n)

	for i := 1; i <= n; i++ {
		id := fmt.Sprintf("%d", i)
		name := fmt.Sprintf("user_%d", i)

		avatar := make([]byte, 512)
		for j := range avatar {
			avatar[j] = byte(i % 256)
		}

		_, err := stmtInsert.Exec(id, name, avatar)
		if err != nil {
			log.Println("[seed] insert error:", err)
		}
	}
}
