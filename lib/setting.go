package lib

import (
	"assist-tix/dto"
	"assist-tix/entity"
	"strconv"

	"github.com/rs/zerolog/log"
)

const (
	// Static
	EventGarudaIdVerificationSettingName              = "IS_GARUDA_ID_VERIFICATION_ACTIVE"
	EventPurchaseAdultTicketPerTransactionSettingName = "MAX_ADULT_TICKET_PURCHASE_PER_TRANSACTION"
	TaxPercentageSettingsName                         = "TAX_PERCENTAGE"
	AdminFeePercentageSettingsName                    = "ADMIN_FEE_PERCENTAGE"

	// Not implemented yet in phase 1
	AdminFeePriceSettingsName = "ADMIN_FEE_PRICE"
)

const (
	SettingsTypeString  = "STRING"
	SettingsTypeBoolean = "BOOLEAN"
	SettingsTypeInteger = "INTEGER"
	SettingsTypeFlow    = "FLOAT"
)

const (
	SettingsValueBooleanTrue  = "true"
	SettingsValueBooleanFalse = "false"
)

func MapEventSettings(settings []entity.EventSetting) dto.EventSettings {
	var res dto.EventSettings
	for _, val := range settings {
		switch val.Setting.Name {
		case EventGarudaIdVerificationSettingName:
			res.GarudaIdVerification = val.SettingValue == SettingsValueBooleanTrue
		case EventPurchaseAdultTicketPerTransactionSettingName:
			ticketInt, err := strconv.Atoi(val.SettingValue)
			if err != nil {
				log.Warn().Str("Key", EventPurchaseAdultTicketPerTransactionSettingName).Str("Value", val.SettingValue).Msg("failed to cast settings value")
				defaultTicketInt, _ := strconv.Atoi(val.Setting.DefaultValue)
				res.MaxAdultTicketPerTransaction = defaultTicketInt
			} else {
				res.MaxAdultTicketPerTransaction = ticketInt
			}
		case TaxPercentageSettingsName:
			taxPercentage, err := strconv.ParseFloat(val.SettingValue, 32)
			if err != nil {
				log.Warn().Str("Key", TaxPercentageSettingsName).Str("Value", val.SettingValue).Msg("failed to cast settings value")
				defaultTaxPercentage, _ := strconv.ParseFloat(val.Setting.DefaultValue, 32)
				res.TaxPercentage = defaultTaxPercentage
			} else {
				res.TaxPercentage = taxPercentage
			}
		case AdminFeePriceSettingsName:
			adminFeePrice, err := strconv.Atoi(val.SettingValue)
			if err != nil {
				log.Warn().Str("Key", AdminFeePriceSettingsName).Str("Value", val.SettingValue).Msg("failed to cast settings value")
				defaultAdminFee, _ := strconv.Atoi(val.Setting.DefaultValue)
				res.AdminFee = defaultAdminFee
			} else {
				res.AdminFee = adminFeePrice
			}
		case AdminFeePercentageSettingsName:
			adminFeePercentage, err := strconv.ParseFloat(val.SettingValue, 32)
			if err != nil {
				log.Warn().Str("Key", AdminFeePercentageSettingsName).Str("Value", val.SettingValue).Msg("failed to cast settings value")
				defaultTaxPercentage, _ := strconv.ParseFloat(val.Setting.DefaultValue, 32)
				res.AdminFeePercentage = defaultTaxPercentage
			} else {
				res.AdminFeePercentage = adminFeePercentage
			}
		}
	}

	return res
}
