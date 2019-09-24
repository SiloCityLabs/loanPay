package main

import (
	"syscall/js"
)

func loadLoansJS() {
	//grab from page
	doc := js.Global().Get("document")
	goLoans = doc.Call("getElementById", "goLoans")
	json := goLoans.Get("value")

	json = json + "test"

	// p, err := json.Marshal(loans)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	goLoans.Set("value", "hello world")
}
