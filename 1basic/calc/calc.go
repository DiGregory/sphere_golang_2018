package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"errors"
)

var ErrBadInput = errors.New("bad input")

func SolveRPN(line string) (int, error) {
	stack := []int{}
	input := strings.Split(line[:len(line)-1], " ")
	fmt.Println(line)
	for _, v := range input {
		if v == "*" {
			stack = append(stack[:len(stack)-2], stack[len(stack)-2]*stack[len(stack)-1])
		} else if v == "+" {
			stack = append(stack[:len(stack)-2], stack[len(stack)-2]+stack[len(stack)-1])
		} else if v == "-" {
			stack = append(stack[:len(stack)-2], stack[len(stack)-2]-stack[len(stack)-1])
		} else if v == "/" {
			if stack[len(stack)-1] == 0 {
				return 0, ErrBadInput
			}
			stack = append(stack[:len(stack)-2], stack[len(stack)-2]/stack[len(stack)-1])

		} else {
			value, _ := strconv.Atoi(v)
			stack = append(stack, value)
		}
	}

	if len(stack) > 2 {

		return 0, ErrBadInput
	}
	return stack[len(stack)-1], nil
}

func main() {
	fmt.Println("Your calculation: ")
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	fmt.Println(SolveRPN(line))

}
