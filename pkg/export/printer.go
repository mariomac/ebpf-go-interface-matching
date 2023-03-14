package export

import "fmt"

func Printer(in <-chan string) {
	for m := range in {
		fmt.Println(m)
	}
}
