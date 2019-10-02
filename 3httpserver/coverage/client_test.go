package main

import (
	"net/http"

	"testing"
	"net/http/httptest"
	"io"
	"time"
	"net/url"
	"encoding/json"
)

// тут писать код тестов

var badCases = []BadTestCase{
	{
		AccessToken: "123456",
		Request: &BadSearchRequest{
			Limit:      "BadLimit",
			Offset:     "0",
			OrderBy:    "0",
			OrderField: "",
			Query:      "",
		},
		IsError: true,
		Response: &SearchResponse{
			Users:
			nil,
			NextPage: true,
		},
	},
	{
		AccessToken: "123456",
		Request: &BadSearchRequest{
			Limit:      "2",
			Offset:     "BadOffset",
			OrderBy:    "0",
			OrderField: "",
			Query:      "",
		},
		IsError: true,
		Response: &SearchResponse{
			Users:
			nil,
			NextPage: true,
		},
	},
	{
		AccessToken: "123456",
		Request: &BadSearchRequest{
			Limit:      "2",
			Offset:     "0",
			OrderBy:    "BadOrderBy",
			OrderField: "",
			Query:      "",
		},
		IsError: true,
		Response: &SearchResponse{
			Users:
			nil,
			NextPage: true,
		},
	},
}

var cases = []TestCase{
	//good request
	{
		AccessToken: "123456",
		Request: &SearchRequest{
			Limit:      3,
			Offset:     0,
			OrderBy:    0,
			OrderField: "",
			Query:      "",
		},
		IsError: false,
		Response: &SearchResponse{
			Users:
			[]User{
				{Id: 0,}, {Id: 1,}, {Id: 2,},
			},
			NextPage: true,
		},
	},
	{
		//request with badtoken
		AccessToken: "badToken",
		Request: &SearchRequest{
			Limit:      3,
			Offset:     0,
			OrderBy:    0,
			OrderField: "",
			Query:      "",
		},

		IsError: true,
		Response: &SearchResponse{
			Users: nil,

			NextPage: true,
		},
	},
	//bad request with big limit
	{
		AccessToken: "123456",
		Request: &SearchRequest{
			Limit:      35,
			Offset:     0,
			OrderBy:    0,
			OrderField: "",
			Query:      "",
		},
		IsError: true,
		Response: &SearchResponse{
			Users: nil,

			NextPage: true,
		},
	},
	//bad request with  offset<0
	{
		AccessToken: "123456",
		Request: &SearchRequest{
			Limit:      10,
			Offset:     -1,
			OrderBy:    0,
			OrderField: "",
			Query:      "",
		},
		IsError: true,
		Response: &SearchResponse{
			Users: nil,

			NextPage: true,
		},
	},
	//bad limit
	{
		AccessToken: "123456",
		Request: &SearchRequest{
			Limit:      -1,
			Offset:     0,
			OrderBy:    0,
			OrderField: "",
			Query:      "",
		},
		IsError: true,
		Response: &SearchResponse{
			Users: nil,

			NextPage: true,
		},
	},
	//bad request with bad orderby
	{
		AccessToken: "123456",
		Request: &SearchRequest{
			Limit:      3,
			Offset:     0,
			OrderBy:    -2,
			OrderField: "",
			Query:      "",
		},
		IsError: true,
		Response: &SearchResponse{
			Users: nil,

			NextPage: true,
		},
	},
	//request with bad orderfield
	{
		AccessToken: "123456",
		Request: &SearchRequest{
			Limit:      3,
			Offset:     0,
			OrderBy:    1,
			OrderField: "gender",
			Query:      "",
		},
		IsError: true,
		Response: &SearchResponse{
			Users: nil,

			NextPage: true,
		},
	},

	//good request with Ordering by Name
	{
		AccessToken: "123456",
		Request: &SearchRequest{
			Limit:      3,
			Offset:     0,
			OrderBy:    1,
			OrderField: "",
			Query:      "",
		},
		IsError: false,
		Response: &SearchResponse{
			Users: []User{
				User{Id: 13,}, User{Id: 33,}, User{Id: 18},},
			NextPage: true,
		},
	},
	//bad orderfield by asc
	{
		AccessToken: "123456",
		Request: &SearchRequest{
			Limit:      3,
			Offset:     0,
			OrderBy:    -1,
			OrderField: "gender",
			Query:      "",
		},
		IsError: true,
		Response: &SearchResponse{
			Users:
			nil,
			NextPage: true,
		},
	},
	{
		AccessToken: "123456",
		Request: &SearchRequest{
			Limit:      3,
			Offset:     0,
			OrderBy:    1,
			OrderField: "",
			Query:      "",
		},
		IsError: false,
		Response: &SearchResponse{
			Users: []User{
				User{Id: 13,}, User{Id: 33,}, User{Id: 18},},
			NextPage: true,
		},
	},
	//good request with reverse ordering by Name
	{
		AccessToken: "123456",
		Request: &SearchRequest{
			Limit:      3,
			Offset:     0,
			OrderBy:    -1,
			OrderField: "",
			Query:      "",
		},
		IsError: false,
		Response: &SearchResponse{
			Users: []User{
				User{Id: 15,}, User{Id: 16,}, User{Id: 19},},
			NextPage: true,
		},
	},
	{
		AccessToken: "123456",
		Request: &SearchRequest{
			Limit:      3,
			Offset:     0,
			OrderBy:    -1,
			OrderField: "",
			Query:      "",
		},
		IsError: false,
		Response: &SearchResponse{
			Users: []User{
				User{Id: 15,}, User{Id: 16,}, User{Id: 19},},
			NextPage: true,
		},
	},

	//good request with reverse ordering by id
	{
		AccessToken: "123456",
		Request: &SearchRequest{
			Limit:      3,
			Offset:     0,
			OrderBy:    -1,
			OrderField: "Id",
			Query:      "",
		},
		IsError: false,
		Response: &SearchResponse{
			Users: []User{
				User{Id: 34,}, User{Id: 33,}, User{Id: 32},},
			NextPage: true,
		},
	},
	{
		AccessToken: "123456",
		Request: &SearchRequest{
			Limit:      3,
			Offset:     0,
			OrderBy:    1,
			OrderField: "Id",
			Query:      "",
		},
		IsError: false,
		Response: &SearchResponse{
			Users: []User{
				User{Id: 0,}, User{Id: 1,}, User{Id: 2},},
			NextPage: true,
		},
	},
	//good request with reverse ordering by Age
	{
		AccessToken: "123456",
		Request: &SearchRequest{
			Limit:      3,
			Offset:     0,
			OrderBy:    -1,
			OrderField: "Age",
			Query:      "",
		},
		IsError: false,
		Response: &SearchResponse{
			Users: []User{
				User{Id: 32,}, User{Id: 13,}, User{Id: 6},},
			NextPage: true,
		},
	},
	//good request with  ordering by Age
	{
		AccessToken: "123456",
		Request: &SearchRequest{
			Limit:      3,
			Offset:     0,
			OrderBy:    1,
			OrderField: "Age",
			Query:      "",
		},
		IsError: false,
		Response: &SearchResponse{
			Users: []User{
				User{Id: 1,}, User{Id: 15,}, User{Id: 23},},
			NextPage: true,
		},
	},
	//good request with query
	{
		AccessToken: "123456",
		Request: &SearchRequest{
			Limit:      3,
			Offset:     0,
			OrderBy:    0,
			OrderField: "",
			Query:      "V",
		},
		IsError: false,
		Response: &SearchResponse{
			Users: []User{
				User{Id: 2,}, User{Id: 6,}, User{Id: 9},},
			NextPage: true,
		},
	},
	{
		AccessToken: "123456",
		Request: &SearchRequest{
			Limit:      3,
			Offset:     0,
			OrderBy:    0,
			OrderField: "",
			Query:      "Nulla cillum enim voluptate",
		},
		IsError: false,
		Response: &SearchResponse{
			Users: []User{
				User{Id: 0,},},
			NextPage: false,
		},
	},
	{
		AccessToken: "TimeOut",
		Request: &SearchRequest{
			Limit:      3,
			Offset:     0,
			OrderBy:    0,
			OrderField: "",
			Query:      " ",
		},
		IsError: true,
		Response: &SearchResponse{
			Users: []User{
				User{Id: 0,},},
			NextPage: false,
		},
	},
}

type TestCase struct {
	Request     *SearchRequest
	Response    *SearchResponse
	IsError     bool
	AccessToken string
}
type BadTestCase struct {
	Request     *BadSearchRequest
	Response    *SearchResponse
	IsError     bool
	AccessToken string
}
type BadSearchRequest struct {
	Limit      interface{}
	Offset     interface{}
	Query      interface{}
	OrderField interface{}

	OrderBy interface{}
}

func compareResult(expectedUsers []User, gotUsers []User) bool {
	if len(expectedUsers) != len(gotUsers) {
		return false
	}
	result := true
	for i, _ := range expectedUsers {
		if expectedUsers[i].Id != gotUsers[i].Id {
			result = false
		}

	}
	return result
}

func funcIsError(t *testing.T, res *SearchResponse, err error, caseIter int, i TestCase) {

	if err != nil && !i.IsError {
		t.Errorf("[%d] unexpected error: %#v", caseIter, err)
	}
	if err == nil && i.IsError {
		t.Errorf("[%d] expected error, got nil", caseIter)
	}
	if err == nil && !compareResult(i.Response.Users, res.Users) {
		t.Errorf("\n[%d] wrong result, expected \n\t%#v, got \n\t%#v", caseIter, i.Response, res.Users)
	}

	if err == nil && i.Response.NextPage != res.NextPage {
		t.Errorf("\n[%d] Incorrect NextPage \n\t%#v, got \n\t%#v", caseIter, i.Response, res)
	}
}

func TestClientServer(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	for CaseI, Case := range cases {
		sc := &SearchClient{}
		sc.AccessToken = Case.AccessToken
		sc.URL = ts.URL

		res, err := sc.FindUsers(*Case.Request)
		funcIsError(t, res, err, CaseI, Case)
	}
	ts.Close()

}
func TestSearchServer(t *testing.T) {
	for _, Case := range badCases {
		searcherParams := url.Values{}
		searcherParams.Add("limit", Case.Request.Limit.(string))
		searcherParams.Add("offset", Case.Request.Offset.(string))
		searcherParams.Add("query", Case.Request.Query.(string))
		searcherParams.Add("order_field", Case.Request.OrderField.(string))
		searcherParams.Add("order_by", Case.Request.OrderBy.(string))

		url := "http://kek.ru:8080" + "?" + searcherParams.Encode()

		req := httptest.NewRequest("GET", url, nil)
		req.Header.Add("AccessToken", "123456")
		w := httptest.NewRecorder()

		SearchServer(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Not Bad request")
		}
	}

}

func SearchServerJsonFail(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, `"err": "bad json"}`)
}

func TestSearchServerJsonFail(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerJsonFail))
	searchClient := &SearchClient{
		URL: ts.URL,
	}
	_, err := searchClient.FindUsers(SearchRequest{})

	if err.Error() != `cant unpack result json: invalid character ':' after top-level value` {
		t.Error("Bad json test :(")
	}
	ts.Close()
}
func SearchServerUnknown(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	jsonResponse, _ := json.Marshal(SearchErrorResponse{Error: "Unknown error"})
	w.Write(jsonResponse)
}

func TestBadRequestError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerUnknown))
	searchClient := &SearchClient{URL: ts.URL}
	_, err := searchClient.FindUsers(SearchRequest{})
	if err == nil {
		t.Error("TestBadRequestError is not found")
	}

	ts.Close()
}

func SearchServerTimeoutError(w http.ResponseWriter, r *http.Request) {
	time.Sleep(time.Second * 2)
	w.WriteHeader(http.StatusOK)
}
func TestTimeoutError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerTimeoutError))
	searchClient := &SearchClient{
		URL: ts.URL,
	}

	_, err := searchClient.FindUsers(SearchRequest{})

	if err == nil {
		t.Error("Timeout check error")
	}

	ts.Close()
}
func SearchServerUnknownError(w http.ResponseWriter, r *http.Request) {
}

func TestUnknownError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerUnknownError))
	searchClient := &SearchClient{
		URL: "bad_link",
	}

	_, err := searchClient.FindUsers(SearchRequest{})

	if err == nil {
		t.Error("TestUnknownError ")
	}

	ts.Close()
}

func SearchServerInternalServerError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}
func TestStatusInternalServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServerInternalServerError))
	searchClient := &SearchClient{URL: ts.URL}
	_, err := searchClient.FindUsers(SearchRequest{})

	if err.Error() != "SearchServer fatal error" {
		t.Error("SearchServer fatal error")
	}

	ts.Close()
}
