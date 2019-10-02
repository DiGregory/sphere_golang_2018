package main

import (
	"testing"
	  )


func TestSolveRPN(t *testing.T) {
	 type expectedData struct{
	 	value interface{}
	 	error

	 }
	var cases = []struct {
		expected expectedData
		input    string
	}{
		{
			expected: expectedData{3,nil},
			input: "1 2 + ",
		},
		{
			expected: expectedData{9,nil},
			input: "3 3 * ",
		},
		{
			expected: expectedData{0,ErrBadInput},
			input: "2 2 3 + 3 4 + ",

		},
		{
			expected: expectedData{-1,nil},
			input: "2 3 - ",
		},{
			expected: expectedData{4/3,nil},
			input: "4 3 / ",
		},
		{
			expected: expectedData{0,ErrBadInput},
			input: "4 0 / ",
		},




	}
	for _, item := range cases {
		result,err := SolveRPN(item.input)
		if err!=item.expected.error{
			t.Error()
		}
		if result!=item.expected.value {

			t.Error("expected", item.expected, "have", result)

		}
	}
}