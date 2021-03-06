package random

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

// GetRandomInt get random int
func GetRandomInt(length int) *big.Int {
	// NewInt allocates and returns a new Int set to x.
	/*one := big.NewInt(1)
	// Lsh sets z = x << n and returns z.
	maxi := new(big.Int).Lsh(one, uint(length))

	// TODO: Random Seed, need to be replace!!!
	// New returns a new Rand that uses random values from src to generate other random values.
	// NewSource returns a new pseudo-random Source seeded with the given value.
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	// Rand sets z to a pseudo-random number in [0, n) and returns z.
	rndNum := new(big.Int).Rand(rnd, maxi)*/
	one := big.NewInt(1)
	maxi := new(big.Int).Lsh(one, uint(length))
	maxi = new(big.Int).Sub(maxi, one)
	rndNum, err := rand.Int(rand.Reader, maxi)
	if err != nil {
		return nil
	}

	return rndNum
}

// GetRandomIntFromZn get random int from n
func GetRandomIntFromZn(n *big.Int) *big.Int {
	var rndNumZn *big.Int
	zero := big.NewInt(0)

	for {
		rndNumZn = GetRandomInt(n.BitLen())
		if rndNumZn != nil && rndNumZn.Cmp(n) < 0 && rndNumZn.Cmp(zero) >= 0 {
			break
		}
	}

	return rndNumZn
}

// GetRandomIntFromZnStar get random int < n, >1,and gcd(val,n) = 1
func GetRandomIntFromZnStar(n *big.Int) *big.Int {
	var rndNumZnStar *big.Int
	gcdNum := big.NewInt(0)
	one := big.NewInt(1)

	for {
		rndNumZnStar = GetRandomInt(n.BitLen())
		if rndNumZnStar != nil && rndNumZnStar.Cmp(n) < 0 && rndNumZnStar.Cmp(one) >= 0 && gcdNum.GCD(nil, nil, rndNumZnStar, n).Cmp(one) == 0 {
			break
		}
	}

	return rndNumZnStar
}

// GetSafeRandomPrimeInt get safe big prime
func GetSafeRandomPrimeInt() (*big.Int, *big.Int) {
	var q *big.Int
	var p *big.Int
	var err error

	one := big.NewInt(1)
	two := big.NewInt(2)
	length := 1024 // L/2

	for {
		q, err = rand.Prime(rand.Reader, length-1)
		if err != nil {
			fmt.Println("Generate Safe Random Prime ERROR!")
			q = nil
			p = nil
			break
		}

		p = new(big.Int).Mul(q, two)
		p = new(big.Int).Add(p, one)
		if p.ProbablyPrime(512) {
			break
		}

		time.Sleep(time.Duration(10000)) //1000 000 000 == 1s
	}

	return q, p
}
