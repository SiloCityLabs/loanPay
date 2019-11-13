
# Loan Pay v1

The purpose of this project is to calculate the fastest and cheapest way to pay off your loans. Currently it is able to perform up to 11-14 loans withouth getting into insane times. This uses the premise of the [stacking method and snowball methods](https://www.thebalance.com/debt-snowball-vs-debt-stacking-453633). Inspiration came from this [stackoverflow question](https://money.stackexchange.com/questions/85187/algorithm-for-multiple-debt-payoff-to-minimize-time-in-debt).

#### Upcomming ideas:

v1.2 - Integrate the idea of balance transfers

v2 - Branch off main, Use remote workers on network to handoff calculations in bulk to spread the load and cut time.

v3 - Use golangjs package to port it over from same code base to a pwa github hosted site

#### How to use

Edit the loans.yaml file that has some examples and run it from command line. The output will be the index array order of how to pay them off followed by the time in months and amount paid.


#### Benchmarks

Loans | Permutations | Time To calculate
-- | -- | --
1  | 1 | >1s
2  | 2 | >1s
3  | 6 | >1s
4  | 24 | >1s
5  | 120 | >1s
6  | 720 | >1s
7  | 5,040 | >1s
8  | 40,320 | >1s
9  | 362,880 | >1s
10 | 3,628,800 | 4s
11 | 39,916,800 | 48s
12 | 479,001,600 | 10m
13 | 6,227,020,800 | 2h
14 | 87,178,291,200 | 2d
15 | 1,307,674,368,000 | 219d

Time to calculate was run on a machine with Windows and a 11381 [benchmark score](https://www.cpubenchmark.net/). Your times may vary. Average memory usage was 6MB of ram throughout the process. Im workings towards achieving 14 loans in under 24 hours
