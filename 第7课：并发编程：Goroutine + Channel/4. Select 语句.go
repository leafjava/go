package main

import "fmt"

//func main() {
//	ch1 := make(chan string)
//	ch2 := make(chan string)
//
//	go func() {
//		time.Sleep(100 * time.Millisecond)
//		ch1 <- "来自 ch1"
//	}()
//
//	go func() {
//		time.Sleep(200 * time.Millisecond)
//		ch2 <- "来自 ch2"
//	}()
//
//	for i := 0; i < 2; i++ {
//		select {
//		case msg1 := <-ch1:
//			fmt.Println(msg1)
//		case msg2 := <-ch2:
//			fmt.Println(msg2)
//		}
//	}
//}

func main() {
	ch1 := make(chan string)
	ch2 := make(chan string)

	go func() {
		ch1 <- "5"
	}()

	go func() {
		ch2 <- "10"
	}()

	for i := 0; i < 2; i++ {
		select {
		case msg1 := <-ch1:
			fmt.Println(msg1)
		case msg2 := <-ch2:
			fmt.Println(msg2)
		}
	}
}
