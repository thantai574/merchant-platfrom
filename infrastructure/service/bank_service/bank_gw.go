package bank_service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"orders-system/domain/constants"
	entities "orders-system/domain/entities/bank_gateway"
	"orders-system/utils/helpers"
	"time"

	"github.com/spf13/cast"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const timeout = time.Minute*3 + time.Second*5

type repoImpl struct {
	Uri        string
	ClientCode string
	Logger     *zap.Logger
}

func (r repoImpl) VerifyOTP(gpayBankCode, refBankTraceID, orderId string, linkId int64, otp string) (response entities.VerifyOTPResponse, err error) {
	request := entities.VerifyOTPRequest{
		ClientCode:   r.ClientCode,
		GPayBankCode: gpayBankCode,
		TransTime:    time.Now().Unix() * 1000,
		Data: entities.VerifyOTPRequestData{
			RefBankTraceID:    refBankTraceID,
			LinkID:            linkId,
			OTP:               otp,
			Channel:           "MOBILE",
			Description:       "verify OTP",
			GpayTransactionId: orderId,
		},
		IpAddress: "ip",
	}

	err = r.httpRequest(struct {
		Path     string
		Method   string
		Headers  map[string]string
		Body     interface{}
		Response interface{}
	}{
		Path:     "verifyotp",
		Method:   "POST",
		Headers:  nil,
		Body:     request,
		Response: &response,
	})

	if err != nil {
		return response, err
	}

	if !response.ErrorCode.IsVerifying() && !response.ErrorCode.IsSuccess() && !response.ErrorCode.IsWrongOTP() {
		return response, errors.New(response.Message)
	}

	return response, err
}

func (r repoImpl) Link(data entities.LinkRequestData, gpBankCode string, ip string) (response entities.LinkResponse, err error) {
	dataSend := entities.LinkRequest{
		ClientCode:   "GPAYWEB",
		GPayBankCode: gpBankCode,
		TransTime:    time.Now().Unix() * 1000,
		Data:         data,
		IpAddress:    ip,
	}

	if !helpers.IsStringSliceContains([]string{"NAPAS"}, gpBankCode) {
		response.DataLinkResp = entities.LinkIdResponse{}
	}

	err = r.httpRequest(struct {
		Path     string
		Method   string
		Headers  map[string]string
		Body     interface{}
		Response interface{}
	}{
		Path:     "link",
		Method:   "POST",
		Headers:  nil,
		Body:     dataSend,
		Response: &response,
	})

	if err != nil {
		return response, err
	}

	if !response.ErrorCode.IsSuccess() && !response.ErrorCode.IsNeedToEnterOTP() {
		return response, errors.New(response.Message)
	}

	return response, err
}

func (r repoImpl) VerifyOTPLink(data entities.LinkRequestData, gpBankCode string, ip string) (response entities.LinkResponse, err error) {
	dataSend := entities.LinkRequest{
		ClientCode:   "GPAYWEB",
		GPayBankCode: gpBankCode,
		TransTime:    time.Now().Unix() * 1000,
		Data:         data,
		IpAddress:    ip,
	}

	err = r.httpRequest(struct {
		Path     string
		Method   string
		Headers  map[string]string
		Body     interface{}
		Response interface{}
	}{
		Path:     "link",
		Method:   "POST",
		Headers:  nil,
		Body:     dataSend,
		Response: &response,
	})

	if err != nil {
		return response, err
	}

	if !response.ErrorCode.IsSuccess() {
		return response, errors.New(response.Message)
	}

	return response, err
}

func (r repoImpl) UnLink(bankCode string, linkId int64) (response entities.UnLinkResponse, err error) {
	data := entities.UnLinkRequest{
		ClientCode:   r.ClientCode,
		GPayBankCode: bankCode,
		TransTime:    time.Now().Unix() * 1000,
		Data: entities.UnLinkRequestData{
			LinkID:      linkId,
			Channel:     "MOBILE",
			Description: "",
		},
		IpAddress: "ip",
	}

	err = r.httpRequest(struct {
		Path     string
		Method   string
		Headers  map[string]string
		Body     interface{}
		Response interface{}
	}{
		Path:     "unlink",
		Method:   "POST",
		Headers:  nil,
		Body:     data,
		Response: &response,
	})

	if err != nil {
		return response, err
	}

	if !response.ErrorCode.IsSuccess() {
		return response, errors.New(response.Message)
	}

	return response, err
}

//@todo rút từ bank -> nạp vào ví
func (r repoImpl) CashIn(gpayBankCode string, linkId int64, amount int64, gpayOrderId, userId string) (response entities.CashInResponse, err error) {
	data := entities.CashInRequest{
		ClientCode:   r.ClientCode,
		GPayBankCode: gpayBankCode,
		TransTime:    time.Now().Unix() * 1000,
		Data: entities.CashInDataRequest{
			Amount:            amount,
			GpayTransactionID: gpayOrderId,
			LinkID:            linkId,
			Channel:           "MOBILE",
			GPayUserID:        userId,
			Description:       "cash in",
		},
		IpAddress: "ip",
	}

	err = r.httpRequest(struct {
		Path     string
		Method   string
		Headers  map[string]string
		Body     interface{}
		Response interface{}
	}{
		Path:     "cashin",
		Method:   "POST",
		Headers:  nil,
		Body:     data,
		Response: &response,
	})

	if err != nil {
		return response, err
	}
	if !response.ErrorCode.IsNeedToEnterOTP() && !response.ErrorCode.IsSuccess() {
		return response, errors.New(response.Message)
	}

	return response, err
}

//@todo Napas Init CashIn
func (r repoImpl) CashInNapas(data entities.NapasCashInDataRequest, gpayBankCode string, ip string) (response entities.NapasCashInResponse, err error) {
	dataSend := entities.NapasCashInRequest{
		ClientCode:   "GPAYWEB",
		GPayBankCode: gpayBankCode,
		TransTime:    time.Now().Unix() * 1000,
		Data:         data,
		IpAddress:    ip,
	}

	err = r.httpRequest(struct {
		Path     string
		Method   string
		Headers  map[string]string
		Body     interface{}
		Response interface{}
	}{
		Path:     "cashin",
		Method:   "POST",
		Headers:  nil,
		Body:     dataSend,
		Response: &response,
	})

	if err != nil {
		return response, err
	}

	if response.ErrorCode.IsFail() {
		return response, errors.New(response.Message)
	}

	return response, err
}

func (r repoImpl) LinkInfo(linkId int64) (response entities.LinkInfoResponse, err error) {
	data := entities.LinkInfoRequest{
		ClientCode: r.ClientCode,
		TransTime:  time.Now().Unix() * 1000,
		Data: entities.LinkInfoRequestData{
			LinkID: linkId,
		},
		IpAddress: "ip",
	}

	err = r.httpRequest(struct {
		Path     string
		Method   string
		Headers  map[string]string
		Body     interface{}
		Response interface{}
	}{
		Path:     "linkinfo",
		Method:   "POST",
		Headers:  nil,
		Body:     data,
		Response: &response,
	})
	if err != nil {
		return response, err
	}

	if response.ErrorCode.IsFail() {
		return response, errors.New(response.Message)
	}

	return response, err
}

func (r repoImpl) LinkList(gpayUserId string) (response entities.ListLinkResponse, err error) {
	data := entities.ListLinkRequest{
		ClientCode:   r.ClientCode,
		GPayBankCode: "B",
		TransTime:    time.Now().Unix() * 1000,
		Data: entities.ListLinkRequestData{
			GpayUserID:  gpayUserId,
			Description: "get linked list",
		},
		IpAddress: "ip",
	}

	err = r.httpRequest(struct {
		Path     string
		Method   string
		Headers  map[string]string
		Body     interface{}
		Response interface{}
	}{
		Path:     "linklist",
		Method:   "POST",
		Headers:  nil,
		Body:     data,
		Response: &response,
	})
	if err != nil {
		return entities.ListLinkResponse{}, err
	}
	if response.ErrorCode.IsFail() {
		return entities.ListLinkResponse{}, errors.New(response.Message)
	}

	return response, err
}

//@todo rút tiền từ ví -> vào tài khoản ngân hàng
func (r repoImpl) CashOut(amount int64, bankCode, orderId string, linkId int64, description, userId string) (response entities.CashOutResponse, err error) {
	data := entities.CashOutRequest{
		ClientCode:   r.ClientCode,
		GPayBankCode: bankCode,
		TransTime:    time.Now().Unix() * 1000,
		Data: entities.CashOutRequestData{
			Amount:            amount,
			GpayTransactionID: orderId,
			LinkID:            linkId,
			Channel:           "MOBILE",
			GpayUserID:        userId,
			Description:       "Cashout",
		},
		IpAddress: "ip",
	}

	err = r.httpRequest(struct {
		Path     string
		Method   string
		Headers  map[string]string
		Body     interface{}
		Response interface{}
	}{
		Path:     "cashout",
		Method:   "POST",
		Headers:  nil,
		Body:     data,
		Response: &response,
	})

	if err != nil {
		return response, err
	}

	if !response.ErrorCode.IsSuccess() && !response.ErrorCode.IsVerifying() {
		return response, errors.New(response.Message)
	}

	return response, err
}

//@todo check tên người nhận IBFT
func (r repoImpl) IBFTInquiry(accountNumber, cardNumber, ibftCode string) (response entities.IBFTInquiryCheckResponse, err error) {
	dataRequest := entities.IBFTInquiryCheckRequest{
		ClientCode:   r.ClientCode,
		GPayBankCode: constants.GPAY_VCCB,
		TransTime:    time.Now().Unix() * 1000,
		Data: entities.IBFTInquiryCheckData{
			AccountNumber: accountNumber,
			CardNumber:    cardNumber,
			BankBin:       ibftCode,
			Description:   "",
		},
		IpAddress: "ip",
	}

	err = r.httpRequest(struct {
		Path     string
		Method   string
		Headers  map[string]string
		Body     interface{}
		Response interface{}
	}{
		Path:     "ibftinquiry",
		Method:   "POST",
		Headers:  nil,
		Body:     dataRequest,
		Response: &response,
	})

	if err != nil {
		return response, err
	}

	if !response.ErrorCode.IsSuccess() {
		return response, errors.New(response.Message)
	}

	return response, err
}

//@todo chuyển tiền IBFT
func (r repoImpl) IBFTTransfer(accountNumber, cardNumber, gpayUserId, orderId, ibftCode, description string, amount int64, bankName string) (response entities.IBFTTransferResponse, err error) {
	data := entities.IBFTTransferRequest{
		ClientCode:   r.ClientCode,
		GPayBankCode: constants.GPAY_VCCB,
		TransTime:    time.Now().Unix() * 1000,
		Data: entities.IBFTTransferReqData{
			AccountNumber:     accountNumber,
			CardNumber:        cardNumber,
			GpayUserID:        gpayUserId,
			GpayTransactionID: orderId,
			Amount:            amount,
			IBFTCode:          ibftCode,
			Description:       description,
			BankName:          bankName,
		},
		IpAddress: "ip",
	}

	err = r.httpRequest(struct {
		Path     string
		Method   string
		Headers  map[string]string
		Body     interface{}
		Response interface{}
	}{
		Path:     "ibftfundtransfer",
		Method:   "POST",
		Headers:  nil,
		Body:     data,
		Response: &response,
	})

	if err != nil {
		return entities.IBFTTransferResponse{}, err
	}

	if !response.ErrorCode.IsVerifying() && !response.ErrorCode.IsSuccess() {
		return response, errors.New(response.Message)
	}

	return response, err
}

// Kiểm tra số dư tkdb của Gpay tại ngân hàng
func (r repoImpl) CheckGpayBankBalance(gpayBankCode, accountNumber string) (response entities.BalanceCheckResponse, err error) {
	data := entities.BalanceCheckRequest{
		ClientCode:   "GPAYCORE",
		GPayBankCode: gpayBankCode,
		TransTime:    time.Now().Unix() * 1000,
		Data: entities.BalanceCheckRequestData{
			AccountNumber: accountNumber,
			Channel:       "MOBILE",
			Description:   "kiem tra tkdb gpay",
		},
		Signature: "",
	}

	err = r.httpRequest(struct {
		Path     string
		Method   string
		Headers  map[string]string
		Body     interface{}
		Response interface{}
	}{
		Path:     "ibftbalance",
		Method:   "POST",
		Headers:  nil,
		Body:     data,
		Response: &response,
	})

	return response, err
}

//@todo check số dư của GPAY IBFT
func (r repoImpl) IBFTCheckBalance(accountNumber string) (response entities.IBFTGpayBalanceCheckResponse, err error) {
	data := entities.IBFTGpayBalanceCheckRequest{
		ClientCode:   r.ClientCode,
		GPayBankCode: "GPNAPASVCCB",
		TransTime:    time.Now().Unix() * 1000,
		Data: entities.IBFTGpayBalanceCheckData{
			AccountNumber: accountNumber,
			Channel:       "MOBILE",
			Description:   "check tai khoan dam bao gpay IBFT",
		},
		Signature: "",
	}

	err = r.httpRequest(struct {
		Path     string
		Method   string
		Headers  map[string]string
		Body     interface{}
		Response interface{}
	}{
		Path:     "balance",
		Method:   "POST",
		Headers:  nil,
		Body:     data,
		Response: &response,
	})

	if err != nil {
		return response, err
	}

	return response, err
}

func (r repoImpl) CreditPayment(dataReq entities.CreditPaymentRequestData) (response entities.CreditPaymentResponse, err error) {
	data := entities.CreditPaymentRequest{
		ClientCode:   "GPAYCORE",
		GPayBankCode: entities.GpayBankCodeCreditPayment,
		TransTime:    time.Now().Unix() * 1000,
		Data:         dataReq,
		IpAddress:    "ip",
	}

	err = r.httpRequest(struct {
		Path     string
		Method   string
		Headers  map[string]string
		Body     interface{}
		Response interface{}
	}{
		Path:     "credit/payment3ds",
		Method:   "POST",
		Headers:  nil,
		Body:     data,
		Response: &response,
	})

	if err != nil {
		return entities.CreditPaymentResponse{}, err
	}
	if !response.ErrorCode.IsVerifying() && !response.ErrorCode.IsSuccess() {
		return entities.CreditPaymentResponse{}, errors.New(response.Message)
	}

	return response, err
}

func (r repoImpl) RetrieveOrderStatus(orderId string, gPayBankCode string) (response entities.RetrieveCreditOrderResponse, err error) {
	data := entities.CreditPaymentRequest{
		ClientCode:   "GPAYCORE",
		GPayBankCode: gPayBankCode,
		TransTime:    time.Now().Unix() * 1000,
		Data: entities.CreditPaymentRequestData{
			GpayTransactionId: orderId,
		},
		IpAddress: "ip",
	}

	err = r.httpRequest(struct {
		Path     string
		Method   string
		Headers  map[string]string
		Body     interface{}
		Response interface{}
	}{
		Path:     "retrieve",
		Method:   "POST",
		Headers:  nil,
		Body:     data,
		Response: &response,
	})

	if err != nil {
		return entities.RetrieveCreditOrderResponse{}, err
	}

	if !response.ErrorCode.IsSuccess() && !response.ErrorCode.IsVerifying() && !response.ErrorCode.IsFail() {
		return entities.RetrieveCreditOrderResponse{}, errors.New(response.Message)
	}

	return response, err
}
func (r repoImpl) ReOpenVA(reqData entities.ReOpenVARequestData, provider string) (response entities.ActionVAResponse, err error) {
	data := entities.ActionVARequest{
		ClientCode:   "GPAYCORE",
		GPayBankCode: provider,
		TransTime:    time.Now().Unix() * 1000,
		Data:         reqData,
		IpAddress:    "ip",
	}

	err = r.httpRequest(struct {
		Path     string
		Method   string
		Headers  map[string]string
		Body     interface{}
		Response interface{}
	}{
		Path:     "va/reopen",
		Method:   "POST",
		Headers:  nil,
		Body:     data,
		Response: &response,
	})

	if err != nil {
		return entities.ActionVAResponse{}, err
	}
	if !response.ErrorCode.IsSuccess() {
		return entities.ActionVAResponse{}, errors.New(response.Message)
	}

	return response, err
}

func (r repoImpl) CreateVA(req entities.CreateVARequestData, provider string) (response entities.ActionVAResponse, err error) {
	data := entities.ActionVARequest{
		ClientCode:   "GPAYCORE",
		GPayBankCode: provider,
		TransTime:    time.Now().Unix() * 1000,
		Data:         req,
		IpAddress:    "ip",
	}

	b, err := json.Marshal(data)
	fmt.Println(string(b))

	err = r.httpRequest(struct {
		Path     string
		Method   string
		Headers  map[string]string
		Body     interface{}
		Response interface{}
	}{
		Path:     "va/create",
		Method:   "POST",
		Headers:  nil,
		Body:     data,
		Response: &response,
	})

	if err != nil {
		return entities.ActionVAResponse{}, err
	}
	if !response.ErrorCode.IsSuccess() {
		return entities.ActionVAResponse{}, errors.New(response.Message)
	}

	return response, err
}

func (r repoImpl) UpdateVA(reqData entities.CreateVARequestData, provider string) (response entities.ActionVAResponse, err error) {
	data := entities.ActionVARequest{
		ClientCode:   "GPAYCORE",
		GPayBankCode: provider,
		TransTime:    time.Now().Unix() * 1000,
		Data:         reqData,
		IpAddress:    "ip",
	}

	err = r.httpRequest(struct {
		Path     string
		Method   string
		Headers  map[string]string
		Body     interface{}
		Response interface{}
	}{
		Path:     "va/update",
		Method:   "POST",
		Headers:  nil,
		Body:     data,
		Response: &response,
	})

	if err != nil {
		return entities.ActionVAResponse{}, err
	}
	if !response.ErrorCode.IsSuccess() {
		return entities.ActionVAResponse{}, errors.New(response.Message)
	}

	return response, err
}

func (r repoImpl) CloseVA(reqData entities.CloseVARequestData, provider string) (response entities.ActionVAResponse, err error) {
	data := entities.ActionVARequest{
		ClientCode:   "GPAYCORE",
		GPayBankCode: provider,
		TransTime:    time.Now().Unix() * 1000,
		Data:         reqData,
		IpAddress:    "ip",
	}

	err = r.httpRequest(struct {
		Path     string
		Method   string
		Headers  map[string]string
		Body     interface{}
		Response interface{}
	}{
		Path:     "va/close",
		Method:   "POST",
		Headers:  nil,
		Body:     data,
		Response: &response,
	})

	if err != nil {
		return entities.ActionVAResponse{}, err
	}
	if !response.ErrorCode.IsSuccess() {
		return entities.ActionVAResponse{}, errors.New(response.Message)
	}

	return response, err
}

func (r repoImpl) DetailVA(accountNumber string, provider string) (response entities.ActionVAResponse, err error) {
	data := entities.ActionVARequest{
		ClientCode:   "GPAYCORE",
		GPayBankCode: provider,
		TransTime:    time.Now().Unix() * 1000,
		Data: entities.DetailVARequestData{
			AccountNumber: accountNumber,
		},
		IpAddress: "ip",
	}

	err = r.httpRequest(struct {
		Path     string
		Method   string
		Headers  map[string]string
		Body     interface{}
		Response interface{}
	}{
		Path:     "va/detail",
		Method:   "POST",
		Headers:  nil,
		Body:     data,
		Response: &response,
	})

	if err != nil {
		return entities.ActionVAResponse{}, err
	}
	if !response.ErrorCode.IsSuccess() {
		return entities.ActionVAResponse{}, errors.New(response.Message)
	}

	return response, err
}

func (r repoImpl) CheckInternationalBankBin(bankbin string) (response entities.CheckBankBinResponse, err error) {
	type CheckBankBin struct {
		BankBin string `json:"bank_bin"`
		Size    int    `json:"size"`
		Status  string `json:"status"`
	}

	data := CheckBankBin{
		BankBin: bankbin,
		Size:    1,
		Status:  "ACTIVE",
	}

	err = r.httpRequest(struct {
		Path     string
		Method   string
		Headers  map[string]string
		Body     interface{}
		Response interface{}
	}{
		Path:     "credit/whitelist/bankbin/search",
		Method:   "POST",
		Headers:  nil,
		Body:     data,
		Response: &response,
	})

	if err != nil {
		return entities.CheckBankBinResponse{}, err
	}
	if !response.ErrorCode.IsSuccess() {
		return entities.CheckBankBinResponse{}, errors.New(response.Message)
	}

	return response, err
}

func (r repoImpl) PGInitTrans(bankCode string, data entities.PGInitTransReqData) (response entities.PGInitTransRes, err error) {
	body := entities.PGInitTransReq{
		ClientCode:   r.ClientCode,
		GPayBankCode: bankCode,
		TransTime:    time.Now().Unix() * 1000,
		Data:         data,
		IpAddress:    "ip",
	}

	err = r.httpRequest(struct {
		Path     string
		Method   string
		Headers  map[string]string
		Body     interface{}
		Response interface{}
	}{
		Path:     "payment-gateway/init-trans",
		Method:   "POST",
		Headers:  nil,
		Body:     body,
		Response: &response,
	})

	if err != nil {
		return entities.PGInitTransRes{}, err
	}
	if !response.ErrorCode.IsSuccess() {
		return entities.PGInitTransRes{}, errors.New(response.Message)
	}

	return response, err
}

func (r repoImpl) PGInitOrder(req entities.PGInitTransReq) (response entities.PGInitOrderRes, err error) {
	req.ClientCode = r.ClientCode
	req.TransTime = time.Now().Unix() * 1000

	err = r.httpRequest(struct {
		Path     string
		Method   string
		Headers  map[string]string
		Body     interface{}
		Response interface{}
	}{
		Path:     "payment-gateway/init-trans",
		Method:   "POST",
		Headers:  nil,
		Body:     req,
		Response: &response,
	})

	if err != nil {
		return entities.PGInitOrderRes{}, err
	}
	if !response.ErrorCode.IsSuccess() {
		return entities.PGInitOrderRes{}, errors.New(response.Message)
	}

	return response, err
}

func (r repoImpl) ReFund(gpayTransactionId string, amount int64) (response entities.RefundRes, err error) {
	data := entities.RefundReqData{
		GPayTransactionID: gpayTransactionId,
		Amount:            cast.ToString(amount),
	}

	body := entities.RefundReq{
		ClientCode:   r.ClientCode,
		GPayBankCode: "gpayBankCode",
		TransTime:    time.Now().Unix() * 1000,
		Data:         data,
		IpAddress:    "ip",
	}

	err = r.httpRequest(struct {
		Path     string
		Method   string
		Headers  map[string]string
		Body     interface{}
		Response interface{}
	}{
		Path:     "refund",
		Method:   "POST",
		Headers:  nil,
		Body:     body,
		Response: &response,
	})

	if err != nil {
		return entities.RefundRes{}, err
	}
	if !response.ErrorCode.IsSuccess() {
		return entities.RefundRes{}, errors.New(response.Message)
	}

	return response, err
}

func (r repoImpl) CheckByPassOTP(req entities.CheckByPassOTPDataReq) (response entities.CheckByPassOTPDataRes, err error) {
	err = r.httpRequest(struct {
		Path     string
		Method   string
		Headers  map[string]string
		Body     interface{}
		Response interface{}
	}{
		Path:     "otpLimit/verify",
		Method:   "POST",
		Headers:  nil,
		Body:     req,
		Response: &response,
	})

	if err != nil {
		return entities.CheckByPassOTPDataRes{}, err
	}
	if !response.ErrorCode.IsSuccess() && !response.ErrorCode.IsNeedToEnterOTP() {
		if response.Message == "" {
			return entities.CheckByPassOTPDataRes{}, errors.New("Đã có lỗi xảy ra , vui lòng thử lại sau")
		}
		return entities.CheckByPassOTPDataRes{}, errors.New(response.Message)
	}

	return response, err
}

func (r repoImpl) httpRequest(request struct {
	Path     string
	Method   string
	Headers  map[string]string
	Body     interface{}
	Response interface{}
}) (err error) {
	client := new(http.Client)

	client.Timeout = timeout

	jsonrequest, err := json.Marshal(request.Body)
	r.Logger.With(zapcore.Field{
		Key:       "request",
		Type:      zapcore.StringType,
		String:    fmt.Sprintf("%v", string(jsonrequest)),
		Interface: nil,
	}).Info("bank_request")
	req, err := http.NewRequest(request.Method, fmt.Sprintf("%v%v", r.Uri, request.Path), bytes.NewReader(jsonrequest))

	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", `application/json`)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == 500 {
		responseByte, _ := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		r.Logger.Error("BANK GATEWAY SERVER ERROR: " + string(responseByte))
		return errors.New("BANK GATEWAY SERVER ERROR")
	}

	responseByte, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	r.Logger.With(zapcore.Field{
		Key:       "uri",
		Type:      zapcore.StringType,
		String:    fmt.Sprintf("%v%v", r.Uri, request.Path),
		Interface: nil,
	}).With(zapcore.Field{
		Key:       "request",
		Type:      zapcore.StringType,
		String:    string(jsonrequest),
		Interface: nil,
	}).With(
		zapcore.Field{
			Key:       "response",
			Type:      zapcore.StringType,
			String:    string(responseByte),
			Interface: nil,
		}).Info("http_request_data")

	err = json.Unmarshal(responseByte, request.Response)
	if err != nil {
		r.Logger.With(zap.Error(err)).Error("can not unmarshal response")
		return err
	}
	//Close Request
	defer func() {
		err = resp.Body.Close()
	}()

	return err
}

func (r repoImpl) GetBanks() (response entities.BankRes, err error) {
	err = r.httpRequest(struct {
		Path     string
		Method   string
		Headers  map[string]string
		Body     interface{}
		Response interface{}
	}{
		Path:     "bank/active",
		Method:   "GET",
		Headers:  nil,
		Response: &response,
	})

	if err != nil {
		return
	}
	if !response.ErrorCode.IsSuccess() {
		err = errors.New(response.Message)
		return
	}

	return response, err
}

func NewRepoImpl(uri string, logger *zap.Logger) *repoImpl {
	return &repoImpl{
		Uri:        uri,
		ClientCode: "GPAYCORE",
		Logger:     logger,
	}
}
