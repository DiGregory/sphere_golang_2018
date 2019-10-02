package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные


type TablesContext struct {
	Tables     map[string]Table
	TableNames []string
}
type Table struct {
	Name   string
	Id     string
	Fields []FieldInfo
}

type FieldInfo struct {
	Name     string
	Type     string
	Required bool
	IsKey    bool
}

func (field *FieldInfo) getValueFromString(value string) (interface{}, error) {
	var result interface{}
	var err error
	switch field.Type {
	case "varchar":
		err = nil
		result = value
	case "text":
		err = nil
		result = value
	case "int":
		result, err = strconv.Atoi(value)
	}
	return result, err
}

func (tablesCtxt *TablesContext) containsTable(table string) bool {
	_, ok := tablesCtxt.Tables[table]
	return ok
}

func (field *FieldInfo) getDefaultValue() interface{} {
	switch field.Type {
	case "varchar":
		return ""
	case "text":
		return ""
	case "int":
		return 0
	}
	return nil
}

func (field *FieldInfo) validateField(value interface{}) error {
	if value == nil && field.Required {
		return fmt.Errorf("field %s have invalid type", field.Name)
	}
	switch value.(type) {
	case float64:
		if field.Type != "int" {
			return fmt.Errorf("field %s have invalid type", field.Name)
		}
	case string:
		if field.Type != "varchar" && field.Type != "text" {
			return fmt.Errorf("field %s have invalid type", field.Name)
		}
	}

	return nil
}

func (table *Table) validateInputParameters(params map[string]interface{}, validateKey bool) error {

	for _, field := range table.Fields {
		if value, ok := params[field.Name]; ok {
			if validateKey && field.IsKey {
				return fmt.Errorf("field %s have invalid type", field.Name)
			}
			err := field.validateField(value)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (table *Table) prepareUpdateParameters(params map[string]interface{}) []interface{} {
	result := make([]interface{}, 0)
	for _, v := range params {
		result = append(result, v)
	}
	return result
}

func (table *Table) transformRow(row []interface{}) map[string]interface{} {
	item := make(map[string]interface{}, len(row))
	for i, v := range row {
		switch v.(type) {
		case *sql.NullString:
			if value, ok := v.(*sql.NullString); ok {
				if value.Valid {
					item[table.Fields[i].Name] = value.String
				} else {
					item[table.Fields[i].Name] = nil
				}

			}
		case *sql.NullInt64:
			if value, ok := v.(*sql.NullInt64); ok {
				if value.Valid {
					item[table.Fields[i].Name] = value.Int64
				} else {
					item[table.Fields[i].Name] = nil
				}

			}
		}
	}
	return item
}
func NewDbExplorer(db *sql.DB) (http.Handler, error) {

	tablesContext, err := initContext(db)
	serverMux := http.NewServeMux()
	if err != nil {
		panic(err)
	}
	serverMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case http.MethodGet:
			path := r.URL.Path
			if path == "/" {
				result, err := json.Marshal(map[string]interface{}{"response": map[string]interface{}{"tables": tablesContext.TableNames}})
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.Write(result)
				return
			}
			fragments := strings.Split(path, "/")
			//	fmt.Println(len(fragments ))
			switch len(fragments) {
			case 2:

				// /$table
				tableName := fragments[1]

				if !tablesContext.containsTable(tableName) {
					result, _ := json.Marshal(map[string]interface{}{"error": "unknown table"})
					w.WriteHeader(http.StatusNotFound)
					w.Write(result)
					return
				}

				limit := 5
				offset := 0

				if r.URL.Query().Get("limit") != "" {
					limit, err = strconv.Atoi(r.URL.Query().Get("limit"))
					if err != nil {
						limit = 5
					}
				}
				if r.URL.Query().Get("offset") != "" {
					offset, err = strconv.Atoi(r.URL.Query().Get("offset"))
					if err != nil {
						offset = 0
					}
				}

				fmt.Sprintf("limit %d offset %d\n", limit, offset)
				rows, err := getRows(db, tablesContext.Tables[tableName], limit, offset)

				//rows:=db.QueryRow("SELECT * FROM items LIMIT 5 OFFSET 0" )
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)

					return
				}
				result, err := json.Marshal(
					map[string]interface{}{"response": map[string]interface{}{"records": rows}})
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)

					return
				}
				w.Write(result)
			case 3:
				table := fragments[1]
				id := fragments[2]
				if !tablesContext.containsTable(table) {
					w.WriteHeader(http.StatusNotFound)

					return
				}
				rows, err := getRow(db, tablesContext.Tables[table], id)
				if err != nil {
					w.WriteHeader(http.StatusNotFound)
					result, _ := json.Marshal(map[string]string{"error": "record not found"})
					w.Write(result)
					return
				}
				result, err := json.Marshal(
					map[string]interface{}{"response": map[string]interface{}{"record": rows}})
				w.Write(result)

			}
		case http.MethodDelete:
			path := r.URL.Path
			fragments := strings.Split(path, "/")
			tableName := fragments[1]
			id := fragments[2]
			if !tablesContext.containsTable(tableName) {
				result, _ := json.Marshal(map[string]interface{}{"error": "unknown table"})
				w.WriteHeader(http.StatusNotFound)
				w.Write(result)
				return
			}
			table := tablesContext.Tables[tableName]
			result, err := deleteRow(db, table, id)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)

				return
			}
			resultBytes, _ := json.Marshal(map[string]interface{}{"response": map[string]interface{}{"deleted": result}})
			w.Write(resultBytes)
		case http.MethodPost:
			path := r.URL.Path
			fragments := strings.Split(path, "/")
			tableName := fragments[1]
			id := fragments[2]
			if !tablesContext.containsTable(tableName) {
				result, _ := json.Marshal(map[string]interface{}{"error": "unknown table"})
				w.WriteHeader(http.StatusNotFound)
				w.Write(result)
				return
			}
			table := tablesContext.Tables[tableName]
			decoder := json.NewDecoder(r.Body)
			requestParams := make(map[string]interface{}, len(table.Fields))
			decoder.Decode(&requestParams)
			validationError := table.validateInputParameters(requestParams, true)
			if validationError != nil {
				result, _ := json.Marshal(map[string]interface{}{"error": validationError.Error()})
				w.WriteHeader(http.StatusBadRequest)
				w.Write(result)
				return
			}
			fmt.Printf("Got parameters %#v\n", requestParams)
			table = tablesContext.Tables[tableName]
			result, err := updateRow(db, table, id, requestParams)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)

				return
			}
			resultBytes, _ := json.Marshal(map[string]interface{}{"response": map[string]interface{}{"updated": result}})
			w.Write(resultBytes)
		case http.MethodPut:
			path := r.URL.Path
			fragments := strings.Split(path, "/")
			tableName := fragments[1]
			if !tablesContext.containsTable(tableName) {
				result, _ := json.Marshal(map[string]interface{}{"error": "unknown table"})
				w.WriteHeader(http.StatusNotFound)
				w.Write(result)
				return
			}
			table := tablesContext.Tables[tableName]

			decoder := json.NewDecoder(r.Body)
			requestParams := make(map[string]interface{}, len(table.Fields))
			decoder.Decode(&requestParams)

			result, err := insertRow(db, tablesContext.Tables[tableName], requestParams)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)

				return
			}
			resultBytes, _ := json.Marshal(map[string]interface{}{"response": map[string]interface{}{table.Id: result}})
			w.Write(resultBytes)
		}
	})
	return serverMux, nil
}

func (table *Table) extractParams(values url.Values) map[string]interface{} {
	result := make(map[string]interface{})
	for _, field := range table.Fields {

		if len(values[field.Name]) == 0 {

			result[field.Name] = nil
		} else {
			v, _ := field.getValueFromString(values[field.Name][0])
			result[field.Name] = v
		}
	}
	return result
}

func getRow(db *sql.DB, table Table, id interface{}) (map[string]interface{}, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s = ?", table.Name, table.Id)
	data := table.prepareRow()
	row := db.QueryRow(query, id)
	err := row.Scan(data...)
	if err != nil {
		return nil, err
	}
	return table.transformRow(data), nil
}
func getTables(db *sql.DB) ([]string, error) {

	rows, err := db.Query("SHOW TABLES")

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]string, 0)
	for rows.Next() {
		var tableName string
		rows.Scan(&tableName)
		result = append(result, tableName)

	}
	return result, nil
}

func insertRow(db *sql.DB, table Table, params map[string]interface{}) (int64, error) {
	query := table.prepareInsertQuery()

	queryParams := table.prepareInsert(params)

	res, err := db.Exec(query, queryParams...)
	if err != nil {
		return 0, err
	} else {
		result, _ := res.LastInsertId()
		return result, nil
	}
}

func updateRow(db *sql.DB, table Table, id interface{}, params map[string]interface{}) (int64, error) {
	query := table.prepareUpdateQuery(params)

	queryParams := table.prepareUpdateParameters(params)
	queryParams = append(queryParams, id)

	res, err := db.Exec(query, queryParams...)
	if err != nil {
		return 0, err
	} else {
		result, _ := res.RowsAffected()
		return result, nil
	}
}

func deleteRow(db *sql.DB, table Table, id interface{}) (int64, error) {
	query := table.prepareDeleteQuery()
	res, err := db.Exec(query, id)
	if err != nil {
		return 0, err
	} else {
		result, _ := res.RowsAffected()
		return result, nil
	}
}
func initContext(db *sql.DB) (*TablesContext, error) {
	tables, err := getTables(db)
	if err != nil {
		return nil, err
	}

	result := new(TablesContext)
	result.TableNames = tables
	result.Tables = make(map[string]Table, len(tables))

	for _, table := range tables {
		//Select
		rows, err := db.Query("SELECT column_name, if (column_key='PRI', true, false) as 'key', DATA_TYPE, if(is_nullable='NO', true, false) as nullable from information_schema.columns where  table_name = ? and table_schema=database()", table)
		if err != nil {
			return nil, err
		}
		var keyName string
		fields := make([]FieldInfo, 0)
		for rows.Next() {

			f := new(FieldInfo)
			rows.Scan(&f.Name, &f.IsKey, &f.Type, &f.Required)
			if f.IsKey {
				keyName = f.Name
			}
			fields = append(fields, *f)
		}

		result.Tables[table] = Table{
			Name:   table,
			Id:     keyName,
			Fields: fields,
		}
		rows.Close()
	}
	return result, nil
}

func getRows(db *sql.DB, table Table, limit int, offset int) ([]interface{}, error) {
	rows, err := db.Query(fmt.Sprintf("SELECT * FROM %s LIMIT %d OFFSET %d", table.Name, limit, offset))

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []interface{}{}
	for rows.Next() {

		row := table.prepareRow()
		rows.Scan(row...)
		result = append(result, table.transformRow(row))
	}

	return result, nil
}
func (table *Table) prepareRow() []interface{} {
	row := make([]interface{}, len(table.Fields))
	for i, field := range table.Fields {

		switch field.Type {
		case "varchar":
			row[i] = new(sql.NullString)
		case "text":
			row[i] = new(sql.NullString)
		case "int":
			row[i] = new(sql.NullInt64)

		}

	}
	return row
}

func (table *Table) prepareInsertQuery() string {
	values := make([]string, len(table.Fields))
	placeholders := make([]string, len(table.Fields))
	for i, field := range table.Fields {
		values[i] = field.Name
		placeholders[i] = "?"
	}
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table.Name, strings.Join(values, ", "), strings.Join(placeholders, ", "))
}

func (table *Table) prepareUpdateQuery(params map[string]interface{}) string {
	values := make([]string, 0)
	for k := range params {
		values = append(values, fmt.Sprintf("%s = ?", k))
	}
	return fmt.Sprintf("UPDATE %s SET %s WHERE %s = ?", table.Name, strings.Join(values, ","), table.Id)
}

func (table *Table) prepareDeleteQuery() string {
	return fmt.Sprintf("DELETE FROM %s WHERE %s = ?", table.Name, table.Id)
}
func (table *Table) prepareInsert(params map[string]interface{}) []interface{} {

	result := make([]interface{}, len(table.Fields))
	for i, field := range table.Fields {
		if table.Id == field.Name {
			continue
		}
		if params[field.Name] == nil {
			if !field.Required {
				result[i] = nil
			} else {
				result[i] = field.getDefaultValue()
			}
		} else {
			result[i] = params[field.Name]
		}
	}
	return result
}
