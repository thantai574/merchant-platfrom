package aggregates

import (
	"orders-system/domain/entities"
	"orders-system/proto/service_user"
)

type UserDetail struct {
	entities.User
	Balances            []entities.Balances        `json:"balances"`
	AccessToken         string                     `json:"access_token" bson:"-"`
	TokenType           string                     `json:"token_type" bson:"-"`
	RefreshToken        string                     `json:"refresh_token" bson:"-"`
	PinCodeOld          string                     `json:"-" bson:"pincode"`
	TransactionWrongOTP *TransactionWrongOTP       `json:"transaction_wrong_otp,omitempty"`
	LinksBank           []*entities.LinkedBankLink `json:"links_bank"`
}

// TransactionWrongOTP -
type TransactionWrongOTP struct {
	TransactionID string `json:"transaction_id"`
	Failed        bool   `json:"failed"`
}

func (UserDetail) ConvertUserDetailDTO2TypeUser(dto service_user.UserDetailDTO) *UserDetail {

	balances := []entities.Balances{}

	for _, v := range dto.Balances {
		balances = append(balances, entities.Balances{
			Id:              v.Id,
			AmountAvailable: v.AmountAvailable,
			AmountFreeze:    v.AmountFreeze,
			Type:            v.Type,
			UserId:          v.UserId,
			FreezeIds:       v.FreezeIds,
			Currency:        v.Currency,
		})
	}

	return &UserDetail{
		User: entities.User{
			Id:                           dto.User.Id,
			PhoneNumber:                  dto.User.PhoneNumber,
			Email:                        dto.User.Email,
			Name:                         dto.User.Name,
			Avatar:                       dto.User.Avatar,
			Gender:                       dto.User.Gender,
			Status:                       dto.User.Status,
			Timezone:                     dto.User.Timezone,
			Language:                     dto.User.Language,
			Title:                        dto.User.Title,
			DateTime:                     dto.User.DateTime,
			Currency:                     dto.User.Currency,
			Role:                         dto.User.Role,
			Password:                     dto.User.Password,
			IdentityNumber:               dto.User.IdentityNumber,
			BirthDay:                     dto.User.BirthDay,
			Address:                      dto.User.Address,
			MessageTopics:                dto.User.MessageTopics,
			DeviceIdLogin:                dto.User.DeviceIdLogin,
			LastDeviceIdLogin:            dto.User.LastDeviceIdLogin,
			OriginDevice:                 dto.User.OriginDevice,
			Source:                       dto.User.Source,
			RandomString:                 dto.User.RandomString,
			LastAmountNetPrimaryWallet:   dto.User.LastAmountNetPrimaryWallet,
			LastAmountPrimaryWallet:      dto.User.LastAmountPrimaryWallet,
			IncrementAmountPrimaryWallet: dto.User.IncrementAmountPrimaryWallet,
			DecrementAmountPrimaryWallet: dto.User.DecrementAmountPrimaryWallet,
			HasEverLinkedBank:            dto.User.HasEverLinkedBank,
			GapoId:                       dto.User.GapoId,
			GapoName:                     dto.User.GapoName,
			PinCode:                      dto.User.PinCode,
			PinCodeExpired:               dto.User.PinCodeExpired,
			LockTimeExpired:              dto.User.LockTimeExpired,
			WrongPinCode:                 dto.User.WrongPinCode,
			KYC:                          dto.User.Kyc,
		},
		Balances: balances,
	}
}
