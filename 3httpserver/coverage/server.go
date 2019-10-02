package main

import (
	"net/http"
	"fmt"
	"io/ioutil"
	"os"
	"encoding/xml"
	"strconv"

	"encoding/json"
	"strings"
	"sort"
)

type ServerUser struct {
	Id         int    `xml:"id"`
	First_name string `xml:"first_name" json:"-"`
	Last_name  string `xml:"last_name" json:"-"`
	Age        int    `xml:"age"`
	About      string `xml:"about"`
	Gender     string `xml:"gender"`
	Name       string `xml:"-" json:"name"`
	Display    bool   `xml:"-" json:"-"`
}

type rows struct {
	Version string       `xml:"version,attr"`
	Users   []ServerUser `xml:"row"`
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	if (r.Header.Get("AccessToken") != "123456") {
		w.WriteHeader(http.StatusUnauthorized)
		return
	} else {

		var err error
		//получение параметров запроса
		RequestParams := new(SearchRequest)

		RequestParams.Query = r.URL.Query().Get("query")
		RequestParams.Limit, err = strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		RequestParams.Offset, err = strconv.Atoi(r.URL.Query().Get("offset"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)

			return
		}
		RequestParams.OrderBy, err = strconv.Atoi(r.URL.Query().Get("order_by"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		RequestParams.OrderField = r.URL.Query().Get("order_field")

		if (RequestParams.Limit > 25) || (RequestParams.Offset < (-1)) || !(RequestParams.OrderBy == OrderByAsIs || RequestParams.OrderBy == OrderByDesc || RequestParams.OrderBy == OrderByAsc) {
			w.WriteHeader(http.StatusBadRequest)

			return

		}

		fmt.Printf("Were recieved params: %v\n\t", RequestParams)
		//получаем xml

		//"\\3\\99_hw\\coverage\\"
		pwd, _ := os.Getwd() //рабочая директория
		xmldata, _ := ioutil.ReadFile(pwd + "\\dataset.xml")

		//анмаршалим xml в rows
		rows := new(rows)
		_ = xml.Unmarshal(xmldata, &rows)

		for i, v := range rows.Users {
			rows.Users[i].Name = v.First_name + " " + v.Last_name
		}

		//ищем строку запроса в Name и About
		if RequestParams.Query != "" {
			for i, v := range rows.Users {
				if strings.Contains(v.Name, RequestParams.Query) || strings.Contains(v.About, RequestParams.Query) {
					rows.Users[i].Display = true
				}
			}
		} else {
			for i, _ := range rows.Users {
				rows.Users[i].Display = true
			}
		}
		//удаление записей, неудовлетворяющих условию поиска
		for i := len(rows.Users) - 1; i >= 0; i-- {
			users := rows.Users[i]
			if !users.Display {
				rows.Users = append(rows.Users[:i], rows.Users[i+1:]...)
			}
		}

		fmt.Printf("Were found good %v records \n\t", len(rows.Users))

		//сортировка
		if RequestParams.OrderBy == OrderByDesc {
			switch RequestParams.OrderField {
			case "", "Name":

				sort.Slice(rows.Users, func(i, j int) bool {
					return rows.Users[i].Name > rows.Users[j].Name
				})

			case "Id":

				sort.Slice(rows.Users, func(i, j int) bool {
					return rows.Users[i].Id < rows.Users[j].Id
				})

			case "Age":

				sort.Slice(rows.Users, func(i, j int) bool {
					return rows.Users[i].Age < rows.Users[j].Age
				})
			default:
				w.WriteHeader(http.StatusBadRequest)
				jsonResponse, _ := json.Marshal(SearchErrorResponse{Error: "ErrorBadOrderField"})
				w.Write(jsonResponse)
				return

			}
		} else if RequestParams.OrderBy == OrderByAsc {
			switch RequestParams.OrderField {
			case "", "Name":

				sort.Slice(rows.Users, func(i, j int) bool {
					return rows.Users[i].Name < rows.Users[j].Name
				})

			case "Id":

				sort.Slice(rows.Users, func(i, j int) bool {
					return rows.Users[i].Id > rows.Users[j].Id
				})

			case "Age":

				sort.Slice(rows.Users, func(i, j int) bool {
					return rows.Users[i].Age > rows.Users[j].Age
				})

			default:
				w.WriteHeader(http.StatusBadRequest)
				jsonResponse, _ := json.Marshal(SearchErrorResponse{Error: "ErrorBadOrderField"})
				w.Write(jsonResponse)
				return
			}
		}

		//если нам пришло количество записей, которое нам требуется вывести больше, чем этих записей есть, то чтоб не выходить из массива
		if RequestParams.Offset+RequestParams.Limit >= len(rows.Users) {
			RequestParams.Limit = len(rows.Users) - RequestParams.Offset
		}

		fmt.Printf("Were shown %v records \n\t\n\t\n\t", RequestParams.Limit-1)

		JsonUser, _ := json.Marshal(rows.Users[RequestParams.Offset : RequestParams.Offset+RequestParams.Limit])

		w.Write(JsonUser)
		w.WriteHeader(http.StatusOK)
		return
	}
}

//func main() {
//	http.HandleFunc("/", SearchServer)
//	http.ListenAndServe(":8080", nil)
//
//}
