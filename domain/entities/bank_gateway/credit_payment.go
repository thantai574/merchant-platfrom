package entities

import "orders-system/domain/constants"

const GpayBankCodeCreditPayment = "GPMPGS"

type CreditPaymentRequest struct {
	ClientCode   string                   `json:"client_code"`
	GPayBankCode string                   `json:"gpay_bank_code"`
	TransTime    int64                    `json:"trans_time"`
	Data         CreditPaymentRequestData `json:"data"`
	IpAddress    string                   `json:"ip_address"`
	Signature    string                   `json:"signature"`
}

type CreditPaymentRequestData struct {
	CardNumber        string `json:"card_number"`
	GpayTransactionId string `json:"gpay_transaction_id"`
	Amount            string `json:"amount"`
	ExpiryYear        string `json:"expiry_year"`
	ExpiryMonth       string `json:"expiry_month"`
	SecurityCode      string `json:"security_code"`
	CardHolderName    string `json:"card_holder_name"`
	RedirectUrl       string `json:"redirect_url"`
	GpayUserId        string `json:"gpay_user_id"`
	Token             string `json:"token"`
	Currency          string `json:"currency"`
	MerchantCode      string `json:"merchant_code"`
	MCC               string `json:"mcc"`
	MccType           string `json:"mcc_type,omitempty"`
}

type CreditPaymentResponse struct {
	ErrorCode constants.BankGwStatus      `json:"error_code"`
	Data      CreditPaymentDetailResponse `json:"data,omitempty"`
	Message   string                      `json:"message,omitempty"`
	Signature string                      `json:"signature"`
}

type CreditPaymentDetailResponse struct {
	TripleDSecure struct {
		AuthenticationRedirect struct {
			Customized struct {
				AcsUrl string `json:"acsUrl"`
				PaReq  string `json:"paReq"`
			} `json:"customized"`
		} `json:"authenticationRedirect"`
	} `json:"3DSecure"`
	Md          string `json:"md"`
	CallbackUrl string `json:"callback_url"`
}

type RetrieveCreditOrderResponse struct {
	ErrorCode constants.BankGwStatus    `json:"error_code"`
	Data      RetrieveOrderDataResponse `json:"data,omitempty"`
	Message   string                    `json:"message,omitempty"`
	Signature string
}

type RetrieveOrderDataResponse struct {
	GpayUserId        string `json:"gpay_user_id,omitempty"`
	Amount            string `json:"amount" `
	BankTransactionId string `json:"bank_transaction_id,omitempty"`
	BankCode          string `json:"bank_code,omitempty"`
	GpayBankCode      string `json:"gpay_bank_code,omitempty"`
}
