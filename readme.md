
# Loan Pay v1

The purpose of this project is to calculate the fastest and cheapest way to pay off your loans. Currently it is able to perform up to 11-14 loans withouth getting into insane times. This uses the premise of the [stacking method and snowball methods](https://www.thebalance.com/debt-snowball-vs-debt-stacking-453633). Inspiration came from this [stackoverflow question](https://money.stackexchange.com/questions/85187/algorithm-for-multiple-debt-payoff-to-minimize-time-in-debt).

#### Upcomming ideas:

v1.2 - Integrate the idea of balance transfers

v2 - Branch off main, Use remote workers on network to handoff calculations in bulk to spread the load and cut time.

v3 - Use golangjs package to port it over from same code base to a pwa github hosted site

#### How to use

Edit the loans.yaml file that has some examples and run it. The output will be the index array order of how to pay them off followed by the time in months and amount paid.


#### Benchmarks

Loans | Permutations | Time To calculate
-- | -- | --
1  | 1 | 5ms
2  | 2 | 5ms
3  | 6 | 5ms
4  | 24 | 5ms
5  | 120 | 5ms
6  | 720 | 5ms
7  | 5,040 | 5ms
8  | 40,320 | 5ms
9  | 362,880 | 5ms
10 | 3,628,800 | 7s
11 | 39,916,800 | 75s
12 | 479,001,600 | 15m
13 | 6,227,020,800 | 4h
14 | 87,178,291,200 | 46h
15 | 1,307,674,368,000 | 28d
16 | 20,922,789,888,000 | 1.25y
17 | 355,687,428,096,000 | 21y
18 | 6,402,373,705,728,000 | 382y
19 | 121,645,100,408,832,000 | 7248y
20 | 2,432,902,008,176,640,000 | 144952y

Time to calculate was run on a machine with a 5147 [benchmark score](https://www.cpubenchmark.net/). Your times may vary. Average memory usage was 105MB of ram throughout the process.


