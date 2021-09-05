package chaincode

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing users
type SmartContract struct {
	contractapi.Contract
}

// Owners describes basic details of registration new owner
type Owners struct {
	Owner_id      string         `json:"owner_id"`
	Bot_types     string         `json:"bot_types"`
	Costs         uint           `json:"costs"`
	Owner_key     *rsa.PublicKey `json:"owner_key"`
	Owner_rate    float32        `json:"owner_rate"`
	Owner_deals   uint           `json:"owner_deals"`
	Owner_Balance int            `json:"owner_balance"`
}

// Renters describes basic details of registration new renter
type Renters struct {
	Renter_id      string         `json:"renter_id"`
	Business_type  string         `json:"business_type"`
	Renter_key     *rsa.PublicKey `json:"renter_key"`
	Renter_rate    float32        `json:"renter_rate"`
	Renter_deals   uint           `json:"renter_deals"`
	Renter_Balance int            `json:"renter_balance"`
}

// Vehicles describes basic details of registration new vehicle
type Vehicles struct {
	Owner_veh_id string         `json:"owner_veh_id"`
	Vehicle_id   string         `json:"vehicle_id"`
	Type_of_work string         `json:"work_type"`
	Vehicle_key  *rsa.PublicKey `json:"vehicle_key"`
}

// Orders describes basic details of registration new orders
type Orders struct {
	Order_id         string  `json:"order_id"`
	Order_properties string  `json:"order_properties"`
	Vehicle_id       string  `json:"vehicle_id"`
	Price            float32 `json:"price"`
}

// InitLedger adds a base set of users to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	OwnerPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println(err.Error)
		os.Exit(1)
	}
	owners := []Owners{
		{Owner_id: "owner1", Bot_types: "delivery drones", Costs: 100, Owner_key: &OwnerPrivateKey.PublicKey, Owner_rate: 0, Owner_deals: 0},
	}

	for _, owner := range owners {
		ownerJSON, err := json.Marshal(owner)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(owner.Owner_id, ownerJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

// Owner_register registers a new owner to the world state with given details.
func (s *SmartContract) Owner_register(ctx contractapi.TransactionContextInterface,
	owner_id string, bot_types string, costs uint) (*rsa.PublicKey, error) {

	exists, err := s.ThisExists(ctx, owner_id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("the owner %s already exists", owner_id)
	}
	OwnerPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println(err.Error)
		os.Exit(1)
	}
	owner := Owners{
		Owner_id:   owner_id,
		Bot_types:  bot_types,
		Costs:      costs,
		Owner_key:  &OwnerPrivateKey.PublicKey,
		Owner_rate: 0,
	}
	ownerJSON, err := json.Marshal(owner)
	if err != nil {
		return nil, err
	}

	return &OwnerPrivateKey.PublicKey, ctx.GetStub().PutState(owner_id, ownerJSON)
}

// Renter_register registers a new renter to the world state with given details.
func (s *SmartContract) Renter_register(ctx contractapi.TransactionContextInterface,
	renter_id string, business_type string) (*rsa.PublicKey, error) {

	exists, err := s.ThisExists(ctx, renter_id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("the renter %s already exists", renter_id)
	}
	RenterPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println(err.Error)
		os.Exit(1)
	}
	renter := Renters{
		Renter_id:     renter_id,
		Business_type: business_type,
		Renter_key:    &RenterPrivateKey.PublicKey,
		Renter_rate:   0,
	}
	renterJSON, err := json.Marshal(renter)
	if err != nil {
		return nil, err
	}

	return &RenterPrivateKey.PublicKey, ctx.GetStub().PutState(renter_id, renterJSON)
}

// Add_vehicle registers a new bot to the world state with given details.
func (s *SmartContract) Add_vehicle(ctx contractapi.TransactionContextInterface,
	owner_veh_id string, vehicle_id string, type_of_work string) (*rsa.PublicKey, error) {

	exists, err := s.ThisExists(ctx, vehicle_id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("the vehicle %s already exists", vehicle_id)
	}
	owner_exists, err := s.ThisExists(ctx, owner_veh_id)
	if err != nil {
		return nil, err
	}
	if !owner_exists {
		return nil, fmt.Errorf("the vehicle owner %s is not registed", owner_veh_id)
	}

	VehiclePrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println(err.Error)
		os.Exit(1)
	}
	vehicle := Vehicles{
		Owner_veh_id: owner_veh_id,
		Vehicle_id:   vehicle_id,
		Type_of_work: type_of_work,
		Vehicle_key:  &VehiclePrivateKey.PublicKey,
	}
	vehicleJSON, err := json.Marshal(vehicle)
	if err != nil {
		return nil, err
	}

	return &VehiclePrivateKey.PublicKey, ctx.GetStub().PutState(vehicle_id, vehicleJSON)
}

// Create_order registers a new order to the world state with given details.
func (s *SmartContract) Create_order(ctx contractapi.TransactionContextInterface,
	order_properties string, renter_id string) error {

	exists, err := s.ThisExists(ctx, renter_id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the renter %s doesn't exist", renter_id)
	}
	order_id := time.Now().Format(time.RFC1123) + renter_id

	//vehicle_id := GetVehicleID(renter_id)
	//price := GetAuctionPrice(renter_id)
	order := Orders{
		Order_id:         order_id,
		Order_properties: order_properties,
		//Vehicle_id:     vehicle_id,
		//Price:		  price
	}
	orderJSON, err := json.Marshal(order)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(order_id, orderJSON)
}

// ReadLedger returns the user stored in the world state with given id.
func (s *SmartContract) ReadLedger(ctx contractapi.TransactionContextInterface, id string) (*Owners, error) {
	userJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if userJSON == nil {
		return nil, fmt.Errorf("the user %s does not exist", id)
	}

	var owner Owners
	err = json.Unmarshal(userJSON, &owner)
	if err != nil {
		return nil, err
	}

	return &owner, nil
}

// UpdateOwner updates an existing user in the world state with provided parameters.
func (s *SmartContract) UpdateOwner(ctx contractapi.TransactionContextInterface,
	owner_id string, bot_types string, costs uint) error {

	exists, err := s.ThisExists(ctx, owner_id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the owner %s does not exist", owner_id)
	}

	// overwriting original record with new parameters
	owner := Owners{
		Owner_id:  owner_id,
		Bot_types: bot_types,
		Costs:     costs,
	}
	ownerJSON, err := json.Marshal(owner)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(owner_id, ownerJSON)
}

// UpdateOwner updates an existing user in the world state with provided parameters.
func (s *SmartContract) UpdateRenter(ctx contractapi.TransactionContextInterface,
	renter_id string, business_type string) error {

	exists, err := s.ThisExists(ctx, renter_id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the renter %s does not exist", renter_id)
	}

	// overwriting original record with new parameters
	renter := Renters{
		Renter_id:     renter_id,
		Business_type: business_type,
	}
	renterJSON, err := json.Marshal(renter)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(renter_id, renterJSON)
}

// UpdateVehicle updates an existing bot in the world state with provided parameters.
func (s *SmartContract) UpdateVehicle(ctx contractapi.TransactionContextInterface, owner_veh_id string,
	vehicle_id string, type_of_work string) error {

	exists, err := s.ThisExists(ctx, vehicle_id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the vehicle %s does not exist", vehicle_id)
	}

	// overwriting original record with new parameters
	vehicle := Vehicles{
		Owner_veh_id: owner_veh_id,
		Vehicle_id:   vehicle_id,
		Type_of_work: type_of_work,
	}
	vehicleJSON, err := json.Marshal(vehicle)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(vehicle_id, vehicleJSON)
}

// DeleteThis deletes an given user or bot from the world state.
func (s *SmartContract) DeleteThis(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.ThisExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}

// ThisExists returns true when user with given ID exists in world state
func (s *SmartContract) ThisExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	thisJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return thisJSON != nil, nil
}

// Count_reward counts reward of the owner and tranfers money from renter to owner
func (s *SmartContract) Count_reward(ctx contractapi.TransactionContextInterface, universal_parameter float32, verdict bool, order_id string) (float32, error) {
	exists, err := s.ThisExists(ctx, order_id)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, fmt.Errorf("the order %s does not exist", order_id)
	}
	if !verdict {
		return 0, fmt.Errorf("the information provided has not been verified")
	}
	orderJSON, err := ctx.GetStub().GetState(order_id)
	var order Orders
	err = json.Unmarshal(orderJSON, &order)
	if err != nil {
		return 0, err
	}

	final_price := order.Price * universal_parameter
	//Payment methods will be implemented here

	return final_price, nil
}
