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

type loansStruct struct {
	Transfer struct {
		Name      string  `yaml:"name"`
		Apy       float32 `yaml:"apy"`
		Fee       float32 `yaml:"fee"`
		Term      int     `yaml:"term"`
		Available float32 `yaml:"available"`
	} `yaml:"transfer"`
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
	Transfer  struct {
		Did    bool
		Loan   int16
		When   int16
		Amount float32
		Saved  float32
	}
}

var loans loansStruct

var jobLoan chan result
var jobResults chan result
var fastestResult result
var cheapestResult result

var threads = 1

func main() {

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

		//reset the monthly extra payment counter
		canPayMonth := canPayExtra

		// fmt.Printf("Balances after %v month(s): ", loan.Months)
		// fmt.Println(balances)

		// Check to see if a balance transfer is possible
		if !loan.Transfer.Did {
			for _, l := range loan.Order {
				if balances[l] == 0 {
					//skip, empty balance
					continue
				}

				// Will we be able to pay this loan off before exiting zero interest?
				if (balances[l] / float32(loans.Transfer.Term)) < (loans.Loans[l].Min + canPayExtra) {
					totalTransfer := (balances[l] * loans.Transfer.Fee) + balances[l]

					var totalNormal float32
					thisBalance := balances[l]
					fmt.Println(thisBalance)
					for thisBalance != 0 {
						thisBalance -= loans.Loans[l].Min
						totalNormal += loans.Loans[l].Min

						if thisBalance <= 0 {
							//Done, readd leftover
							totalNormal -= thisBalance
							thisBalance = 0
						} else {
							//Calculate interest
							thisBalance = (thisBalance * loans.Loans[l].Apy) + thisBalance
							//Not even gonna round cuz this is theoreticall anyway
						}

						fmt.Println(thisBalance)
					}

					//Yeah!!!! we saved some cash, lets transfer this loan
					if totalNormal > totalTransfer {
						fmt.Printf("Can do transfer on %v, calculating for %v > %v\n", l, totalNormal, totalTransfer)

						//Tell the user when and what we transfered
						loan.Transfer.Did = true
						loan.Transfer.Loan = int16(l)
						loan.Transfer.When = loan.Months
						loan.Transfer.Amount = balances[l]
						loan.Transfer.Saved = totalNormal - totalTransfer
						newMinimum := totalTransfer / float32(loans.Transfer.Term)
						newMinimum = float32(math.RoundToEven(float64(newMinimum)*100) / 100)

						//Current loan becomes new loan
						balances[l] = totalTransfer

						//We dont plan on paying interest
						loans.Loans[l].Apy = 0.00

						//if we cant pay using old minimum, take from canPayExtra
						if newMinimum > loans.Loans[l].Min {
							canPayExtra -= (newMinimum - loans.Loans[l].Min)
							fmt.Printf("need more cash to pay\n")
						}

						//Figured out where the money comes from, set that shit
						loans.Loans[l].Min = newMinimum

						os.Exit(1)
					}

				}
			}
		}

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
			return
		}
	}

	jobResults <- loan
}
