package entities

type LinkDataRequest struct {
	ClientCode   string     `json:"clientCode"`
	GPayBankCode string     `json:"gpayBankCode"`
	TransTime    int64      `json:"transTime"`
	Data         DetailData `json:"data"`
	Signature    string     `json:"signature"`
}

type DetailData struct {
	CustomerCardInfo    CustomerCardInfoDetail    `json:"customerCardInfo"`
	CustomerAccountInfo CustomerAccountInfoDetail `json:"customerAccountInfo"`
	GpayUserInfo        GpayUserInfoDetail        `json:"gpayUserInfo"`
	GpayUserID          string                    `json:"gpayUserID"` // Id nguoi dung vi
	Channel             string                    `json:"channel"`    // kênh : MOBILE, WEB , POS , DESKTOP, SMS
	ReturnUrl           string                    `json:"returnUrl"`  // return về khi KH bấm submit
	CancelUrl           string                    `json:"cancelUrl"`  //  return khi KH bấm hủy
	Description         string                    `json:"description"`
}

type CustomerCardInfoDetail struct {
	CardNumber     string `json:"cardNumber"`     //số thẻ
	CardHolderName string `json:"cardHolderName"` // ten tren thẻ
	Cvv            string `json:"cvv"`
	IssueDate      string `json:"issueDate"`
	ExpireDate     string `json:"expireDate"`
}

type CustomerAccountInfoDetail struct {
	AccountNumber string `json:"accountNumber"`
	FullName      string `json:"fullName"`
}

type GpayUserInfoDetail struct {
	CustomerID          string `json:"customerID"` // cmt nhan dan/ ho chieu
	FullName            string `json:"fullName"  `
	Gender              string `json:"gender"`
	District            string `json:"district"`
	City                string `json:"city"`
	Country             string `json:"country"`
	Email               string `json:"email"`
	PhoneNumber         string `json:"phoneNumber"`
	IpAddress           string `json:"IpAddress"`
	Dob                 string `json:"dob"` // ngày sinh (format YYYY-MM-DD)
	CustomerIDIssueDate string `json:"customerIDIssueDate"`
	FirstName           string `json:"firstName"`
	LastName            string `json:"lastName"`
	Address             string `json:"address"`
}

type DataLinkResponse struct {
	ErrorCode string         `json:"errorCode"`
	Data      DataLinkDetail `json:"data"`
	Message   string         `json:"message"`
	Signature string         `json:"signature"`
}

type DataLinkDetail struct {
	LinkID        string       `json:"linkID,omitempty"`
	CardNumber    string       `json:"cardNumber,omitempty"`
	AccountNumber string       `json:"accountNumber,omitempty"`
	FullName      string       `json:"fullName,omitempty"`
	GpayUserID    string       `json:"gpayUserID,omitempty"`
	PhoneNumber   string       `json:"phoneNumber,omitempty"`
	Provider      string       `json:"provider,omitempty"`
	BankCode      string       `json:"bankCode,omitempty"`
	Status        string       `json:"status,omitempty"`
	RedirectUrl   string       `json:"redirectUrl,omitempty"`
	StbSignature  StbSignature `json:"stbSignature,omitempty"` //  response lấy signature STB
}
type StbSignature struct {
	ProfileID           string `json:"ProfileID"`
	AccessKey           string `json:"AccessKey"`
	TransactionID       string `json:"TransactionID"`
	TransactionDateTime string `json:"TransactionDateTime"`
	Language            string `json:"Language"`
	SubscribeOnly       string `json:"SubscribeOnly"`
	IsTokenRequest      string `json:"IsTokenRequest"`
	TotalAmount         int64  `json:"TotalAmount"`
	SSN                 string `json:"SSN"`
	Currency            string `json:"Currency"`
	FirstName           string `json:"FirstName"`
	LastName            string `json:"LastName"`
	Gender              string `json:"Gender"`
	Description         string `json:"Description"`
	Address             string `json:"Address"`
	District            string `json:"District"`
	City                string `json:"City"`
	PostalCode          string `json:"PostalCode"`
	Country             string `json:"Country"`
	Email               string `json:"Email"`
	Mobile              string `json:"Mobile"`
	ReturnUrl           string `json:"ReturnUrl"`
	CancelUrl           string `json:"CancelUrl"`
	Signature           string `json:"Signature"`
	RawText             string `json:"RawText"`
}
