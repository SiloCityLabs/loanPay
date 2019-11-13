package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

type loansStruct struct {
	Extra float32 `yaml:"extra"`
	Loans []struct {
		Apy     float32 `yaml:"apy"`
		Balance float32 `yaml:"balance"`
		Min     float32 `yaml:"min"`
		//Name    string  `yaml:"name"` // Dont need this right now yet
	} `yaml:"loans"`
}

type result struct {
	Order     []int
	Months    int16
	TotalPaid float32
}

var loans loansStruct

var jobLoan chan result
var jobResults chan result
var fastestResult result
var cheapestResult result

var threads = 8

func main() {
	start := time.Now()

	var wgp sync.WaitGroup // Permutation
	var wgr sync.WaitGroup // Result
	jobLoan = make(chan result, 1000)
	jobResults = make(chan result, 100)

	//Load the loans.yaml file
	loadFile()

	// This starts up 8 workers, initially blocked
	// because there are no jobs yet.
	for w := 1; w <= threads; w++ {
		wgp.Add(1)
		go worker(w, &wgp)
	}

	//Start processing the results while we wait
	go comparator(&wgr)

	// Here we send `jobs` and then `close` that
	// channel to indicate that's all the work we have.
	//Generate possible combinations and add them to queue
	permutation(rangeSlice(0, len(loans.Loans)), &wgr)

	// Finally we collect all the results of the work.
	// This also ensures that the worker goroutines have
	// finished.
	wgp.Wait()
	wgr.Wait()
	close(jobResults)

	fmt.Println(fastestResult)
	fmt.Println(cheapestResult)

	fmt.Println(time.Since(start))
}

func loadFile() {
	file, err := os.Open("loans.yaml")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	b, err := ioutil.ReadAll(file)

	err2 := yaml.Unmarshal([]byte(b), &loans)
	if err2 != nil {
		log.Fatal(err2)
	}
}

func worker(id int, waitgroup *sync.WaitGroup) {
	for j := range jobLoan {
		//fmt.Printf("Started job %v\n", j)
		processLoanOrder(j)
	}

	waitgroup.Done()
}

func comparator(waitgroup *sync.WaitGroup) {
	for loan := range jobResults {
		if fastestResult.Months == 0 {
			fastestResult = loan
			cheapestResult = loan
			fmt.Printf("Winner 0 (by default): %v\n", loan)
		}

		//eliminate emmediately, shaves about 0.5% of time
		if (loan.Months > fastestResult.Months) || (loan.TotalPaid > fastestResult.TotalPaid) {
			waitgroup.Done()
			continue
		}

		if fastestResult.Months >= loan.Months && fastestResult.TotalPaid > loan.TotalPaid {
			fastestResult = loan
			fmt.Printf("Replacement Fastest Winner: %v\n", loan)
		}

		if cheapestResult.TotalPaid > loan.TotalPaid && cheapestResult.Months >= loan.Months {
			cheapestResult = loan
			fmt.Printf("Replacement Cheapest Winner: %v\n", loan)
		}

		waitgroup.Done()
	}
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

func permutation(xs []int, waitgroup *sync.WaitGroup) {

	var rc func([]int, int)
	rc = func(a []int, k int) {
		if k == len(a) {
			// append is important to keep order of array
			waitgroup.Add(1)
			jobLoan <- result{Order: append([]int{}, a...)}
		} else {
			for i := k; i < len(xs); i++ {
				a[k], a[i] = a[i], a[k]
				rc(a, k+1)
				a[k], a[i] = a[i], a[k]
			}
		}
	}
	rc(xs, 0)

	close(jobLoan)
}

func processLoanOrder(loan result) {
	var balances []float32
	canPayExtra := loans.Extra

	//Insert balances
	for _, l := range loan.Order {
		balances = append(balances, loans.Loans[l].Balance)
	}

	//fmt.Print("Balances: ")
	//fmt.Println(balances)

	for {
		// One month has elapsed
		loan.Months++

		//This permutation already lost time or money, get out. 5% faster calculations with elimination check
		if (fastestResult.Months != 0 && fastestResult.Months < loan.Months && fastestResult.TotalPaid < loan.TotalPaid) ||
			(fastestResult.Months != 0 && cheapestResult.TotalPaid < loan.TotalPaid && cheapestResult.Months < loan.Months) {
			loan.Months = 9999
			loan.TotalPaid = 99999999999999999999999 //who is this rich? send me some money lol
			break
		}

		//reset the monthly extra payment counter
		canPayMonth := canPayExtra

		// fmt.Printf("Balances after %v month(s): ", loan.Months)
		// fmt.Println(balances)

		for _, l := range loan.Order {
			if balances[l] == 0 {
				canPayMonth += loans.Loans[l].Min //Rollover method
				//Complete loan
				continue
			}

			//Make the minimum payment
			balances[l] -= loans.Loans[l].Min
			loan.TotalPaid += loans.Loans[l].Min

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

			interest := balances[l] * loans.Loans[l].Apy / 12
			interest = float32(math.RoundToEven(float64(interest)*100) / 100) // Bank round?
			balances[l] += interest
		}

		// fmt.Printf("Balances after interest: ")
		// fmt.Println(balances)

		//If all balances are empty, we are done
		var totalBalances float32
		for _, bal := range balances {
			totalBalances += bal
		}
		if totalBalances == 0 {
			break // We are done!!
		}

		//Impossible payment plan, 50 years +
		if loan.Months >= 600 {
			//Need to send something to prevent deadlock
			loan.Months = 9999
			loan.TotalPaid = 99999999999999999999999
			break
		}
	}

	jobResults <- loan
}
