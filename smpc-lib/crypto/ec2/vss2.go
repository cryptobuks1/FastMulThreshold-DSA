/*
 *  Copyright (C) 2020-2021  AnySwap Ltd. All rights reserved.
 *  Copyright (C) 2020-2021  xing.chang@anyswap.exchange
 *
 *  This library is free software; you can redistribute it and/or
 *  modify it under the Apache License, Version 2.0.
 *
 *  This library is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
 *
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */

package ec2

import (
	"math/big"
	"errors"

	s256 "github.com/anyswap/Anyswap-MPCNode/crypto/secp256k1"
	"github.com/anyswap/Anyswap-MPCNode/internal/common/math/random"
)

// PolyGStruct2 (x,y)
type PolyGStruct2 struct {
	PolyG [][]*big.Int //x and y
}

// PolyStruct2 coefficient set
type PolyStruct2 struct {
	Poly []*big.Int // coefficient set
}

// ShareStruct2 f(xi)
type ShareStruct2 struct {
	ID    *big.Int // ID, x coordinate
	Share *big.Int
}

// GetSharesID get ID
func GetSharesID(ss *ShareStruct2) *big.Int {
	if ss != nil {
		return ss.ID
	}

	return nil
}

// Vss2Init  Initialize Lagrange polynomial coefficients 
func Vss2Init(secret *big.Int, t int) (*PolyStruct2, *PolyGStruct2, error) {

	poly := make([]*big.Int, 0)
	polyG := make([][]*big.Int, 0)

	poly = append(poly, secret)
	pointX, pointY := s256.S256().ScalarBaseMult(secret.Bytes())
	polyG = append(polyG, []*big.Int{pointX, pointY})

	for i := 0; i < t-1; i++ {
		rndInt := random.GetRandomIntFromZn(s256.S256().N)
		poly = append(poly, rndInt)

		pointX, pointY := s256.S256().ScalarBaseMult(rndInt.Bytes())
		polyG = append(polyG, []*big.Int{pointX, pointY})
	}
	polyStruct := &PolyStruct2{Poly: poly}
	polyGStruct := &PolyGStruct2{PolyG: polyG}

	return polyStruct, polyGStruct, nil
}

// Vss2  Calculate Lagrange polynomial value 
func (polyStruct *PolyStruct2) Vss2(ids []*big.Int) ([]*ShareStruct2, error) {

	shares := make([]*ShareStruct2, 0)

	for i := 0; i < len(ids); i++ {
		shareVal,err := calculatePolynomial2(polyStruct.Poly, ids[i])
		if err != nil {
		    return nil,errors.New("calc share error")
		}

		shareStruct := &ShareStruct2{ID: ids[i], Share: shareVal}
		shares = append(shares, shareStruct)
	}

	return shares, nil
}

// Verify2 Verify Lagrange polynomial value
func (share *ShareStruct2) Verify2(polyG *PolyGStruct2) bool {

	idVal := share.ID

	computePointX, computePointY := polyG.PolyG[0][0], polyG.PolyG[0][1]

	for i := 1; i < len(polyG.PolyG); i++ {
		pointX, pointY := s256.S256().ScalarMult(polyG.PolyG[i][0], polyG.PolyG[i][1], idVal.Bytes())

		computePointX, computePointY = s256.S256().Add(computePointX, computePointY, pointX, pointY)
		idVal = new(big.Int).Mul(idVal, share.ID)
		idVal = new(big.Int).Mod(idVal, s256.S256().N)
	}

	originalPointX, originalPointY := s256.S256().ScalarBaseMult(share.Share.Bytes())

	if computePointX.Cmp(originalPointX) == 0 && computePointY.Cmp(originalPointY) == 0 {
		return true
	}
	
	return false
}

// Combine2 Calculating Lagrange interpolation formula 
func Combine2(shares []*ShareStruct2) (*big.Int, error) {

	// build x coordinate set
	xSet := make([]*big.Int, 0)
	for _, share := range shares {
		xSet = append(xSet, share.ID)
	}

	// for
	secret := big.NewInt(0)

	for i, share := range shares {
		times := big.NewInt(1)

		// calculate times()
		for j := 0; j < len(xSet); j++ {
			if j != i {
				sub := new(big.Int).Sub(xSet[j], share.ID)
				subInverse := new(big.Int).ModInverse(sub, s256.S256().N)
				if subInverse == nil {
				    return nil,errors.New("calc times fail")
				}
				div := new(big.Int).Mul(xSet[j], subInverse)
				times = new(big.Int).Mul(times, div)
				times = new(big.Int).Mod(times, s256.S256().N)
			}
		}

		// calculate sum(f(x) * times())
		fTimes := new(big.Int).Mul(share.Share, times)
		secret = new(big.Int).Add(secret, fTimes)
		secret = new(big.Int).Mod(secret, s256.S256().N)
	}

	return secret, nil
}

func calculatePolynomial2(poly []*big.Int, id *big.Int) (*big.Int,error) {
    idnum := new(big.Int).Mod(id,s256.S256().N)
    if idnum.Cmp(zero) == 0 || id.Cmp(zero) == 0 {
	return nil,errors.New("id can not be equal to 0 or 0 modulo the order of the curve")
    }

	lastIndex := len(poly) - 1
	result := poly[lastIndex]

	for i := lastIndex - 1; i >= 0; i-- {
		result = new(big.Int).Mul(result, id)
		result = new(big.Int).Add(result, poly[i])
		result = new(big.Int).Mod(result, s256.S256().N)
	}

	return result,nil
}
