package chaincode

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type Auction struct {
	Type        string          `json:"bot_type"`
	Bot_owner   string          `json:"owner_id"`
	BidsList    map[string]Bids `json:"bids"`
	Start_price int             `json:"start_price"`
	Highest_bid int             `json:"highest_bid"`
	Final_bid   int             `json:"final_bid"`
	Step        int             `json:"step"`
	Status      string          `json:"status"`
	Winner      string          `json:"winner"`
	Exp_time    string          `json:"expiration_time"`
}

// Bids is the structure of a reveald bids
type Bids struct {
	ID     string `json:"bidder_id"`
	Price  int    `json:"price"`
	Wallet string `json:"renter_wallet"`
}

const bidKeyType = "bid"

// min function finds the minimum of two integer numbers
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// CheckBalance function checks if the reneter able to pay off the bid
func (s *SmartContract) CheckBalance(ctx contractapi.TransactionContextInterface, BidPrice int, RenterID string) (bool, error) {

	renterJSON, err := ctx.GetStub().GetState(RenterID)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	if renterJSON == nil {
		return true, fmt.Errorf("the user %s does not exist", RenterID)
	}
	var renter Renters
	err = json.Unmarshal(renterJSON, &renter)
	if err != nil {
		return false, fmt.Errorf("failed to create renter object JSON: %v", err)
	}
	balance := renter.Renter_Balance - BidPrice
	if balance >= 0 {
		return false, nil
	} else {
		return true, nil
	}
}

// CreateAuction creates on auction on the public channel. The identity that
// submits the transacion becomes the seller of the auction
func (s *SmartContract) CreateAuction(ctx contractapi.TransactionContextInterface, auctionID string,
	OwnerID string, BotType string, StartPrice int, ExpTime string) error {

	userJSON, err := ctx.GetStub().GetState(OwnerID)

	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if userJSON == nil {
		return fmt.Errorf("the user %s does not exist", OwnerID)
	}

	revealedBids := make(map[string]Bids)

	auction := Auction{
		Type:        BotType,
		Bot_owner:   OwnerID,
		BidsList:    revealedBids,
		Start_price: StartPrice,
		Highest_bid: 0,
		Final_bid:   StartPrice,
		Step:        1,
		Status:      "open",
		Winner:      "nobody",
		Exp_time:    ExpTime,
	}

	auctionBytes, err := json.Marshal(auction)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(auctionID, auctionBytes)
	if err != nil {
		return fmt.Errorf("failed to put auction in public data: %v", err)
	}

	return nil
}

// JoinAuction allows to add a user's bid to the auction.
// The function returns the current bid
func (s *SmartContract) JoinAuction(ctx contractapi.TransactionContextInterface, auctionID string,
	ProposedPrice int, RenterWallet string, RenterID string) (int, error) {

	userJSON, err := ctx.GetStub().GetState(RenterID)

	if err != nil {
		return 0, fmt.Errorf("failed to read from world state: %v", err)
	}
	if userJSON == nil {
		return 0, fmt.Errorf("the user %s does not exist", RenterID)
	}
	auctionJSON, err := ctx.GetStub().GetState(auctionID)
	var auction Auction

	if auctionJSON == nil {
		return 0, fmt.Errorf("Auction not found: %v", auctionID)
	}
	err = json.Unmarshal(auctionJSON, &auction)
	if err != nil {
		return 0, fmt.Errorf("failed to create auction object JSON: %v", err)
	}

	last_bid := auction.Final_bid

	if auction.Exp_time > time.Now().Format("20060102150405") {
		if ProposedPrice > auction.Final_bid+auction.Step {
			if ProposedPrice >= auction.Highest_bid {
				auction.Final_bid = min(ProposedPrice, auction.Highest_bid+auction.Step)
				auction.Highest_bid = ProposedPrice
			}
			auction.Final_bid = min(ProposedPrice+auction.Step, auction.Highest_bid)
		}
		return 0, fmt.Errorf("incorrectly offered price, enter a price higher than %d", auction.Final_bid+auction.Step)
	} else {
		return 0, fmt.Errorf("the auction has expired")
	}

	status, err := CheckBalance(ctx, auction.Final_bid, RenterID)

	if err != nil {
		return 0, fmt.Errorf("failed to check the balance: %v", err)
	}
	if status == false {
		auction.Final_bid = last_bid
		return 0, fmt.Errorf("not enough funds to create a bid")
	}

	txID := ctx.GetStub().GetTxID()

	bidKey, err := ctx.GetStub().CreateCompositeKey(bidKeyType, []string{auctionID, txID})
	if err != nil {
		return 0, fmt.Errorf("failed to create composite key: %v", err)
	}
	NewBid := Bids{
		ID:     RenterID,
		Price:  ProposedPrice,
		Wallet: RenterWallet,
	}

	bidders := make(map[string]Bids)
	bidders = auction.BidsList
	bidders[bidKey] = NewBid
	auction.BidsList = bidders

	newAuctionBytes, _ := json.Marshal(auction)

	err = ctx.GetStub().PutState(auctionID, newAuctionBytes)
	if err != nil {
		return 0, fmt.Errorf("failed to update auction: %v", err)
	}

	return 0, nil
}

// QueryAuction allows all members of the channel to read a public auction
func (s *SmartContract) QueryAuction(ctx contractapi.TransactionContextInterface, auctionID string) (*Auction, error) {

	auctionJSON, err := ctx.GetStub().GetState(auctionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get auction object %v: %v", auctionID, err)
	}
	if auctionJSON == nil {
		return nil, fmt.Errorf("auction does not exist")
	}

	var auction *Auction
	err = json.Unmarshal(auctionJSON, &auction)
	if err != nil {
		return nil, err
	}

	return auction, nil
}

// EndAuction both changes the auction status to closed and calculates the winners
// of the auction
func (s *SmartContract) EndAuction(ctx contractapi.TransactionContextInterface, auctionID string) (int, string, error) {

	auctionBytes, err := ctx.GetStub().GetState(auctionID)
	if err != nil {
		return 0, "", fmt.Errorf("failed to get auction %v: %v", auctionID, err)
	}

	if auctionBytes == nil {
		return 0, "", fmt.Errorf("Auction interest object %v not found", auctionID)
	}

	var auction Auction
	err = json.Unmarshal(auctionBytes, &auction)
	if err != nil {
		return 0, "", fmt.Errorf("failed to create auction object JSON: %v", err)
	}
	auction.Status = string("closed")

	BidMap := auction.BidsList
	if len(auction.BidsList) == 0 {
		return 0, "", fmt.Errorf("No bids have been revealed, cannot end auction: %v", err)
	}

	// determine the highest bid
	for _, bid := range BidMap {
		if bid.Price > auction.Final_bid {
			auction.Winner = bid.ID
			auction.Final_bid = bid.Price
		}
	}

	return auction.Final_bid, auction.Winner, nil
}
