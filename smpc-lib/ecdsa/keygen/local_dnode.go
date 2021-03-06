/*
 *  Copyright (C) 2020-2021  AnySwap Ltd. All rights reserved.
 *  Copyright (C) 2020-2021  haijun.cai@anyswap.exchange
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

// Package keygen MPC implementation of generating pubkey 
package keygen

import (
	"errors"
	"fmt"
	"github.com/anyswap/Anyswap-MPCNode/crypto/secp256k1"
	"github.com/anyswap/Anyswap-MPCNode/internal/common/math/random"
	"github.com/anyswap/Anyswap-MPCNode/smpc-lib/crypto/ec2"
	"github.com/anyswap/Anyswap-MPCNode/smpc-lib/smpc"
	"math/big"
)

// LocalDNode current local node
type LocalDNode struct {
	*smpc.BaseDNode
	temp localTempData
	data LocalDNodeSaveData
	out  chan<- smpc.Message
	end  chan<- LocalDNodeSaveData
}

// localTempData  Store some data of MPC calculation process 
type localTempData struct {
	kgRound0Messages,
	kgRound1Messages,
	kgRound2Messages,
	kgRound2Messages1,
	kgRound2Messages2,
	kgRound3Messages,
	kgRound3Messages1,
	kgRound4Messages,
	kgRound5Messages,
	kgRound5Messages1,
	kgRound6Messages,
	kgRound6Messages1,
	kgRound7Messages []smpc.Message

	// temp data (thrown away after keygen)

	//round 1
	u1           *big.Int
	u1Poly       *ec2.PolyStruct2
	u1PolyG      *ec2.PolyGStruct2
	commitU1G    *ec2.Commitment
	c1           *big.Int
	commitC1G    *ec2.Commitment
	u1PaillierPk *ec2.PublicKey
	u1PaillierSk *ec2.PrivateKey
	// paillier.N = p*q
	p *big.Int 
	q *big.Int

	//round 2
	u1Shares []*ec2.ShareStruct2
	x	[]*big.Int

	//round 3

	//round 4
	// Ntilde = p1*p2
	p1 *big.Int
	p2 *big.Int
	commitXiG  *ec2.Commitment

	//round 5
	roh [][]*big.Int

	//round 6

	//round 7
}

// NewLocalDNode new a DNode data struct for current node
func NewLocalDNode(
	out chan<- smpc.Message,
	end chan<- LocalDNodeSaveData,
	DNodeCountInGroup int,
	threshold int,
	paillierkeylength int,
) smpc.DNode {

	data := NewLocalDNodeSaveData(DNodeCountInGroup)
	p := &LocalDNode{
		BaseDNode: new(smpc.BaseDNode),
		temp:      localTempData{},
		data:      data,
		out:       out,
		end:       end,
	}

	uid := random.GetRandomIntFromZn(secp256k1.S256().N)
	p.ID = fmt.Sprintf("%v", uid)
	fmt.Printf("=========== NewLocalDNode, uid = %v, p.ID = %v =============\n", uid, p.ID)

	p.DNodeCountInGroup = DNodeCountInGroup
	p.ThresHold = threshold
	p.PaillierKeyLength = paillierkeylength

	p.temp.roh = make([][]*big.Int,DNodeCountInGroup)
	p.temp.x = make([]*big.Int,DNodeCountInGroup)

	p.temp.kgRound0Messages = make([]smpc.Message, 0)
	p.temp.kgRound1Messages = make([]smpc.Message, DNodeCountInGroup)
	p.temp.kgRound2Messages = make([]smpc.Message, DNodeCountInGroup)
	p.temp.kgRound2Messages1 = make([]smpc.Message, DNodeCountInGroup)
	p.temp.kgRound2Messages2 = make([]smpc.Message, DNodeCountInGroup)
	p.temp.kgRound3Messages = make([]smpc.Message, DNodeCountInGroup)
	p.temp.kgRound3Messages1 = make([]smpc.Message, DNodeCountInGroup)
	p.temp.kgRound4Messages = make([]smpc.Message, DNodeCountInGroup)
	p.temp.kgRound5Messages = make([]smpc.Message, DNodeCountInGroup)
	p.temp.kgRound5Messages1 = make([]smpc.Message, DNodeCountInGroup)
	p.temp.kgRound6Messages = make([]smpc.Message, DNodeCountInGroup)
	p.temp.kgRound6Messages1 = make([]smpc.Message, DNodeCountInGroup)
	return p
}

// FirstRound first round
func (p *LocalDNode) FirstRound() smpc.Round {
	return newRound0(&p.data, &p.temp, p.out, p.end, p.ID, p.DNodeCountInGroup, p.ThresHold, p.PaillierKeyLength)
}

// FinalizeRound get finalize round
func (p *LocalDNode) FinalizeRound() smpc.Round {
	return nil
}

// Finalize weather gg20 round
func (p *LocalDNode) Finalize() bool {
	return false
}

// Start generating pubkey start 
func (p *LocalDNode) Start() error {
	return smpc.BaseStart(p)
}

// Update Collect data from other nodes and enter the next round 
func (p *LocalDNode) Update(msg smpc.Message) (ok bool, err error) {
	return smpc.BaseUpdate(p, msg)
}

// DNodeID get the ID of current DNode
func (p *LocalDNode) DNodeID() string {
	return p.ID
}

// SetDNodeID set the ID of current DNode
func (p *LocalDNode) SetDNodeID(id string) {
	p.ID = id
}

// CheckFull  Check for empty messages 
func CheckFull(msg []smpc.Message) bool {
	if len(msg) == 0 {
		return false
	}

	for _, v := range msg {
		if v == nil {
			return false
		}
	}

	return true
}

// StoreMessage Collect data from other nodes
func (p *LocalDNode) StoreMessage(msg smpc.Message) (bool, error) {
	switch msg.(type) {
	case *KGRound0Message:
		if len(p.temp.kgRound0Messages) < p.DNodeCountInGroup {
			p.temp.kgRound0Messages = append(p.temp.kgRound0Messages, msg)
		}

		if len(p.temp.kgRound0Messages) == p.DNodeCountInGroup {
			//fmt.Printf("================ StoreMessage,get all 0 messages ==============\n")
			//time.Sleep(time.Duration(120) * time.Second) //tmp code
			return true, nil
		}
	case *KGRound1Message:
		index := msg.GetFromIndex()
		p.temp.kgRound1Messages[index] = msg
		m := msg.(*KGRound1Message)
		p.data.U1PaillierPk[index] = m.U1PaillierPk
		if len(p.temp.kgRound1Messages) == p.DNodeCountInGroup && CheckFull(p.temp.kgRound1Messages) {
			//fmt.Printf("================ StoreMessage,get all 1 messages ==============\n")
			//time.Sleep(time.Duration(20) * time.Second) //tmp code
			return true, nil
		}
	case *KGRound2Message:
		index := msg.GetFromIndex()
		p.temp.kgRound2Messages[index] = msg
		if len(p.temp.kgRound2Messages) == p.DNodeCountInGroup && CheckFull(p.temp.kgRound2Messages) {
			//fmt.Printf("================ StoreMessage,get all 2 messages ==============\n")
			//time.Sleep(time.Duration(20) * time.Second) //tmp code
			return true, nil
		}
	case *KGRound2Message1:
		index := msg.GetFromIndex()
		p.temp.kgRound2Messages1[index] = msg
		if len(p.temp.kgRound2Messages1) == p.DNodeCountInGroup && CheckFull(p.temp.kgRound2Messages1) {
			//fmt.Printf("================ StoreMessage,get all 2-1 messages ==============\n")
			//time.Sleep(time.Duration(20) * time.Second) //tmp code
			return true, nil
		}
	case *KGRound2Message2:
		index := msg.GetFromIndex()
		p.temp.kgRound2Messages2[index] = msg
		if len(p.temp.kgRound2Messages2) == p.DNodeCountInGroup && CheckFull(p.temp.kgRound2Messages2) {
			//fmt.Printf("================ StoreMessage,get all 2-1 messages ==============\n")
			//time.Sleep(time.Duration(20) * time.Second) //tmp code
			return true, nil
		}
	case *KGRound3Message:
		index := msg.GetFromIndex()
		p.temp.kgRound3Messages[index] = msg
		if len(p.temp.kgRound3Messages) == p.DNodeCountInGroup && CheckFull(p.temp.kgRound3Messages) {
			//fmt.Printf("================ StoreMessage,get all 3 messages ==============\n")
			//time.Sleep(time.Duration(20) * time.Second) //tmp code
			return true, nil
		}
	case *KGRound3Message1:
		index := msg.GetFromIndex()
		p.temp.kgRound3Messages1[index] = msg
		if len(p.temp.kgRound3Messages1) == p.DNodeCountInGroup && CheckFull(p.temp.kgRound3Messages1) {
			//fmt.Printf("================ StoreMessage,get all 3 messages ==============\n")
			//time.Sleep(time.Duration(20) * time.Second) //tmp code
			return true, nil
		}
	case *KGRound4Message:
		index := msg.GetFromIndex()
		m := msg.(*KGRound4Message)

		////////add for ntilde zk proof check
		H1 := m.U1NtildeH1H2.H1
		H2 := m.U1NtildeH1H2.H2
		Ntilde := m.U1NtildeH1H2.Ntilde
		pf1 := m.NtildeProof1
		pf2 := m.NtildeProof2
		//fmt.Printf("=========================keygen StoreMessage, message 4, curindex = %v, h1 = %v, h2 = %v, ntilde = %v, pf1 = %v, pf2 = %v ===========================\n", index, H1, H2, Ntilde, pf1, pf2)
		zero,_ := new(big.Int).SetString("0",10)
		one,_ := new(big.Int).SetString("1",10)
		h1modn := new(big.Int).Mod(H1,Ntilde)
		h2modn := new(big.Int).Mod(H2,Ntilde)
		if h1modn.Cmp(zero) == 0 || h2modn.Cmp(zero) == 0 {
		    return false,errors.New("h1 or h2 is equal 0 mod Ntilde")
		}
		if h1modn.Cmp(one) == 0 || h2modn.Cmp(one) == 0 {
		    return false,errors.New("h1 or h2 is equal 1 mod Ntilde")
		}

		if h1modn.Cmp(h2modn) == 0 {
			return false, errors.New("h1 and h2 were equal mod Ntilde")
		}
		
		if !pf1.Verify(H1, H2, Ntilde) || !pf2.Verify(H2, H1, Ntilde) {
			return false, errors.New("ntilde zk proof check fail")
		}
		////////

		p.data.U1NtildeH1H2[index] = m.U1NtildeH1H2
		p.temp.kgRound4Messages[index] = msg
		if len(p.temp.kgRound4Messages) == p.DNodeCountInGroup && CheckFull(p.temp.kgRound4Messages) {
			//fmt.Printf("================ StoreMessage,get all 4 messages ==============\n")
			//time.Sleep(time.Duration(20) * time.Second) //tmp code
			return true, nil
		}
	case *KGRound5Message:
		index := msg.GetFromIndex()
		p.temp.kgRound5Messages[index] = msg
		if len(p.temp.kgRound5Messages) == p.DNodeCountInGroup && CheckFull(p.temp.kgRound5Messages) {
			//fmt.Printf("================ StoreMessage,get all 5 messages ==============\n")
			//time.Sleep(time.Duration(20) * time.Second) //tmp code
			return true, nil
		}
	case *KGRound5Message1:
		index := msg.GetFromIndex()
		p.temp.kgRound5Messages1[index] = msg
		if len(p.temp.kgRound5Messages1) == p.DNodeCountInGroup && CheckFull(p.temp.kgRound5Messages1) {
			//time.Sleep(time.Duration(20) * time.Second) //tmp code
			return true, nil
		}
	case *KGRound6Message:
		index := msg.GetFromIndex()
		p.temp.kgRound6Messages[index] = msg
		if len(p.temp.kgRound6Messages) == p.DNodeCountInGroup && CheckFull(p.temp.kgRound6Messages) {
			//fmt.Printf("================ StoreMessage,get all 6 messages ==============\n")
			//time.Sleep(time.Duration(20) * time.Second) //tmp code
			return true, nil
		}
	case *KGRound6Message1:
		index := msg.GetFromIndex()
		p.temp.kgRound6Messages1[index] = msg
		if len(p.temp.kgRound6Messages1) == p.DNodeCountInGroup && CheckFull(p.temp.kgRound6Messages1) {
			//time.Sleep(time.Duration(20) * time.Second) //tmp code
			return true, nil
		}
	default: // unrecognised message, just ignore!
		fmt.Printf("storemessage,unrecognised message ignored: %v\n", msg)
		return false, nil
	}

	return false, nil
}


