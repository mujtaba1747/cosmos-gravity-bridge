package types

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	// GravityDenomPrefix indicates the prefix for all assests minted by this module
	GravityDenomPrefix = ModuleName

	// GravityDenomSeparator is the separator for gravity denoms
	GravityDenomSeparator = ""

	// ETHContractAddressLen is the length of contract address strings
	ETHContractAddressLen = 42

	// GravityDenomLen is the length of the denoms generated by the gravity module
	GravityDenomLen = len(GravityDenomPrefix) + len(GravityDenomSeparator) + ETHContractAddressLen

	// ZeroAddress is an EthAddress containing 0x0000000000000000000000000000000000000000
	ZeroAddressString = "0x0000000000000000000000000000000000000000"
)

// Regular EthAddress
type EthAddress struct {
	address string
}

// Returns the contained address as a string
func (ea EthAddress) GetAddress() string {
	return ea.address
}

// Sets the contained address, performing validation before updating the value
func (ea EthAddress) SetAddress(address string) error {
	if err := ValidateEthAddress(address); err != nil {
		return err
	}
	ea.address = address
	return nil
}

// Creates a new EthAddress from a string, performing validation and returning any validation errors
func NewEthAddress(address string) (*EthAddress, error) {
	if err := ValidateEthAddress(address); err != nil {
		return nil, sdkerrors.Wrap(err, "invalid input address")
	}
	addr := EthAddress{address}
	return &addr, nil
}

// Returns a new EthAddress with 0x0000000000000000000000000000000000000000 as the wrapped address
func ZeroAddress() *EthAddress {
	return &EthAddress{ZeroAddressString}
}

// Validates the input string as an Ethereum Address
// Addresses must not be empty, have 42 character length, start with 0x and have 40 remaining characters in [0-9a-fA-F]
func ValidateEthAddress(address string) error {
	if address == "" {
		return fmt.Errorf("empty")
	}
	if len(address) != ETHContractAddressLen {
		return fmt.Errorf("address(%s) of the wrong length exp(%d) actual(%d)", address, ETHContractAddressLen, len(address))
	}
	if !regexp.MustCompile("^0x[0-9a-fA-F]{40}$").MatchString(address) {
		return fmt.Errorf("address(%s) doesn't pass regex", address)
	}

	return nil
}

// Performs validation on the wrapped string
func (ea EthAddress) ValidateBasic() error {
	return ValidateEthAddress(ea.address)
}

// EthAddrLessThan migrates the Ethereum address less than function
func EthAddrLessThan(e *EthAddress, o *EthAddress) bool {
	return bytes.Compare([]byte(e.GetAddress())[:], []byte(o.GetAddress())[:]) == -1
}

// Nillable EthAddress
type OptionalEthAddress struct {
	isNil    bool
	optional *EthAddress
}

// Indicates whether the wrapped address is nil or not
func (o OptionalEthAddress) IsNil() bool {
	return o.isNil
}

// Returns the wrapped address if it exists, or an error if it is nil
func (o OptionalEthAddress) Unwrap() (*EthAddress, error) {
	if o.isNil {
		return nil, fmt.Errorf("nil value")
	}
	return o.optional, nil
}

// Sets the wrapped EthAddress
func (o OptionalEthAddress) SetEthAddress(ethAddress *EthAddress) {
	if ethAddress == nil {
		o.isNil = true
		o.optional = nil
		return
	}

	o.isNil = false
	o.optional = ethAddress
}

// Creates an OptionalEthAddress where the wrapped EthAddress is nil
func NilEthAddress() *OptionalEthAddress {
	return &OptionalEthAddress{
		isNil:    true,
		optional: nil,
	}
}

// Creates a new OptionalEthAddress, returns any error from calling NewEthAddress
func NewOptionalEthAddress(address string) (*OptionalEthAddress, error) {
	ethAddress, err := NewEthAddress(address)
	return &OptionalEthAddress{
		isNil:    ethAddress == nil,
		optional: ethAddress,
	}, err
}

// Checks for invalid combinations of isNil and the wrapped EthAddress, also returns any errors from creating the
func (oea OptionalEthAddress) ValidateBasic() error {
	if oea.isNil && oea.optional != nil {
		return fmt.Errorf("supposed to be nil but isn't")
	}
	if !oea.isNil && oea.optional == nil {
		return fmt.Errorf("not supposed to be nil but is")
	}
	if !oea.isNil {
		err := oea.optional.ValidateBasic()
		// TODO: Reconsider this
		//if err != nil && !strings.Contains(err.Error(), "empty") {
		return err
		//}
	}
	return nil
}

/////////////////////////
//     ERC20Token      //
/////////////////////////

// NewERC20Token returns a new instance of an ERC20
func NewERC20Token(amount uint64, contract EthAddress) *ERC20Token {
	return &ERC20Token{Amount: sdk.NewIntFromUint64(amount), Contract: contract.GetAddress()}
}

func NewSDKIntERC20Token(amount sdk.Int, contract EthAddress) *ERC20Token {
	return &ERC20Token{Amount: amount, Contract: contract.GetAddress()}
}

// GravityCoin returns the gravity representation of the ERC20
func (e *ERC20Token) GravityCoin() sdk.Coin {
	contract, err := NewEthAddress(e.Contract)
	if err != nil {
		panic(fmt.Sprintf("invalid contract address", err))
	}
	return sdk.NewCoin(GravityDenom(contract), e.Amount)
}

func GravityDenom(tokenContract *EthAddress) string {
	return fmt.Sprintf("%s%s%s", GravityDenomPrefix, GravityDenomSeparator, tokenContract.GetAddress())
}

// ValidateBasic permforms stateless validation
func (e *ERC20Token) ValidateBasic() error {
	_, err := NewEthAddress(e.Contract)
	if err != nil {
		return sdkerrors.Wrap(err, "ethereum address")
	}
	// TODO: Validate all the things
	return nil
}

// Add adds one ERC20 to another
// TODO: make this return errors instead
func (e *ERC20Token) Add(o *ERC20Token) *ERC20Token {
	eContract, eErr := NewEthAddress(e.Contract)
	if eErr != nil {
		panic(fmt.Sprintf("invalid parent contract address: %v", eErr))
	}
	oContract, oErr := NewEthAddress(o.Contract)
	if oErr != nil {
		panic(fmt.Sprintf("invalid argument contract address: %v", oErr))
	}
	if eContract.GetAddress() != oContract.GetAddress() {
		panic("inconsistent contract addresses")
	}
	sum := e.Amount.Add(o.Amount)
	if !sum.IsUint64() {
		panic("invalid amount")
	}
	return NewERC20Token(sum.Uint64(), *eContract)
}

func GravityDenomToERC20(denom string) (*EthAddress, error) {
	fullPrefix := GravityDenomPrefix + GravityDenomSeparator
	if !strings.HasPrefix(denom, fullPrefix) {
		return nil, fmt.Errorf("denom prefix(%s) not equal to expected(%s)", denom, fullPrefix)
	}
	contract, err := NewEthAddress(strings.TrimPrefix(denom, fullPrefix))
	switch {
	case err != nil:
		return nil, fmt.Errorf("error(%s) validating ethereum contract address", err)
	case len(denom) != GravityDenomLen:
		return nil, fmt.Errorf("len(denom)(%d) not equal to GravityDenomLen(%d)", len(denom), GravityDenomLen)
	default:
		return contract, nil
	}
}
