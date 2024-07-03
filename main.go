package main

import (
  "database/sql"
  "fmt"
  "log"

  _ "github.com/lib/pq"
)

func main() {
  connStr := "postgresql://inventory-sys_owner:xedTjh9BD2iZ@ep-tiny-mouse-a5awt0fq.us-east-2.aws.neon.tech/inventory-sys?sslmode=require"
  db, err := sql.Open("postgres", connStr)
  if err != nil {
    log.Fatal(err)
  }
  defer db.Close()

  rows, err := db.Query("select version()")
  if err != nil {
    log.Fatal(err)
  }
  defer rows.Close()

  var version string
  for rows.Next() {
    err := rows.Scan(&version)
    if err != nil {
      log.Fatal(err)
    }
  }
  fmt.Printf("version=%s\n", version)
}