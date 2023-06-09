package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

var (
	DB_CONN *sql.DB
)

func open_connection() {
	dbHost := "localhost"
	dbUser := "root"
	dbPass := "root123##"
	dbName := "d"
	dbPort := "3306"

	connString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?tls=skip-verify&parseTime=true", dbUser, dbPass, dbHost, dbPort, dbName)

	var err error
	DB_CONN, err = sql.Open("mysql", connString)
	if err != nil {
		fmt.Println("Error: Could not open database connection")
		os.Exit(1)
	}
}

// Select selects rows from a table based on the provided columns and WHERE clause.
func Select(db *sql.DB, tableName string, columns []string, whereClause map[string]interface{}) ([]map[string]interface{}, error) {
	query := "SELECT " + strings.Join(columns, ", ") + " FROM " + tableName

	// Prepare the WHERE clause if it exists
	var whereValues []interface{}
	if len(whereClause) > 0 {
		whereConditions := []string{}
		for key, value := range whereClause {
			whereConditions = append(whereConditions, fmt.Sprintf("%s = ?", key))
			whereValues = append(whereValues, value)
		}
		query += " WHERE " + strings.Join(whereConditions, " AND ")
	}

	rows, err := db.Query(query, whereValues...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columnNames, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	result := []map[string]interface{}{}

	for rows.Next() {
		columnPointers := make([]interface{}, len(columnNames))
		columnValues := make([]interface{}, len(columnNames))

		for i := range columnValues {
			columnPointers[i] = &columnValues[i]
		}

		err := rows.Scan(columnPointers...)
		if err != nil {
			return nil, err
		}

		rowData := make(map[string]interface{})
		for i, name := range columnNames {
			switch v := columnValues[i].(type) {
			case []byte:
				rowData[name] = string(v)
			default:
				rowData[name] = v
			}
		}

		result = append(result, rowData)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// Insert inserts multiple rows into a table.
func Insert(db *sql.DB, tableName string, data []map[string]interface{}) error {
	if len(data) == 0 {
		return nil // Nothing to insert
	}

	columns := make([]string, 0, len(data[0]))
	for key := range data[0] {
		columns = append(columns, key)
	}

	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	var values []interface{}
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES", tableName, strings.Join(columns, ", "))

	rowsValues := make([]string, 0, len(data))
	for _, row := range data {
		rowValues := make([]string, len(columns))
		for i, col := range columns {
			values = append(values, row[col])
			rowValues[i] = "?"
		}
		rowsValues = append(rowsValues, fmt.Sprintf("(%s)", strings.Join(rowValues, ", ")))
	}

	query += strings.Join(rowsValues, ", ")

	fmt.Printf("Query: %s\n", query)
	fmt.Printf("Values: %+v\n", values)

	_, err := db.Exec(query, values...)
	if err != nil {
		return err
	}

	return nil
}

// Update updates multiple rows in a table based on the provided data and WHERE conditions.
func Update(db *sql.DB, table string, data map[string]interface{}, where []map[string]interface{}) error {
	query := "UPDATE %s SET "

	keys := []string{}
	values := []interface{}{}
	for key, value := range data {
		keys = append(keys, fmt.Sprintf("%s = ?", key))
		values = append(values, value)
	}
	query = fmt.Sprintf(query+strings.Join(keys, ", "), table)

	whereConditions := []string{}
	for _, condition := range where {
		for key, value := range condition {
			whereConditions = append(whereConditions, fmt.Sprintf("%s = ?", key))
			values = append(values, value)
		}
	}
	query += " WHERE " + strings.Join(whereConditions, " AND ")

	fmt.Printf("Query: %s\n", query)
	fmt.Printf("Values: %+v\n", values)

	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(values...)
	return err
}

func main() {
	open_connection()

	tableName := "student"

	// Inserting Rows
	InsertData := []map[string]interface{}{
		{"id": 1, "name": "Alice", "age": "25", "email": "alice@example.com"},
		{"id": 2, "name": "Bob", "age": "24", "email": "bob@example.com"},
	}
	err := Insert(DB_CONN, tableName, InsertData)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	SelectColumns := []string{"id", "name", "age", "email"}
	// Add columns and their corresponding values to the whereClause map when you want to include a WHERE clause
	// in the SELECT query. Keep the whereClause map empty when you don't want to include a WHERE clause.
	whereClause := map[string]interface{}{
		"id":   1,
		"name": "Alice",
	}

	results, err := Select(DB_CONN, tableName, SelectColumns, whereClause)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Process the query results
	for _, row := range results {
		fmt.Println(row)
	}

	UpdateColumn := map[string]interface{}{
		"age": "1",
	}

	UpdateWhere := []map[string]interface{}{
		{"id": 1},
		{"name": "Alice"},
	}

	err = Update(DB_CONN, tableName, UpdateColumn, UpdateWhere)
	if err != nil {
		fmt.Printf("\nUpdate Error: %v", err.Error())
	} else {
		fmt.Printf("\nUpdated.")
	}
}
