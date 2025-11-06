package application

import (
	"context"
	"errors"
	"fmt"
	"github.com/spf13/cast"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"orders-system/domain/constants"
	"orders-system/domain/entities"
	eBankGw "orders-system/domain/entities/bank_gateway"
	"orders-system/domain/request_params"
	"orders-system/domain/value_objects"
	"orders-system/proto/order_system"
	"orders-system/proto/service_merchant_fee"
	"orders-system/utils/configs"
	"orders-system/utils/gpooling"
	"orders-system/utils/helpers"
	logger2 "orders-system/utils/logger"
)

type iVa interface {
	CreateVA(context.Context, *order_system.CreateVARequest, *order_system.CreateVAResponse) error
	UpdateVa(context.Context, *order_system.UpdateVARequest, *order_system.UpdateVAResponse) error
	CloseVa(context.Context, *order_system.CloseVARequest, *order_system.CloseVAResponse) error
	GetDetailVA(context.Context, *order_system.DetailVARequest, *order_system.DetailVAResponse) error
	ReOpenVa(context.Context, *order_system.ReOpenVARequest, *order_system.ReOpenVAResponse) error
}

type VAStrategy struct {
	iVa
}

type MSBVa struct {
	Name               string
	MerchantIdentifier string
	OrderApplication
}

func (this *MSBVa) UpdateVa(ctx context.Context, request *order_system.UpdateVARequest, response *order_system.UpdateVAResponse) error {
	getVAByAccountName, err := this.IVA.GetVAByAccountNumber(request.AccountNumber)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("Account Number doesn't exist")
		}
		return err
	}

	if getVAByAccountName.MerchantCode != request.MerchantCode {
		return fmt.Errorf("Invalid Merchant %v with account number %v", request.MerchantCode, request.AccountNumber)
	}

	_, err = this.BankServiceRepository.UpdateVA(eBankGw.CreateVARequestData{
		AccountName:     request.GetAccountName(),
		AccountNumber:   request.GetAccountNumber(),
		ReferenceNumber: getVAByAccountName.MapId,
		Status:          "1", // active
		MaxAmount:       cast.ToString(request.MaxAmount),
		MinAmount:       cast.ToString(request.MinAmount),
		EqualAmount:     cast.ToString(request.EqualAmount),
		AccountType:     request.AccountType.String(),
	}, this.Name)
	if err != nil {
		return err
	}

	updateVaAccountNameReq := bson.M{
		"account_name":            request.AccountName,
		"updated_at":              helpers.GetCurrentTime(),
		"min_amount":              request.MinAmount,
		"max_amount":              request.MaxAmount,
		"equal_amount":            request.EqualAmount,
		"account_type":            request.AccountType.String(),
		"account_type_many_limit": request.GetManyTypeAccountLimit(),
	}
	updateVaAccountNameDB, err := this.IVA.UpdateVA(request.AccountNumber, updateVaAccountNameReq)
	if err != nil {
		return err
	}

	response.DetailVA = updateVaAccountNameDB.ConvertToProto()

	return err
}

func (this *MSBVa) CloseVa(ctx context.Context, request *order_system.CloseVARequest, response *order_system.CloseVAResponse) error {
	getVAByAccountByNumber, err := this.IVA.GetVAByAccountNumber(request.AccountNumber)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("Account Number doesn't exist")
		}
		return err
	}

	if getVAByAccountByNumber.MerchantCode != request.MerchantCode {
		return fmt.Errorf("Invalid Merchant %v with account number %v", request.MerchantCode, request.AccountNumber)
	}

	_, err = this.BankServiceRepository.UpdateVA(eBankGw.CreateVARequestData{
		AccountNumber:   request.GetAccountNumber(),
		ReferenceNumber: getVAByAccountByNumber.MapId,
		AccountName:     getVAByAccountByNumber.AccountName,
		Status:          "2", // any != 1
	}, this.Name)
	if err != nil {
		return err
	}

	updateField := bson.M{"status": "CLOSE", "close_reason": request.CloseReason,
		"expire_date": helpers.GetCurrentTime().Format("02-01-2006"), "updated_at": helpers.GetCurrentTime(), "closed_by": request.CloseBy}
	closeVARes, err := this.IVA.UpdateVA(request.AccountNumber, updateField)
	if err != nil {
		return err
	}

	_ = this.CreateMessageMqtt(context.TODO(), constants.TopicMQTTCloseVA, constants.MQTTEventBackground, constants.TopicMQTTCloseVA, closeVARes, false)

	response.DetailVA = closeVARes.ConvertToProto()
	return err
}

func (this *MSBVa) GetDetailVA(ctx context.Context, request *order_system.DetailVARequest, response *order_system.DetailVAResponse) error {
	getVAByAccountNumber, err := this.IVA.GetVAByAccountNumber(request.AccountNumber)
	if err != nil {
		return err
	}

	if getVAByAccountNumber.MerchantCode != request.MerchantCode {
		return fmt.Errorf("Invalid Merchant %v with account number %v", request.MerchantCode, request.AccountNumber)
	}

	response.DetailVA = getVAByAccountNumber.ConvertToProto()
	return err
}

func (this *MSBVa) ReOpenVa(ctx context.Context, request *order_system.ReOpenVARequest, response *order_system.ReOpenVAResponse) error {
	getVAByAccountNumber, err := this.IVA.GetVAByAccountNumber(request.AccountNumber)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("Account Number doesn't exist")
		}
		return err
	}

	if getVAByAccountNumber.MerchantCode != request.MerchantCode {
		return fmt.Errorf("Invalid Merchant %v with account number %v", request.MerchantCode, request.AccountNumber)
	}

	_, err = this.BankServiceRepository.UpdateVA(eBankGw.CreateVARequestData{
		AccountNumber:   request.GetAccountNumber(),
		Status:          "1",
		ReferenceNumber: getVAByAccountNumber.MapId,
		AccountName:     getVAByAccountNumber.AccountName,
	}, this.Name)
	if err != nil {
		return err
	}

	updateReOpenVAReq := bson.M{"status": "OPEN", "updated_at": helpers.GetCurrentTime(), "reopen_time": helpers.GetCurrentTime()}
	updateVaAccountDB, err := this.IVA.UpdateVA(request.AccountNumber, updateReOpenVAReq)
	if err != nil {
		return err
	}

	if updateVaAccountDB.AccountType == constants.VA_ACCOUNT_TYPE_MANY {
		_, err = this.MerchantFeeRepository.LogVAManagementFee(context.TODO(), &service_merchant_fee.LogVAManagementFeeReq{
			MerchantId: updateVaAccountDB.MerchantId,
			VaType:     service_merchant_fee.LogVAManagementFeeReq_RE_OPEN,
		})

		if err != nil {
			this.Logger.With(zap.Error(err)).Error("SERVICE_MERCHANT_FEE.error")
		}
	}

	response.DetailVA = updateVaAccountDB.ConvertToProto()
	return err
}

type VCCBVa struct {
	Name               string
	MerchantIdentifier string
	OrderApplication
}

func (this VCCBVa) UpdateVa(ctx context.Context, request *order_system.UpdateVARequest, response *order_system.UpdateVAResponse) error {
	getVAByAccountName, err := this.IVA.GetVAByAccountNumber(request.AccountNumber)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("Account Number doesn't exist")
		}
		return err
	}

	if getVAByAccountName.MerchantCode != request.MerchantCode {
		return fmt.Errorf("Invalid Merchant %v with account number %v", request.MerchantCode, request.AccountNumber)
	}

	updateVABankResp, err := this.BankServiceRepository.UpdateVA(eBankGw.CreateVARequestData{
		AccountName:   request.GetAccountName(),
		AccountNumber: request.GetAccountNumber(),
	}, this.Name)
	if err != nil {
		return err
	}

	updateVaAccountNameReq := bson.M{"account_name": updateVABankResp.Data.AccountName, "updated_at": helpers.GetCurrentTime()}
	updateVaAccountNameDB, err := this.IVA.UpdateVA(request.AccountNumber, updateVaAccountNameReq)
	if err != nil {
		return err
	}

	response.DetailVA = updateVaAccountNameDB.ConvertToProto()

	return err
}

func (this VCCBVa) CloseVa(ctx context.Context, request *order_system.CloseVARequest, response *order_system.CloseVAResponse) error {
	getVAByAccountByNumber, err := this.IVA.GetVAByAccountNumber(request.AccountNumber)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("Account Number doesn't exist")
		}
		return err
	}

	if getVAByAccountByNumber.MerchantCode != request.MerchantCode {
		return fmt.Errorf("Invalid Merchant %v with account number %v", request.MerchantCode, request.AccountNumber)
	}

	closeVABankResp, err := this.BankServiceRepository.CloseVA(eBankGw.CloseVARequestData{
		AccountNumber: request.GetAccountNumber(),
	}, this.Name)
	if err != nil {
		return err
	}

	updateField := bson.M{"status": closeVABankResp.Data.Status, "close_reason": request.CloseReason, "expire_date": closeVABankResp.Data.ExpireDate, "updated_at": helpers.GetCurrentTime(), "closed_by": request.CloseBy}
	closeVARes, err := this.IVA.UpdateVA(request.AccountNumber, updateField)
	if err != nil {
		return err
	}

	_ = this.CreateMessageMqtt(context.TODO(), constants.TopicMQTTCloseVA, constants.MQTTEventBackground, constants.TopicMQTTCloseVA, closeVARes, false)

	response.DetailVA = closeVARes.ConvertToProto()
	return err
}

func (this VCCBVa) GetDetailVA(ctx context.Context, request *order_system.DetailVARequest, response *order_system.DetailVAResponse) error {
	getVAByAccountNumber, err := this.IVA.GetVAByAccountNumber(request.AccountNumber)
	if err != nil {
		return err
	}

	if getVAByAccountNumber.MerchantCode != request.MerchantCode {
		return fmt.Errorf("Invalid Merchant %v with account number %v", request.MerchantCode, request.AccountNumber)
	}

	detailVABankResp, err := this.BankServiceRepository.DetailVA(request.AccountNumber, this.Name)
	if err != nil {
		return err
	}

	getVAByAccountNumber.AccountName = detailVABankResp.Data.AccountName
	getVAByAccountNumber.AccountNunmber = detailVABankResp.Data.AccountNumber
	getVAByAccountNumber.Status = detailVABankResp.Data.Status
	getVAByAccountNumber.ExpiredDate = detailVABankResp.Data.ExpireDate
	getVAByAccountNumber.AccountType = detailVABankResp.Data.AccountType
	response.DetailVA = getVAByAccountNumber.ConvertToProto()

	return err
}

func (this VCCBVa) ReOpenVa(ctx context.Context, request *order_system.ReOpenVARequest, response *order_system.ReOpenVAResponse) error {
	getVAByAccountNumber, err := this.IVA.GetVAByAccountNumber(request.AccountNumber)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("Account Number doesn't exist")
		}
		return err
	}

	if getVAByAccountNumber.MerchantCode != request.MerchantCode {
		return fmt.Errorf("Invalid Merchant %v with account number %v", request.MerchantCode, request.AccountNumber)
	}

	reOpenVaBankResp, err := this.BankServiceRepository.ReOpenVA(eBankGw.ReOpenVARequestData{
		AccountNumber: request.GetAccountNumber(),
	}, this.Name)
	if err != nil {
		return err
	}

	updateReOpenVAReq := bson.M{"status": reOpenVaBankResp.Data.Status, "updated_at": helpers.GetCurrentTime(), "reopen_time": helpers.GetCurrentTime()}
	updateVaAccountDB, err := this.IVA.UpdateVA(request.AccountNumber, updateReOpenVAReq)
	if err != nil {
		return err
	}

	if updateVaAccountDB.AccountType == constants.VA_ACCOUNT_TYPE_MANY {
		_, err = this.MerchantFeeRepository.LogVAManagementFee(context.TODO(), &service_merchant_fee.LogVAManagementFeeReq{
			MerchantId: updateVaAccountDB.MerchantId,
			VaType:     service_merchant_fee.LogVAManagementFeeReq_RE_OPEN,
		})

		if err != nil {
			this.Logger.With(zap.Error(err)).Error("SERVICE_MERCHANT_FEE.error")
		}
	}

	response.DetailVA = updateVaAccountDB.ConvertToProto()

	return err
}

func (this VCCBVa) CreateVA(ctx context.Context, request *order_system.CreateVARequest, response *order_system.CreateVAResponse) error {
	createVABankResponse, err := this.BankServiceRepository.CreateVA(eBankGw.CreateVARequestData{
		AccountName: request.GetAccountName(),
		AccountType: request.AccountType.String(),
	}, this.Name)
	if err != nil {
		return err
	}

	vaReq := entities.VirtualAccounts{
		Status:         createVABankResponse.Data.Status,
		Provider:       constants.VCCB,
		Balance:        0,
		AccountNunmber: createVABankResponse.Data.AccountNumber,
		AccountName:    createVABankResponse.Data.AccountName,
		AccountType:    request.AccountType.String(),
		BankCode:       constants.VCCB,
		CreatedAt:      helpers.GetCurrentTime(),
		UpdatedAt:      helpers.GetCurrentTime(),
		MerchantId:     request.MerchantId,
		MerchantCode:   request.MerchantCode,
		MapId:          request.MapId,
		MapType:        request.MapType,
		LimitAmount:    request.LimitAmount,
	}

	createVAResponseDb, err := this.IVA.CreateVA(vaReq)
	if err != nil {
		return err
	}

	if request.AccountType.String() == constants.VA_ACCOUNT_TYPE_MANY {
		_, err = this.MerchantFeeRepository.LogVAManagementFee(context.TODO(), &service_merchant_fee.LogVAManagementFeeReq{
			MerchantId: request.MerchantId,
			VaType:     service_merchant_fee.LogVAManagementFeeReq_NEW_OPEN,
		})

		if err != nil {
			this.Logger.With(zap.Error(err)).Error("SERVICE_MERCHANT_FEE.error")
		}
	}

	response.DetailVA = createVAResponseDb.ConvertToProto()
	return err
}

func (this *MSBVa) CreateVA(ctx context.Context, request *order_system.CreateVARequest, res *order_system.CreateVAResponse) error {
	vaReq := entities.VirtualAccounts{
		Status:               "OPEN",
		Provider:             constants.MSB,
		Balance:              0,
		AccountName:          request.AccountName,
		AccountType:          request.AccountType.String(),
		BankCode:             constants.MSB,
		CreatedAt:            helpers.GetCurrentTime(),
		UpdatedAt:            helpers.GetCurrentTime(),
		MerchantId:           request.MerchantId,
		MerchantCode:         request.MerchantCode,
		MapId:                request.MapId,
		MapType:              request.MapType,
		LimitAmount:          request.LimitAmount,
		IsAutoIncrement:      true,
		MerchantIdentifier:   this.MerchantIdentifier,
		ManyAccountTypeLimit: request.GetManyTypeAccountLimit(),
		MinAmount:            request.GetMinAmount(),
		MaxAmount:            request.GetMaxAmount(),
		EqualAmount:          request.GetEqualAmount(),
	}

	createVAResponseDb, err := this.IVA.CreateVA(vaReq)
	if err != nil {
		return err
	}

	_, err = this.BankServiceRepository.CreateVA(eBankGw.CreateVARequestData{
		AccountName:     createVAResponseDb.AccountName,
		AccountType:     createVAResponseDb.AccountType,
		AccountNumber:   createVAResponseDb.AccountNunmber,
		ReferenceNumber: createVAResponseDb.MapId,
		Status:          "1",
		MaxAmount:       cast.ToString(createVAResponseDb.MaxAmount),
		MinAmount:       cast.ToString(createVAResponseDb.MinAmount),
		EqualAmount:     cast.ToString(createVAResponseDb.EqualAmount),
	}, this.Name)
	if err != nil {
		errorName := err.Error()
		err := this.IVA.DeleteVAAccount(createVAResponseDb.AccountNunmber)
		if err != nil {
			return err
		}
		return errors.New(errorName)
	}

	if request.AccountType.String() == constants.VA_ACCOUNT_TYPE_MANY {
		_, err = this.MerchantFeeRepository.LogVAManagementFee(ctx, &service_merchant_fee.LogVAManagementFeeReq{
			MerchantId: request.MerchantId,
			VaType:     service_merchant_fee.LogVAManagementFeeReq_NEW_OPEN,
		})

		if err != nil {
			this.Logger.With(zap.Error(err)).Error("SERVICE_MERCHANT_FEE.error")
		}
	}

	res.DetailVA = createVAResponseDb.ConvertToProto()
	return err
}

func (usecase OrderApplication) ActionCreateVA(ctx context.Context, req *order_system.CreateVARequest, res *order_system.CreateVAResponse) error {
	return usecase.initVAStrategy(ctx, req.MerchantId).CreateVA(ctx, req, res)
}

func (usecase OrderApplication) ActionUpdateVA(ctx context.Context, req *order_system.UpdateVARequest, res *order_system.UpdateVAResponse) error {
	return usecase.initVAStrategy(ctx, req.MerchantId).UpdateVa(ctx, req, res)
}

func (usecase OrderApplication) ActionCloseVA(ctx context.Context, req *order_system.CloseVARequest, res *order_system.CloseVAResponse) error {
	return usecase.initVAStrategy(ctx, req.MerchantId).CloseVa(ctx, req, res)
}

func (usecase OrderApplication) ActionGetDetailVA(ctx context.Context, req *order_system.DetailVARequest, res *order_system.DetailVAResponse) error {
	return usecase.initVAStrategy(ctx, req.MerchantId).GetDetailVA(ctx, req, res)
}

func (usecase OrderApplication) ActionReOpenAccountVA(ctx context.Context, req *order_system.ReOpenVARequest, res *order_system.ReOpenVAResponse) error {
	return usecase.initVAStrategy(ctx, req.MerchantId).ReOpenVa(ctx, req, res)
}

func (us OrderApplication) initVAStrategy(ctx context.Context, merchantId string) *VAStrategy {
	config, err := configs.LoadConfig()
	if err != nil {
		panic(err)
	}
	lg, _ := logger2.NewLogger("production")

	poolGoRoutine, _ := gpooling.NewPooling(config.MaxPoolSize)
	application := NewOrderApplication(config, lg, poolGoRoutine)

	getConfig, err := us.IWalletConfig.GetMerchantConfig(ctx, request_params.GetMerchantConfigReq{
		MerchantId:   merchantId,
		ServiceType:  constants.SERVICE_TYPE_COLLECTION_AND_PAY,
		TransType:    constants.TRANSTYPE_PAY_VA,
		SubTransType: "",
	})
	if err != nil {
		panic("Không tìm thấy cấu hình")
	}

	var provider iVa
	var getSettingVAProvider value_objects.MerchantConfigSetting

	for _, v := range getConfig.Settings {
		if v.TransType == constants.TRANSTYPE_PAY_VA {
			getSettingVAProvider = v
			break
		}
	}

	var getProviderConfig = getSettingVAProvider.Provider
	switch getProviderConfig {
	case constants.VCCB:
		provider = VCCBVa{
			Name:               "GP" + getProviderConfig,
			OrderApplication:   *application,
			MerchantIdentifier: getSettingVAProvider.VACode,
		}
		break
	case constants.MSB:
		provider = &MSBVa{
			Name:               "GP" + getProviderConfig,
			MerchantIdentifier: getSettingVAProvider.VACode,
			OrderApplication:   *application,
		}
		break
	default:
		panic("Provider is not configured")
	}

	return &VAStrategy{
		provider,
	}
}
