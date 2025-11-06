package bank_service

import (
	"fmt"
	"math/rand"
	"orders-system/domain/constants"
	entities "orders-system/domain/entities/bank_gateway"
	"orders-system/utils/helpers"
	"orders-system/utils/logger"
	"testing"
)

//@todo xác thực OTP nạp tiền vào ví
func Test_repoImpl_VerifyOTP(t *testing.T) {
	log, _ := logger.NewLogger("DEV")
	impleBank := NewRepoImpl("http://34.126.107.53/", log)

	type args struct {
		linkId         int64
		otp            string
		gpayBankCode   string
		refBankTraceID string
		orderId        string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		orderId string
	}{
		{
			name: "case VIETIN BANK",
			args: args{
				linkId:       1705,
				otp:          "860486",
				gpayBankCode: "GPCTG",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := impleBank
			if response, err := r.VerifyOTP(tt.args.gpayBankCode, tt.args.refBankTraceID, tt.args.orderId, tt.args.linkId, tt.args.otp); (err != nil) != tt.wantErr {
				t.Errorf("VerifyOTP() error = %v, wantErr %v", err, tt.wantErr)
				fmt.Println("response", response)
			}
		})
	}
}

func Test_repoImpl_UnLink(t *testing.T) {
	logger, _ := logger.NewLogger("DEV")

	impl_Bank := NewRepoImpl("DEV", logger)

	type fields struct {
		Env string
	}
	type args struct {
		linkId   int64
		bankCode string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "un link",
			fields: fields{},
			args: args{
				linkId:   1,
				bankCode: "GPSTB",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := impl_Bank
			if _, err := r.UnLink(tt.args.bankCode, tt.args.linkId); (err != nil) != tt.wantErr {
				t.Errorf("UnLink() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

//@todo nạp tiền về ví (PASS)
func Test_repoImpl_CashIn(t *testing.T) {
	log, _ := logger.NewLogger("DEV")
	impleBank := NewRepoImpl("http://34.126.107.53/", log)

	type args struct {
		gpayBankCode string
		linkId       int64
		amount       int64
		gpayOrderId  string
		userId       string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		//{
		//	name: "nap tien vao vi gpay BIDV",
		//	args: args{
		//		linkId:       497,
		//		amount:       88888,
		//		gpayOrderId:  "BIDV_GPAYe_wq1fda",
		//		userId:       "kma12",
		//		gpayBankCode: "GPBIDV",
		//	},
		//	wantErr: false,
		//},
		{
			name: "nap tien vao vi gpay SACOMBANK",
			args: args{
				gpayBankCode: "GPNAPASGPBANK",
				linkId:       1423,
				amount:       50000,
				gpayOrderId:  "CASHIN11",
				userId:       "20201202-US5FC70B126D017888332",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := impleBank
			res, err := r.CashIn(tt.args.gpayBankCode, tt.args.linkId, tt.args.amount, tt.args.gpayOrderId, tt.args.userId)
			if err != nil {
				t.Error(err)
			}
			t.Log("ressss", res)
		})
	}
}

//@todo rút tiền về tài khoản ngân hàng (PASS)
func Test_repoImpl_CashOut(t *testing.T) {
	logger, _ := logger.NewLogger("DEV")
	r := NewRepoImpl("DEV", logger)

	type args struct {
		amount   int64
		bankCode string
		orderId  string
		linkId   int64
		userId   string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "rut tien tu vi vao tai khoan ngan hang",
			args: args{
				amount:   99999,
				bankCode: "GPBIDV",
				orderId:  "orderasxi1ưe23",
				linkId:   215,
				userId:   "rut tien tdu vi vao tai khoan ngan hang",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := r.CashOut(tt.args.amount, tt.args.bankCode, tt.args.orderId, tt.args.linkId, "cashOut", tt.args.userId); (err != nil) != tt.wantErr {
				t.Errorf("CashOut() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

//
////@todo check balance Gpay (FAIL)
func Test_repoImpl_CheckGpayBankBalance(t *testing.T) {
	log, _ := logger.NewLogger("DEV")

	type args struct {
		gpayBankCode  string
		accountNumber string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "check bank balance of gpay account",
			args: args{
				gpayBankCode:  "GPBIDV",
				accountNumber: "9704180023232312",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRepoImpl("http://34.126.107.53/", log)
			if _, err := r.CheckGpayBankBalance(tt.args.gpayBankCode, tt.args.accountNumber); (err != nil) != tt.wantErr {
				t.Errorf("CheckGpayBankBalance() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

////@todo check balance ibft gpay (FAIL)
func Test_repoImpl_IBFTCheckBalance(t *testing.T) {
	log, _ := logger.NewLogger("DEV")

	type args struct {
		accountNumber string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "check current balance of GPAY IBFT",
			args: args{
				accountNumber: "1231231243123123",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRepoImpl("http://34.126.107.53/", log)
			if _, err := r.IBFTCheckBalance(tt.args.accountNumber); (err != nil) != tt.wantErr {
				t.Errorf("IBFTCheckBalance() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

//@todo check thong tin nguoi nhận IBFT
func Test_repoImpl_IBFTInquiry(t *testing.T) {
	log, _ := logger.NewLogger("DEV")
	type args struct {
		accountNumber string
		cardNumber    string
		ibftCode      string
		description   string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "check ibft username",
			args: args{
				accountNumber: "",
				cardNumber:    "9704542000234196",
				ibftCode:      "970454",
				description:   "lay thong tin ibft",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRepoImpl("http://34.126.107.53/", log)
			if _, err := r.IBFTInquiry(tt.args.accountNumber, tt.args.cardNumber, tt.args.ibftCode); (err != nil) != tt.wantErr {
				t.Errorf("IBFTInquiry() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

//@todo chuyển tiền ibft
func Test_repoImpl_IBFTTransfer(t *testing.T) {
	log, _ := logger.NewLogger("DEV")
	type args struct {
		accountNumber string
		cardNumber    string
		gpayUserId    string
		orderId       string
		ibftCode      string
		description   string
		amount        int64
		bankName      string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ibft transfer",
			args: args{
				accountNumber: "",
				cardNumber:    "9704542000234196",
				gpayUserId:    "namle123",
				orderId:       "order1+asdas",
				ibftCode:      "970454",
				description:   "chuyen tiefn ibft",
				amount:        50000,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRepoImpl("DEV", log)
			response, err := r.IBFTTransfer(tt.args.accountNumber, tt.args.cardNumber,
				tt.args.gpayUserId, tt.args.orderId, tt.args.ibftCode, tt.args.description, tt.args.amount, "")
			if err != nil {
				t.Log(err)
			}
			t.Log("response", response)
		})
	}
}

func Test_repoImpl_LinkList(t *testing.T) {
	log, _ := logger.NewLogger("DEV")

	type args struct {
		gpayUserId string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "get linked list bannk",
			args: args{
				gpayUserId: "20201229-US5FEAFC9F25562924793",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRepoImpl("http://34.126.107.53/", log)
			gotResponse, err := r.LinkList(tt.args.gpayUserId)
			if err == nil {
				t.Log(gotResponse.Data[0])
			}
		})
	}
}

func Test_repoImpl_CreditPayment(t *testing.T) {
	log, _ := logger.NewLogger("DEV")

	type args struct {
		dataReq entities.CreditPaymentRequestData
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "credit payment",
			//args: args{
			//	cardNumber:   "4508750015741019",
			//	expiryMonth:  "05",
			//	expiryYear:   "21",
			//	securityCode: "111",
			//	orderId:      "order123456",
			//	amount:       "100000",
			//},
			args: args{dataReq: entities.CreditPaymentRequestData{
				CardNumber:        "4508750015741019",
				GpayTransactionId: "order12345678",
				Amount:            "100000",
				ExpiryYear:        "21",
				ExpiryMonth:       "05",
				SecurityCode:      "111",
				CardHolderName:    "",
				RedirectUrl:       "",
				GpayUserId:        "",
				Token:             "",
				Currency:          "",
				MCC:               "",
				MccType:           "",
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRepoImpl("http://34.126.107.53/", log)
			gotResponse, err := r.CreditPayment(tt.args.dataReq)
			if err == nil {
				t.Log("gotResponse", gotResponse)
			}
		})
	}
}

func Test_repoImpl_CreateVA(t *testing.T) {
	log, _ := logger.NewLogger("DEV")

	type args struct {
		entities.CreateVARequestData
		provider string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "create VA",
			args: args{
				CreateVARequestData: entities.CreateVARequestData{
					AccountName:     "namle",
					AccountType:     "MANYTIME",
					AccountNumber:   "M010200000000010",
					ReferenceNumber: "163377835",
				},
				provider: "GPMSB",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRepoImpl("http://34.126.107.53/", log)
			gotResponse, err := r.CreateVA(tt.args.CreateVARequestData, tt.args.provider)
			if err == nil {
				t.Log("gotResponse", gotResponse)
			}

		})
	}
}

func Test_repoImpl_UpdateVA(t *testing.T) {
	log, _ := logger.NewLogger("DEV")

	type args struct {
		entities.CreateVARequestData
		provider string
	}
	tests := []struct {
		args args
		name string
	}{
		{
			name: "update VA",
			args: args{
				CreateVARequestData: entities.CreateVARequestData{
					AccountName:     "namle",
					AccountType:     "",
					AccountNumber:   "M010200000000119",
					ReferenceNumber: "",
				},
				provider: "GPMSB",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRepoImpl("http://34.126.107.53/", log)
			gotResponse, err := r.UpdateVA(tt.args.CreateVARequestData, tt.args.provider)
			if err == nil {
				t.Log("gotResponse", gotResponse)
			}
		})
	}
}

func Test_repoImpl_DetailVA(t *testing.T) {
	log, _ := logger.NewLogger("DEV")

	type args struct {
		accNumber string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "create VA",
			args: args{
				accNumber: "sd",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRepoImpl("http://34.126.107.53/", log)
			gotResponse, err := r.DetailVA(tt.args.accNumber, "GPMSB")
			if err == nil {
				t.Log("gotResponse", gotResponse.Data)
			}
		})
	}
}

func Test_repoImpl_CloseVA(t *testing.T) {
	// log, _ := logger.NewLogger("DEV")

	// type args struct {
	// 	accountNumber string
	// 	provider      string
	// }
	// tests := []struct {
	// 	name string
	// 	args args
	// }{
	// 	{
	// 		name: "create VA",
	// 		args: args{
	// 			accountNumber: "99800020000000000038",
	// 		},
	// 	},
	// }
	// for _, tt := range tests {
	// 	t.Run(tt.name, func(t *testing.T) {
	// 		r := NewRepoImpl("http://34.126.107.53/", log)
	// 		gotResponse, err := r.CloseVA(tt.args.accountNumber, tt.args.provider)
	// 		if err == nil {
	// 			t.Log("gotResponse", gotResponse.Data)
	// 		}
	// 	})
	// }
}

func Test_repoImpl_RetrieveCreditOrderStatus(t *testing.T) {
	log, _ := logger.NewLogger("DEV")

	type args struct {
		orderId  string
		bankCode string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: " ",
			args: args{
				orderId:  "GPOS202104140000000054",
				bankCode: constants.GPNAPASGPBANK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRepoImpl("http://34.126.107.53/", log)
			gotResponse, err := r.RetrieveOrderStatus(tt.args.orderId, tt.args.bankCode)
			if err == nil {
				t.Log("gotResponse", gotResponse.Data.Amount)
				t.Log("gotResponse", gotResponse.Data.GpayUserId)
			}
		})
	}
}

func Test_repoImpl_CheckBankBin(t *testing.T) {
	log, _ := logger.NewLogger("DEV")

	type args struct {
		bankbin string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "create VA",
			args: args{
				bankbin: "401200",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRepoImpl("http://34.126.107.53/", log)
			gotResponse, err := r.CheckInternationalBankBin(tt.args.bankbin)
			if err == nil {
				t.Log("gotResponse", gotResponse)
			} else {
				t.Log("err", err)
			}
		})
	}
}

func Test_repoImpl_PGInitTrans(t *testing.T) {
	log, _ := logger.NewLogger("DEV")

	r := NewRepoImpl("http://34.126.107.53/", log)
	gotResponse, err := r.PGInitTrans("GPBIDV", entities.PGInitTransReqData{
		GPayTransactionID: helpers.GetUUId(),
		Description:       "hello world",
		Amount:            "10000",
		RedirectURL:       "https://example.com",
	})
	if err == nil {
		t.Log("gotResponse", gotResponse)
	}
}

func Test_repoImpl_Link(t *testing.T) {
	log, _ := logger.NewLogger("DEV")
	r := NewRepoImpl("http://34.126.107.53/", log)

	type args struct {
		data         entities.LinkRequestData
		gpayBankCOde string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "case 1",
			args: args{
				data: entities.LinkRequestData{

					GpayUserId:  "namle123",
					Channel:     "WEB",
					ReturnUrl:   "",
					CancelUrl:   "",
					Description: "",
				},
				gpayBankCOde: "GPNAPASPGBANK",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResponse, err := r.Link(tt.args.data, tt.args.gpayBankCOde, "tét")
			if (err != nil) != tt.wantErr {
				t.Errorf("Link() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Log("gotResponse", gotResponse)
		})
	}
}

func Test_repoImpl_CashInNapas(t *testing.T) {
	log, _ := logger.NewLogger("DEV")
	r := NewRepoImpl("http://34.126.107.53/", log)

	type args struct {
		data           entities.NapasCashInDataRequest
		gpay_bank_code string
	}
	gpayCashInOrder := "ORDER_CASHIN_" + helpers.GetCurrentTime().Format("20060102150405") + "_" + fmt.Sprint(rand.Int63n(9999-1000)+1000)

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{

			name: "",
			args: args{
				data: entities.NapasCashInDataRequest{
					Amount:            100000,
					GpayTransactionID: gpayCashInOrder,
					LinkID:            1423,
					GPayUserID:        "20201202-US5FC70B126D017888332",
					Channel:           "WEB",
				},
				gpay_bank_code: "GPNAPASBIDV",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResponse, err := r.CashInNapas(tt.args.data, tt.args.gpay_bank_code, "tét")
			if (err != nil) != tt.wantErr {
				t.Errorf("CashInNapas() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Log(gotResponse)
		})
	}
}

func Test_repoImpl_ReFund(t *testing.T) {
	log, _ := logger.NewLogger("DEV")
	r := NewRepoImpl("http://34.126.107.53/", log)

	type args struct {
		gpayTrans    string
		amount       int64
		gpayBankCode string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		//{
		//
		//	name: "",
		//	args: args{
		//		gpayTrans: "GPOS202104090000000123",
		//		amount:    999,
		//		gpayBankCode: entities.GpayBankCodeCreditPayment ,
		//	},
		//	wantErr: false,
		//},
		{

			name: "",
			args: args{
				gpayTrans:    "GPOS202104150000000008",
				amount:       999,
				gpayBankCode: constants.GPNAPASGPBANK,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResponse, err := r.ReFund(tt.args.gpayTrans, tt.args.amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReFund() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Log(gotResponse)
		})
	}
}

func Test_repoImpl_CheckByPassOTP(t *testing.T) {
	log, _ := logger.NewLogger("DEV")
	r := NewRepoImpl("http://34.126.107.53/", log)

	type args struct {
		userId   string
		amount   string
		apiName  string
		bankCOde string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{

			name: "",
			args: args{
				userId:   "US539661US5F15117FAC252397921",
				amount:   "1111111",
				apiName:  "CASHIN",
				bankCOde: "GPBIDV",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResponse, err := r.CheckByPassOTP(entities.CheckByPassOTPDataReq{
				Amount:       tt.args.amount,
				ApiName:      tt.args.apiName,
				GpayBankCode: tt.args.bankCOde,
				GpayUserId:   tt.args.userId,
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("ReFund() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Log("gotResponse", gotResponse)
		})
	}
}

func Test_repoImpl_GetBanks(t *testing.T) {
	log, _ := logger.NewLogger("DEV")

	r := NewRepoImpl("http://34.126.107.53/", log)
	gotResponse, err := r.GetBanks()
	if err == nil {
		t.Log("gotResponse", gotResponse)
	}
}
