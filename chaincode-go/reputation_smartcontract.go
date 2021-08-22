package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Owner_rate gets parameters for evaluating the bot owner and changes his rating
func (s *SmartContract) Owner_rate(ctx contractapi.TransactionContextInterface, id string, U float32, I float32, T float32) (float32, error) {
	userJSON, err := ctx.GetStub().GetState(id)
	var Rep float32
	if err != nil {
		return 0, fmt.Errorf("failed to read from world state: %v", err)
	}
	if userJSON == nil {
		return 0, fmt.Errorf("the user %s does not exist", id)
	}
	if U >= -1 && U <= 1 && I >= 0 && I <= 1 && T >= -1 && T < 1 {
		Rep = U * I * T
	} else {
		return 0, fmt.Errorf("Wrong parameters for evaluating")
	}

	var user Owners
	err = json.Unmarshal(userJSON, &user)
	if err != nil {
		return 0, err
	}
	count := user.Owner_deals
	if count == 0 {
		rate := Owners{
			Owner_rate: Rep,
		}
		ownerJSON, err := json.Marshal(rate)
		if err != nil {
			return 0, err
		}
		return Rep, ctx.GetStub().PutState(id, ownerJSON)
	} else {
		//Rep = Upgrade(Rep)
		rate := Owners{
			Owner_rate: Rep,
		}
		ownerJSON, err := json.Marshal(rate)
		if err != nil {
			return 0, err
		}
		return Rep, ctx.GetStub().PutState(id, ownerJSON)
	}
}

// Renter_rate gets parameters for evaluating the bot renter and changes his rating
func (s *SmartContract) Renter_rate(ctx contractapi.TransactionContextInterface, id string, E float32) (float32, error) {
	userJSON, err := ctx.GetStub().GetState(id)
	var Rep float32
	if err != nil {
		return 0, fmt.Errorf("failed to read from world state: %v", err)
	}
	if userJSON == nil {
		return 0, fmt.Errorf("the user %s does not exist", id)
	}

	var user Renters
	err = json.Unmarshal(userJSON, &user)
	if err != nil {
		return 0, err
	}
	count := user.Renter_deals
	if count == 0 {
		Rep = E
		rate := Renters{
			Renter_rate: Rep,
		}
		renterJSON, err := json.Marshal(rate)
		if err != nil {
			return 0, err
		}
		return Rep, ctx.GetStub().PutState(id, renterJSON)
	} else {
		//Rep = Upgrade(Rep)
		rate := Renters{
			Renter_rate: Rep,
		}
		renterJSON, err := json.Marshal(rate)
		if err != nil {
			return 0, err
		}
		return Rep, ctx.GetStub().PutState(id, renterJSON)
	}
}
