package main

import (
	"go/token"
	"go/parser"
	"os"
	"log"
	"fmt"
	"go/ast"
	"strings"
	"encoding/json"
	"reflect"
	"net/http"
)

// код писать тут
//go build ./handlers_gen && handlers_gen.exe api.go handlers.go

type Generation struct {
	Name          string //имя апишки
	MethodsParams []MethodGen
}

type MethodGen struct {
	MethodName  string       `json:",omitempty"`
	URL         string       `json:"url"`
	Auth        bool         `json:"auth"`
	Method      string       `json:"method,omitempty"`
	ApiName     string       `json:",omitempty"`
	ParamName   string       `json:",omitempty"`
	FieldParams []FieldParam `json:",omitempty"`
}

type FieldParam struct {
	Name             string
	Type             string
	ValidationParams map[string]interface{}
}

func CheckValidationParams(tags *string, vp map[string]interface{}) {
	if strings.Contains(*tags, "required") {
		vp["required"] = true
	}

}

func main() {
	fset := token.NewFileSet()

	node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := os.Create(os.Args[2])

	//получаем все имена структур апишек
	myApis := make([]Generation, 0)
	for _, f := range node.Decls {
		g, ok := f.(*ast.FuncDecl)
		if !ok {

			continue
		}

		if g.Doc != nil {
			NeedNewApi := true

			strName := g.Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name
			for _, v := range myApis {
				if v.Name == strName {
					NeedNewApi = false
				}

			}
			if !NeedNewApi {
				continue
			}

			if NeedNewApi && strName != "" {
				NewApi := new(Generation)
				NewApi.Name = strName

				myApis = append(myApis, *NewApi)
			}

		}
	}

	fmt.Println(myApis)
	for _, f := range node.Decls {
		g, ok := f.(*ast.FuncDecl)
		if !ok {

			continue
		}

		if g.Doc != nil {
			NeedMethod := false
			ApiName := g.Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name

			for _, v := range myApis {
				if v.Name == ApiName {
					NeedMethod = true

					break
				}
			}
			if !NeedMethod {
				continue
			}
			MethodGen := new(MethodGen)

			MethodGen.ApiName = ApiName
			//получаем метку метода
			mark := g.Doc.List[0].Text

			MyApiGen := strings.Split(mark, "// apigen:api")

			err := json.Unmarshal([]byte(MyApiGen[1]), MethodGen)
			if err != nil {
				fmt.Println("cant unmarshal json")
			}
			MethodGen.MethodName = g.Name.Name
			MethodGen.ParamName = g.Type.Params.List[1].Type.(*ast.Ident).Name

			for i, v := range myApis {
				if v.Name == MethodGen.ApiName {
					myApis[i].MethodsParams = append(myApis[i].MethodsParams, *MethodGen)
				}
			}

		}
	}

	//получение тэгов валидации параметров структур

	for _, f := range node.Decls {
		g, ok := f.(*ast.GenDecl)
		if !ok {

			continue
		}

		for _, spec := range g.Specs {
			currType, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			currStruct, ok := currType.Type.(*ast.StructType)
			if !ok {

				continue
			}
			fmt.Println(currType.Name.Name)
			needStructure := false
			//индексы нужного нам апи и метода

			var indexOfOurStructer int
			var indexOfOurParams int

			//является ли структура нужной, есть ли у нее методы, которые стоит обернуть
			for i, v := range myApis {
				for j, names := range v.MethodsParams {
					if names.ParamName == currType.Name.Name {
						needStructure = true
						indexOfOurStructer = i
						indexOfOurParams = j
					}
				}
			}
			if !needStructure {
				continue
			}

			for _, field := range currStruct.Fields.List {

				myField := new(FieldParam)
				myField.Name = field.Names[0].Name
				myField.Type = field.Type.(*ast.Ident).Name

				//парсинг валидации

				if field.Tag != nil {
					tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
					tags := tag.Get("apivalidator")

					fmt.Println(strings.Split(tags, ","))
					//парсим в мапу
					myField.ValidationParams = make(map[string]interface{})

					if strings.Contains(tags, "required") {
						myField.ValidationParams["required"] = "true"
					} else {
						myField.ValidationParams["required"] = "false"
					}

					if strings.Contains(tags, "paramname") {
						myField.ValidationParams["paramname"] = strings.Split(tags, "=")[1]
					}

					if strings.Contains(tags, "enum") {
						myField.ValidationParams["enum"] = make([]string, 0)
						enums := strings.Split(tags, ",")[0]
						enums = strings.Split(enums, "=")[1]
						simpleEnum := strings.Split(enums, "|")
						for _, v := range simpleEnum {
							myField.ValidationParams["enum"] = append(myField.ValidationParams["enum"].([]string), v)
						}

						if strings.Contains(tags, "default") {
							def := strings.Split(tags, ",")[1]
							def = strings.Split(def, "=")[1]
							myField.ValidationParams["default"] = def
						}
					}
					if strings.Contains(tags, "min=") && !strings.Contains(tags, "max") {
						Values := strings.Split(tags, ",")[1]
						myField.ValidationParams["min"] = strings.Split(Values, "=")[1]
					}

					if strings.Contains(tags, "min") && strings.Contains(tags, "max") {
						Values := strings.Split(tags, ",")
						myField.ValidationParams["min"] = strings.Split(Values[0], "=")[1]
						myField.ValidationParams["max"] = strings.Split(Values[1], "=")[1]
					}

				}
				fmt.Println(myField)
				myApis[indexOfOurStructer].MethodsParams[indexOfOurParams].FieldParams = append(myApis[indexOfOurStructer].MethodsParams[indexOfOurParams].FieldParams, *myField)
			}
		}
	}

	fmt.Println(myApis)

	//кодогенерируем импорты
	fmt.Fprintln(out, `package `+node.Name.Name)
	fmt.Fprintln(out) // empty line
	fmt.Fprintln(out, `import "encoding/json"`)
	fmt.Fprintln(out, `import "net/http"`)
	fmt.Fprintln(out, `import "fmt"`)
	fmt.Fprintln(out, `import "strconv"`)
	fmt.Fprintln(out) // empty line

	// генерация структуры ответа
	fmt.Fprintln(out, " type myResponse struct { \n\tError string      `json:\"error\"` \n	Response  interface{} `json:\"response,omitempty\"`\n}")

	for _, Api := range myApis {

		//создает serveHTTP для каждой структуры
		fmt.Fprint(out, `
func (h *`+ Api.Name+ `) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	switch r.URL.Path {	
	`)
		for _, Method := range Api.MethodsParams {
			fmt.Fprint(out, "\t", `case `, `"`, Method.URL, `"`, ":\n")
			fmt.Fprint(out, "\t\th.handler"+Method.MethodName+"(w,r)")
			fmt.Fprintln(out, )
		}

		fmt.Fprint(out, `
	default:
		MyResp := new(myResponse)
		MyResp.Error = "unknown method"
		mr, err := json.Marshal(MyResp)
		if err != nil {
			fmt.Println("Cant pack json:", err)
		}
		w.WriteHeader(http.StatusNotFound)
		w.Write(mr)
	}
}
`)

		for _, Method := range Api.MethodsParams {
			fmt.Fprint(out, `func (h *`, Api.Name, `) handler`, Method.MethodName, `(w http.ResponseWriter, r *http.Request) {`)
			fmt.Fprintln(out, )
			fmt.Fprintln(out, "\t",
				`   ctx := r.Context()

				   //ответ
				   MyResp := new(myResponse)`)

			fmt.Fprintln(out, "\tvar Params ", Method.ParamName)

			//проверка авторизации
			if Method.Auth == true {
				fmt.Fprintln(out, `if r.Header.Get("X-Auth") != "100500" {
		w.WriteHeader(http.StatusForbidden)
		MyResp.Error = "unauthorized"
		mr, _ := json.Marshal(MyResp)
		w.Write(mr)
		return
	}`)
			}

			fmt.Fprintln(out, `switch r.Method {`, "\n")
			if Method.Method == http.MethodPost {

				//генерация GET

				fmt.Fprintln(out, `	case http.MethodGet:
		w.WriteHeader(http.StatusNotAcceptable)
		MyResp.Error = "bad method"
		mr, _ := json.Marshal(MyResp)
		w.Write(mr)
		return`)
			} else {
				fmt.Fprint(out, `	case http.MethodGet:`, "\n\t\t")
				for _, Param := range Method.FieldParams {
					if Param.Type == "string" {
						if val, ok := Param.ValidationParams["paramname"]; ok {
							fmt.Fprint(out, `Params.`, Param.Name, `=r.URL.Query().Get(`, `"`, strings.ToLower(val.(string)), `")`, "\n")
						} else {
							fmt.Fprint(out, `Params.`, Param.Name, `=r.URL.Query().Get(`, `"`, strings.ToLower(Param.Name), `")`, "\n")
						}
					} else if Param.Type == "int" {
						fmt.Fprintln(out, "var err error")
						if val, ok := Param.ValidationParams["paramname"]; ok {
							fmt.Fprint(out, `Params.`, Param.Name, `, err = strconv.Atoi(r.URL.Query().Get(`, `"`, strings.ToLower(val.(string)), `"))`, "\n")
						} else {
							fmt.Fprint(out, `Params.`, Param.Name, `, err = strconv.Atoi(r.URL.Query().Get(`, `"`, strings.ToLower(Param.Name), `"))`, "\n")
						}
						fmt.Fprint(out, `if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			MyResp.Error = "`, strings.ToLower(Param.Name), ` must be int"
			mr, _ := json.Marshal(MyResp)
			w.Write(mr)
			return

		}`)
					}
				}
			}

			//генерация POST`a
			fmt.Fprintln(out, `   case http.MethodPost:`)
			fmt.Fprintln(out, "\tr.ParseForm()")
			for _, Param := range Method.FieldParams {
				if Param.Type == "string" {
					if val, ok := Param.ValidationParams["paramname"]; ok {
						fmt.Fprint(out, `Params.`, Param.Name, `=r.Form.Get(`, `"`, strings.ToLower(val.(string)), `")`, "\n")
					} else {
						fmt.Fprint(out, `Params.`, Param.Name, `=r.Form.Get(`, `"`, strings.ToLower(Param.Name), `")`, "\n")
					}
				} else if Param.Type == "int" {
					fmt.Fprintln(out, "var err error")
					if val, ok := Param.ValidationParams["paramname"]; ok {
						fmt.Fprint(out, `Params.`, Param.Name, `, err = strconv.Atoi(r.Form.Get(`, `"`, strings.ToLower(val.(string)), `"))`, "\n")
					} else {
						fmt.Fprint(out, `Params.`, Param.Name, `, err = strconv.Atoi(r.Form.Get(`, `"`, strings.ToLower(Param.Name), `"))`, "\n")
					}
					fmt.Fprint(out, `if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			MyResp.Error = "`, strings.ToLower(Param.Name), ` must be int"
			mr, _ := json.Marshal(MyResp)
			w.Write(mr)
			return

		}`)
				}
			}
			fmt.Fprintln(out, `}`)

			//проверка валидности данных
			//проверка required
			for _, p := range Method.FieldParams {
				if p.ValidationParams["required"] == "true" {
					fmt.Fprint(out, `if Params.`, p.Name, ` == "" {
		w.WriteHeader(http.StatusBadRequest)
		MyResp.Error = "`, strings.ToLower(p.Name), ` must me not empty"
		mr, _ := json.Marshal(MyResp)
		w.Write(mr)
		return
	}`)
				}
			}
			fmt.Fprint(out, "\n")

			//проверка на наличие дефолтного значения
			for _, p := range Method.FieldParams {
				if val, ok := p.ValidationParams["default"]; ok {
					fmt.Fprint(out, `if Params.`, p.Name, `==""{`)
					fmt.Fprint(out, `Params.`, p.Name, `="`, val, `"`)
					fmt.Fprint(out, `}`, "\n")

				}
			}

			for _, p := range Method.FieldParams {
				if val, ok := p.ValidationParams["enum"]; ok {
					fmt.Fprint(out, `if `)
					for i, v := range val.([]string) {
						fmt.Fprint(out, `(Params.`, p.Name, `!="`, v, `")  `)
						if i < len(val.([]string))-1 {
							fmt.Fprintf(out, ` && `)
						}

					}
					fmt.Fprint(out, `{
		w.WriteHeader(http.StatusBadRequest)
		MyResp.Error = "`, strings.ToLower(p.Name), ` must be one of [`, val.([]string)[0], ", ", val.([]string)[1], ", ", val.([]string)[2], `]"
		mr, _ := json.Marshal(MyResp)
		w.Write(mr)
		return
	}`)
				}

			}

			fmt.Fprint(out, "\n")
			//проверка min,max
			for _, p := range Method.FieldParams {
				if val, ok := p.ValidationParams["min"]; ok && p.Type == "string" {
					fmt.Fprint(out, `if len(Params.`, p.Name, `) <=`, val, `  {
		w.WriteHeader(http.StatusBadRequest)
		MyResp.Error = "`, strings.ToLower(p.Name), ` len must be >= `, val, `"
		mr, _ := json.Marshal(MyResp)
		w.Write(mr)
		return
	}`)
				}
			}
			fmt.Fprint(out, "\n")

			for _, p := range Method.FieldParams {
				if val, ok := p.ValidationParams["min"]; ok && p.Type == "int" {
					fmt.Fprint(out, `if  Params.`, p.Name, `  < `, val, `  {
		w.WriteHeader(http.StatusBadRequest)
		MyResp.Error = "`, strings.ToLower(p.Name), ` must be >= `, val, `"
		mr, _ := json.Marshal(MyResp)
		w.Write(mr)
		return
	}`)
				}
			}
			fmt.Fprint(out, "\n")

			for _, p := range Method.FieldParams {
				if val, ok := p.ValidationParams["max"]; ok && p.Type == "int" {
					fmt.Fprint(out, `if  Params.`, p.Name, `  >=`, val, `  {
		w.WriteHeader(http.StatusBadRequest)
		MyResp.Error = "`, strings.ToLower(p.Name), ` must be <= `, val, `"
		mr, _ := json.Marshal(MyResp)
		w.Write(mr)
		return
	}`)
				}
			}
			fmt.Fprint(out, "\n")

			//вызываем наш метод
			fmt.Fprint(out, `user,err:=h.`, Method.MethodName, `(ctx,Params)`)
			fmt.Fprint(out, "\n")
			fmt.Fprint(out, "\n")
			//обрабатываем ошибку, если есть
			fmt.Fprint(out, `if err, ok := err.(ApiError); ok {
		w.WriteHeader(err.HTTPStatus)
	}
 
	if  user != nil {
		MyResp.Response = user
	}
	if err != nil {
		MyResp.Error = err.Error()
		if _, ok := err.(ApiError); ok {
			w.WriteHeader(http.StatusNotFound)
			mr, _ := json.Marshal(MyResp)
			w.Write(mr)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}
	mr, err := json.Marshal(MyResp)
	if err != nil {
		fmt.Println("Cant pack json:", err)
	}
	w.Write(mr)`)

			fmt.Fprintln(out, `}`)
		}
	}

}
