package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

func asyncCrc32(data string) chan string {
	out := make(chan string)
	go func(data string, out chan string) {
		out <- DataSignerCrc32(data)
		close(out)
	}(data, out)
	return out
}

var TH = 6
var ExecutePipeline = func(jobs ...job) {

	wg := &sync.WaitGroup{}
	in := make(chan interface{})

	for _, MyJob := range jobs {

		wg.Add(1)

		out := make(chan interface{})

		go func(job job, in, out chan interface{}) {
			defer wg.Done()
			defer close(out)
			job(in, out)
		}(MyJob, in, out)
		in = out
	}
	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	for i := range in {
		wg.Add(1)
		go func(i interface{}, out chan interface{}, wg *sync.WaitGroup) {
			defer wg.Done()
			data := strconv.Itoa(i.(int))

			mu.Lock()
			md5Data := DataSignerMd5(data)
			mu.Unlock()

			crc32Md5Data := asyncCrc32(md5Data)
			crc32Data := asyncCrc32(data)

			out <- <-crc32Data + "~" + <-crc32Md5Data

		}(i, out, wg)

	}
	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}

	for i := range in {

		wg.Add(1)
		go func(input string, out chan interface{}, wg *sync.WaitGroup) {
			defer wg.Done()

			wgCrc32 := &sync.WaitGroup{}

			crc32Array := make([]string, TH)

			for i := 0; i < TH; i++ {
				wgCrc32.Add(1)
				data := strconv.Itoa(i) + input
				go func(crc32Array []string, data string, i int, wg *sync.WaitGroup) {
					defer wg.Done()

					data = DataSignerCrc32(data)

					crc32Array[i] = data

				}(crc32Array, data, i, wgCrc32)

			}
			wgCrc32.Wait()
			result := strings.Join(crc32Array, "")
			out <- result

		}(i.(string), out, wg)

	}
	wg.Wait()

}

var CombineResults = func(in, out chan interface{}) {

	var array []string

	for i := range in {

		array = append(array, i.(string))

	}
	sort.Strings(array)
	result := strings.Join(array, "_")
	out <- result
}
