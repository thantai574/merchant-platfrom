package entities

import "orders-system/domain/constants"

type LinkRequest struct {
	ClientCode   string          `json:"client_code"`
	GPayBankCode string          `json:"gpay_bank_code"`
	TransTime    int64           `json:"trans_time"`
	Data         LinkRequestData `json:"data"`
	IpAddress    string          `json:"ip_address"`
	Signature    string          `json:"signature"`
}

type LinkRequestData struct {
	GPayUserInfo struct {
		CustomerId  string `json:"customer_id"`
		Gender      string `json:"gender,omitempty"`
		PhoneNumber string `json:"phone_number"`
		FullName    string `json:"full_name,omitempty" `

		District  string `json:"district"`
		City      string `json:"city"`
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Address   string `json:"address"`
	} `json:"gpay_user_info"`
	GpayTransactionId string `json:"gpay_transaction_id"`
	GpayUserId        string `json:"gpay_user_id"`
	Channel           string `json:"channel,omitempty"`
	ReturnUrl         string `json:"return_url"`
	CancelUrl         string `json:"cancel_url"`
	Amount            int64  `json:"amount"`
	Description       string `json:"description"`
	CustomerCardInfo  struct {
		CardNumber     string `json:"card_number"`
		CardHolderName string `json:"card_holder_name"`
		Cvv            string `json:"cvv,omitempty"`
		IssueDate      string `json:"issue_date"`
		ExpireDate     string `json:"expire_date"`
	} `json:"customer_card_info"`
}

type LinkResponse struct {
	ErrorCode    constants.BankGwStatus `json:"error_code"`
	Message      string                 `json:"message"`
	Signature    string                 `json:"signature"`
	DataLinkResp interface{}            `json:"data"`
}

type LinkIdResponse struct {
	LinkID int64 `json:"link_id,omitempty"`
}

type NapasDetailResponse struct {
	ApiOperation string `json:"apiOperation"`
	Order        struct {
		Amount    string `json:"amount"`
		Id        string `json:"id"`
		Reference string `json:"reference"`
	} `json:"order"`
	DataKey         string `json:"dataKey"`
	NapasKey        string `json:"napasKey"`
	MerchantId      string `json:"merchant_id"`
	DeviceId        string `json:"device_id"`
	OrderNo         string `json:"order_no"`
	OrderRef        string `json:"order_ref"`
	ChannelRes      string `json:"channel_res"`
	ApiOperationRes string `json:"api_operation_res"`
	EnvironmentRes  string `json:"environment_res"`
	Amount          string `json:"amount"`
	LinkId          int64  `json:"link_id"`
}

// @todo initNAPAS CashIn
type NapasCashInRequest struct {
	ClientCode   string                 `json:"client_code"`
	GPayBankCode string                 `json:"gpay_bank_code"`
	TransTime    int64                  `json:"trans_time"`
	Data         NapasCashInDataRequest `json:"data,omitempty"`
	IpAddress    string                 `json:"ip_address"`
	Signature    string                 `json:"signature"`
}

type NapasCashInDataRequest struct {
	Amount            int64  `json:"amount"`
	GpayTransactionID string `json:"gpay_transaction_id"`
	LinkID            int64  `json:"link_id"`
	Channel           string `json:"channel"`
	GPayUserID        string `json:"gpay_user_id"`
	Description       string `json:"description"`
}

type NapasCashInResponse struct {
	ErrorCode constants.BankGwStatus `json:"error_code"`
	Data      interface{}            `json:"data"`
	Message   string                 `json:"message"`
	Signature string                 `json:"signature"`
}
