package domain

import (
	"errors"
	"time"

	uuid "github.com/satori/go.uuid"
)

type Material struct {
	UID            uuid.UUID        `json:"uid"`
	Name           string           `json:"name"`
	PricePerUnit   Money            `json:"price_per_unit"`
	Type           MaterialType     `json:"type"`
	Quantity       MaterialQuantity `json:"quantity"`
	ExpirationDate *time.Time       `json:"expiration_date"`
	Notes          *string          `json:"notes"`
	IsExpense      *bool            `json:"is_expense"`
	ProducedBy     *string          `json:"produced_by"`

	// Events
	Version            int
	UncommittedChanges []interface{}
}

const (
	MoneyEUR = "EUR"
	MoneyIDR = "IDR"
)

type Money interface {
	Code() string
	Symbol() string
	Amount() string
	SetAmount(amount string)
}

type EUR struct {
	amount string
}

func (e EUR) Code() string {
	return MoneyEUR
}

func (e EUR) Symbol() string {
	return "€"
}

func (e EUR) Amount() string {
	return e.amount
}

func (e *EUR) SetAmount(amount string) {
	e.amount = amount
}

func CreateMoney(price, priceUnit string) (Money, error) {
	if price == "" {
		return nil, errors.New("price cannot be empty")
	}

	var money Money
	switch priceUnit {
	case EUR{}.Code():
		money = &EUR{}
		money.SetAmount(price)
	default:
		return nil, errors.New("money not found")
	}

	return money, nil
}

const (
	MaterialUnitSeeds      = "SEEDS"
	MaterialUnitPackets    = "PACKETS"
	MaterialUnitGram       = "GRAM"
	MaterialUnitKilogram   = "KILOGRAM"
	MaterialUnitBags       = "BAGS"
	MaterialUnitBottles    = "BOTTLES"
	MaterialUnitCubicMetre = "CUBIC_METRE"
	MaterialUnitPieces     = "PIECES"
	MaterialUnitUnits      = "UNITS"
)

type MaterialQuantity struct {
	Value float32              `json:"value"`
	Unit  MaterialQuantityUnit `json:"unit"`
}

type MaterialQuantityUnit struct {
	Code  string `json:"code"`
	Label string `json:"label"`
}

func MaterialQuantityUnits(materialTypeCode string) []MaterialQuantityUnit {
	switch materialTypeCode {
	case MaterialTypeSeedCode:
		return []MaterialQuantityUnit{
			{Code: MaterialUnitSeeds, Label: "Seeds"},
			{Code: MaterialUnitPackets, Label: "Packets"},
			{Code: MaterialUnitGram, Label: "Gram"},
			{Code: MaterialUnitKilogram, Label: "Kilogram"},
		}
	case MaterialTypeAgrochemicalCode:
		return []MaterialQuantityUnit{
			{Code: MaterialUnitPackets, Label: "Packets"},
			{Code: MaterialUnitBottles, Label: "Bottles"},
			{Code: MaterialUnitBags, Label: "Bags"},
		}
	case MaterialTypeGrowingMediumCode:
		return []MaterialQuantityUnit{
			{Code: MaterialUnitBags, Label: "Bags"},
			{Code: MaterialUnitCubicMetre, Label: "Cubic Metre"},
		}
	case MaterialTypeLabelAndCropSupportCode:
		return []MaterialQuantityUnit{
			{Code: MaterialUnitPieces, Label: "Pieces"},
		}
	case MaterialTypeSeedingContainerCode:
		return []MaterialQuantityUnit{
			{Code: MaterialUnitPieces, Label: "Pieces"},
		}
	case MaterialTypePostHarvestSupplyCode:
		return []MaterialQuantityUnit{
			{Code: MaterialUnitPieces, Label: "Pieces"},
		}
	case MaterialTypePlantCode:
		return []MaterialQuantityUnit{
			{Code: MaterialUnitUnits, Label: "Units"},
			{Code: MaterialUnitPackets, Label: "Packets"},
		}
	case MaterialTypeOtherCode:
		return []MaterialQuantityUnit{
			{Code: MaterialUnitPieces, Label: "Pieces"},
		}
	}

	return nil
}

func GetMaterialQuantityUnit(materialTypeCode string, code string) MaterialQuantityUnit {
	for _, v := range MaterialQuantityUnits(materialTypeCode) {
		if v.Code == code {
			return v
		}
	}

	return MaterialQuantityUnit{}
}

func (state *Material) TrackChange(event interface{}) {
	state.UncommittedChanges = append(state.UncommittedChanges, event)
	state.Transition(event)
}

func (state *Material) Transition(event interface{}) {
	switch e := event.(type) {
	case MaterialCreated:
		state.UID = e.UID
		state.Name = e.Name
		state.PricePerUnit = e.PricePerUnit
		state.Type = e.Type
		state.Quantity = e.Quantity
		state.ExpirationDate = e.ExpirationDate
		state.Notes = e.Notes
		state.ProducedBy = e.ProducedBy
		state.IsExpense = e.IsExpense
	}
}

func CreateMaterial(
	name string,
	price string,
	priceUnit string,
	materialType MaterialType,
	quantity float32,
	quantityUnit string,
	expirationDate *time.Time,
	notes *string,
	producedBy *string,
	isExpense *bool) (*Material, error) {

	uid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	money, err := CreateMoney(price, priceUnit)
	if err != nil {
		return nil, err
	}

	if materialType == nil {
		return nil, errors.New("cannot be empty")
	}

	err = validateQuantity(quantity)
	if err != nil {
		return nil, err
	}

	qu, err := validateQuantityUnit(quantityUnit, materialType)
	if err != nil {
		return nil, err
	}

	initial := &Material{
		UID:          uid,
		Name:         name,
		PricePerUnit: money,
		Type:         materialType,
		Quantity: MaterialQuantity{
			Value: quantity,
			Unit:  qu,
		},
		ExpirationDate: expirationDate,
		Notes:          notes,
		ProducedBy:     producedBy,
		IsExpense:      isExpense,
	}

	initial.TrackChange(MaterialCreated{
		UID:            initial.UID,
		Name:           initial.Name,
		PricePerUnit:   initial.PricePerUnit,
		Type:           initial.Type,
		Quantity:       initial.Quantity,
		ExpirationDate: initial.ExpirationDate,
		Notes:          initial.Notes,
		ProducedBy:     initial.ProducedBy,
		IsExpense:      initial.IsExpense,
	})

	return initial, nil
}

func (m *Material) ChangeName(name string) error {
	if name == "" {
		return errors.New("cannot be empty")
	}

	if len(name) <= 5 {
		return errors.New("too few characters")
	}

	m.TrackChange(MaterialNameChanged{UID: m.UID, Name: name})

	return nil
}

func (m *Material) ChangePricePerUnit(price, priceUnit string) error {
	money, err := CreateMoney(price, priceUnit)
	if err != nil {
		return err
	}

	m.TrackChange(MaterialPriceChanged{UID: m.UID, Price: money})

	return nil
}

func (m *Material) ChangeQuantityUnit(quantity float32, quantityUnit string, materialType MaterialType) error {
	err := validateQuantity(quantity)
	if err != nil {
		return err
	}

	qu, err := validateQuantityUnit(quantityUnit, materialType)
	if err != nil {
		return err
	}

	m.TrackChange(MaterialQuantityChanged{
		UID: m.UID,
		Quantity: MaterialQuantity{
			Value: quantity,
			Unit:  qu,
		},
	})

	return nil
}

func validateQuantity(quantity float32) error {
	if quantity <= 0 {
		return errors.New("Cannot be empty")
	}

	return nil
}

func validateQuantityUnit(quantityUnit string, materialType MaterialType) (MaterialQuantityUnit, error) {
	qu := GetMaterialQuantityUnit(materialType.Code(), quantityUnit)

	if qu == (MaterialQuantityUnit{}) {
		return MaterialQuantityUnit{}, errors.New("Cannot be empty")
	}

	return qu, nil
}