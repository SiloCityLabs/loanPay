// In this example we'll look at how to implement
// a _worker pool_ using goroutines and channels.

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"sync"

	"gopkg.in/yaml.v2"
)

type loans struct {
	Extra float32 `yaml:"extra"`
	Loans []struct {
		Apy     float32 `yaml:"apy"`
		Balance float32 `yaml:"balance"`
		Min     float32 `yaml:"min"`
		//Name    string  `yaml:"name"` // Dont need this right now yet
	} `yaml:"loans"`
	FastestMethod  loanResult
	CheapestMethod loanResult
}

type loanResult struct {
	Order     []int
	Months    int16
	TotalPaid float32
}

var loanResults loans
var canPay float32
var jobs chan loanResult

func main() {

	var waitgroup sync.WaitGroup
	jobs = make(chan loanResult, 100)

	//Load the loans.yaml file
	loadFile()

	// This starts up 8 workers, initially blocked
	// because there are no jobs yet.
	for w := 1; w <= 8; w++ {
		waitgroup.Add(1)
		go worker(w, &waitgroup)
	}

	// Here we send `jobs` and then `close` that
	// channel to indicate that's all the work we have.
	//Generate possible combinations and add them to queue
	permutation(rangeSlice(0, len(loanResults.Loans)))
	close(jobs)

	// Finally we collect all the results of the work.
	// This also ensures that the worker goroutines have
	// finished.
	waitgroup.Wait()

	fmt.Println(loanResults)
}

func loadFile() {
	file, err := os.Open("loans.yaml")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	b, err := ioutil.ReadAll(file)

	err2 := yaml.Unmarshal([]byte(b), &loanResults)
	if err2 != nil {
		log.Fatal(err2)
	}
}

func worker(id int, waitgroup *sync.WaitGroup) {
	for j := range jobs {
		//fmt.Printf("Started job %v\n", j)
		processLoanOrder(j)
	}

	waitgroup.Done()
}

func rangeSlice(start, stop int) []int {
	if start > stop {
		panic("Slice ends before it started")
	}
	xs := make([]int, stop-start)
	for i := 0; i < len(xs); i++ {
		xs[i] = i + start
	}
	return xs
}

func permutation(xs []int) {

	var rc func([]int, int)
	rc = func(a []int, k int) {
		if k == len(a) {
			loanorder := loanResult{}
			loanorder.Order = append([]int{}, a...) // Important to keep order
			jobs <- loanorder
		} else {
			for i := k; i < len(xs); i++ {
				a[k], a[i] = a[i], a[k]
				rc(a, k+1)
				a[k], a[i] = a[i], a[k]
			}
		}
	}
	rc(xs, 0)
}

func processLoanOrder(loan loanResult) {
	var balances []float32
	canPayExtra := loanResults.Extra

	//Insert balances
	for _, l := range loan.Order {
		balances = append(balances, loanResults.Loans[l].Balance)
	}

	//fmt.Print("Balances: ")
	//fmt.Println(balances)

	for {
		//reset the monthly extra payment counter
		canPayMonth := canPayExtra

		// fmt.Printf("Balances after %v month(s): ", loan.Months)
		// fmt.Println(balances)

		for _, l := range loan.Order {
			if balances[l] == 0 {
				canPayMonth += loanResults.Loans[l].Min //Rollover method
				//Complete loan
				continue
			}

			//Make the minimum payment
			balances[l] -= loanResults.Loans[l].Min
			loan.TotalPaid += loanResults.Loans[l].Min

			// check if balance is overpaid
			if balances[l] < 0 {
				//add this to canpay extra
				overpaid := (balances[l] * -1)
				canPayMonth += overpaid
				loan.TotalPaid -= overpaid
				balances[l] = 0
			}
		}

		// fmt.Printf("Balances after first payment: ")
		// fmt.Println(balances)

		//Pay each loan extra in order until we are out of money
		for canPayMonth != 0 {
			//Lets quickly see if they are all zero to break out
			extraCanDoMore := false
			for _, bal := range balances {
				if bal != 0 {
					extraCanDoMore = true //lets keep paying extra then
					break
				}
			}
			if extraCanDoMore == false {
				loan.TotalPaid -= canPayMonth //lets not count that against our numbers
				break
			}

			for _, l := range loan.Order {
				if balances[l] == 0 {
					//Complete loan
					continue
				}

				//Pay whatever extra we can
				balances[l] -= canPayMonth
				loan.TotalPaid += canPayMonth
				canPayMonth = 0

				// check if balance is overpaid
				if balances[l] < 0 {
					//readd this to canpay
					overpaid := (balances[l] * -1)
					canPayMonth += overpaid
					loan.TotalPaid -= overpaid
					balances[l] = 0
				}
			}
		}

		// fmt.Printf("Balances after extra: ")
		// fmt.Println(balances)

		//Calculate interest for each loan
		for _, l := range loan.Order {
			if balances[l] == 0 {
				//Complete loan
				continue
			}

			interest := balances[l] * loanResults.Loans[l].Apy / 12
			interest = float32(math.RoundToEven(float64(interest)*100) / 100) // Bank round?
			balances[l] += interest
		}

		// fmt.Printf("Balances after interest: ")
		// fmt.Println(balances)

		// One month has elapsed
		loan.Months++

		//If all balances are empty, we are done
		var totalBalances float32
		for _, bal := range balances {
			totalBalances += bal
		}
		//fmt.Println(totalBalances)
		if totalBalances == 0 {
			break // We are done!!
		}

		//Impossible payment plan, 50 years +
		if loan.Months >= 600 {
			return
		}
	}

	if loanResults.FastestMethod.Months == 0 {
		loanResults.FastestMethod = loan
		loanResults.CheapestMethod = loan
	}

	//Was it the fastest then cheapest
	if loanResults.FastestMethod.Months >= loan.Months && loanResults.FastestMethod.TotalPaid > loan.TotalPaid {
		loanResults.FastestMethod = loan
	}

	//was it cheapest then fastest
	if loanResults.CheapestMethod.TotalPaid >= loan.TotalPaid && loanResults.CheapestMethod.Months > loan.Months {
		loanResults.CheapestMethod = loan
	}
}
