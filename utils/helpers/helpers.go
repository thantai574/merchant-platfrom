package helpers

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/jakehl/goid"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/metadata"
	"io/ioutil"
	"net/http"
	"orders-system/domain/constants"
	"orders-system/utils/logger"
	"sort"
	"time"
)

func CreateHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

func CreateMd5(key string) []byte {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hasher.Sum(nil)
}

func Ksort(c map[string]interface{}) (map[string]interface{}, []string) {
	to := make(map[string]interface{})
	var keys []string
	for s := range c {
		if s != "hash" {
			keys = append(keys, s)
		}

	}
	sort.Strings(keys)

	for _, v := range keys {
		str := fmt.Sprint(c[v])
		to[v] = str
	}

	return to, keys
}

func GetUUId() string {
	v4UUID := goid.NewV4UUID()
	return fmt.Sprint(v4UUID.String())
}

func LocationVietNam() *time.Location {
	location, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		fmt.Println(err)
	}
	return location
}

func GetCurrentTime() time.Time {
	location, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		fmt.Println(err)
	}

	timeNow := time.Now()

	return timeNow.In(location)
}

func CurrentHour() time.Time {
	timezone, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	year, week := GetCurrentTime().ISOWeek()

	date := time.Date(year, time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, timezone)

	isoYear, isoWeek := date.ISOWeek()

	for date.Day() < time.Now().Day() { // iterate back to Monday
		date = date.AddDate(0, 0, 1)
		isoYear, isoWeek = date.ISOWeek()
	}
	for isoYear < year { // iterate forward to the first day of the first week
		date = date.AddDate(0, 0, 1)
		isoYear, isoWeek = date.ISOWeek()
	}
	for isoWeek < week { // iterate forward to the first day of the given week
		date = date.AddDate(0, 0, 1)
		isoYear, isoWeek = date.ISOWeek()
	}

	for date.Hour() != GetCurrentTime().Hour() { // iterate back to Monday
		date = date.Add(time.Hour)
		isoYear, isoWeek = date.ISOWeek()
	}
	return date
}

func PrevHour() time.Time {
	timezone, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	year, week := GetCurrentTime().ISOWeek()

	date := time.Date(year, time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, timezone)
	isoYear, isoWeek := date.ISOWeek()

	for isoYear < year { // iterate forward to the first day of the first week
		date = date.AddDate(0, 0, 7)
		isoYear, isoWeek = date.ISOWeek()
	}
	for isoWeek < week { // iterate forward to the first day of the given week
		date = date.AddDate(0, 0, 7)
		isoYear, isoWeek = date.ISOWeek()
	}

	for date.Hour() != GetCurrentTime().Hour()-1 { // iterate back to Monday
		date = date.Add(time.Hour)
		isoYear, isoWeek = date.ISOWeek()
	}
	return date
}

func NextHour() time.Time {
	timezone, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	year, week := GetCurrentTime().ISOWeek()

	date := time.Date(year, time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, timezone)
	isoYear, isoWeek := date.ISOWeek()

	for isoYear < year { // iterate forward to the first day of the first week
		date = date.AddDate(0, 0, 7)
		isoYear, isoWeek = date.ISOWeek()
	}
	for isoWeek < week { // iterate forward to the first day of the given week
		date = date.AddDate(0, 0, 7)
		isoYear, isoWeek = date.ISOWeek()
	}

	next_hour := GetCurrentTime().Hour() + 1
	if GetCurrentTime().Hour() == 23 {
		next_hour = 0
	}
	for date.Hour() != next_hour { // iterate back to Monday
		date = date.Add(time.Hour)
		fmt.Println(2)
		isoYear, isoWeek = date.ISOWeek()
	}
	return date
}

func FirstDayOfISOWeek() time.Time {
	timezone, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	year, week := GetCurrentTime().ISOWeek()

	date := time.Date(year, time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, timezone)
	isoYear, isoWeek := date.ISOWeek()

	for date.Weekday() != time.Monday { // iterate back to Monday
		date = date.AddDate(0, 0, -1)
		isoYear, isoWeek = date.ISOWeek()
	}
	for isoYear < year { // iterate forward to the first day of the first week
		date = date.AddDate(0, 0, 7)
		isoYear, isoWeek = date.ISOWeek()
	}
	for isoWeek < week { // iterate forward to the first day of the given week
		date = date.AddDate(0, 0, 7)
		isoYear, isoWeek = date.ISOWeek()
	}
	return date
}

func EndDayOfISOWeek() time.Time {
	timezone, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	year, week := GetCurrentTime().ISOWeek()

	date := time.Date(year, time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, timezone)
	isoYear, isoWeek := date.ISOWeek()

	for date.Weekday() != time.Sunday { // iterate back to Monday
		date = date.AddDate(0, 0, 1)
		isoYear, isoWeek = date.ISOWeek()
	}
	for isoYear < year { // iterate forward to the first day of the first week
		date = date.AddDate(0, 0, 7)
		isoYear, isoWeek = date.ISOWeek()
	}
	for isoWeek < week { // iterate forward to the first day of the given week
		date = date.AddDate(0, 0, 7)
		isoYear, isoWeek = date.ISOWeek()
	}
	return date
}

func ContextMetada(kv []string, parent context.Context) (ctx context.Context) {
	md := metadata.Pairs(kv...)

	ctx = metadata.NewOutgoingContext(parent, md)
	return
}

func IsStringSliceContains(stringSlice []string, searchString string) bool {
	for _, value := range stringSlice {
		if value == searchString {
			return true
		}
	}
	return false
}

func ContextWithTimeOut() (context_data context.Context) {
	context_data, _ = context.WithTimeout(context.Background(), time.Minute*10)
	return context_data
}

func HttpRequest(request struct {
	Uri      string
	Path     string
	Method   string
	Headers  map[string]string
	Body     interface{}
	Response interface{}
}) (err error) {
	log, err := logger.NewLogger("production")

	client := new(http.Client)

	client.Timeout = time.Minute * 1

	jsonrequest, err := json.Marshal(request.Body)

	req, err := http.NewRequest(request.Method, fmt.Sprintf("%v%v", request.Uri, request.Path), bytes.NewReader(jsonrequest))

	if err != nil {
		return err
	}

	if request.Method == http.MethodGet {
		for key, value := range request.Headers {
			req.Header.Add(key, value)
		}
	}

	req.Header.Add("Content-Type", `application/json`)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	responseByte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.With(zapcore.Field{
		Key:       "uri",
		Type:      zapcore.StringType,
		String:    fmt.Sprintf("%v%v", request.Uri, request.Path),
		Interface: nil,
	}).With(
		zapcore.Field{
			Key:       "data",
			Type:      zapcore.StringType,
			String:    string(responseByte),
			Interface: nil,
		}).Info("http_request_data")

	err = json.Unmarshal(responseByte, request.Response)
	if err != nil {
		return err
	}
	//Close Request
	defer func() {
		err = resp.Body.Close()
	}()

	return err
}

func MapETransTypeToOldTrans(ctx context.Context, eTransType, eSubTransType string) (old_transaction_type string, err error) {

	// ---------------------- MAP BY SUB_TRANSTYPE -----------------------

	// Hoá đơn tiền điện
	if eSubTransType == constants.SUB_TRANSTYPE_WALLET_PAY_BILL_ELECTRIC {
		return constants.TransactionTypePaidBillElectric, nil
	}

	// Hoá đơn tiền nước
	if eSubTransType == constants.SUB_TRANSTYPE_WALLET_PAY_BILL_WATTER {
		return constants.TransactionTypePaidBillWater, nil
	}

	// Hoá đơn vay tiêu dùng
	if eSubTransType == constants.SUB_TRANSTYPE_WALLET_PAY_BILL_LOAN {
		return constants.TransactionTypePaidBillCredit, nil
	}

	// Nạp tiền điện thoại
	if eSubTransType == constants.SUB_TRANSTYPE_WALLET_TOPUP_CARD {
		return constants.TransactionTypeTopup, nil
	}

	// Mua thẻ cào  constants.TransactionTypeBuyCard
	if eSubTransType == constants.SUB_TRANSTYPE_WALLET_BUY_CARD {
		return constants.TransactionTypeBuyCard, nil
	}

	// MERCHANT BILL
	if eSubTransType == constants.TransactionTypePaidBillMerchantWallet {
		return constants.TRANSTYPE_WALLET_PAY, nil
	}

	// Mua thẻ data
	if eSubTransType == constants.SUB_TRANSTYPE_WALLET_BUY_DATA {
		return constants.TransactionTypeBuyData, nil
	}

	// Chuyển tiền, Nhận tiền TransactionTypeTransferGpoint
	if eTransType == constants.TRANSTYPE_WALLET_TRANSFER {
		return constants.TransactionTypeTransferGpoint, nil
	}

	// ---------------------- MAP BY TRANSTYPE -----------------------
	// Rút tiền
	if eTransType == constants.TRANSTYPE_WALLET_CASH_OUT {
		return constants.TransactionTypeBalanceChangeWithdraw, nil
	}

	// Nạp tiền constants.TransactionTypeBalanceChangeDeposit
	if eTransType == constants.TRANSTYPE_WALLET_CASH_IN {
		return constants.TransactionTypeBalanceChangeDeposit, nil
	}

	// cashback
	if eTransType == constants.TRANSTYPE_WALLET_CASHBACK {
		return constants.TransactionTypeCashback, nil
	}

	//Rút tiền khỏi ví IBFT
	if eTransType == constants.TRANSTYPE_WALLET_TRANS2BANK {
		return constants.TRANSTYPE_WALLET_TRANS2BANK, nil
	}

	// thu hộ chị hộ merchant -> user qua ví gpay
	if eTransType == constants.TRANSTYPE_BEHALF_PAY_WALLET_TO_WALLET {
		return constants.TRANSTYPE_BEHALF_PAY_WALLET_TO_WALLET, nil
	}

	if eTransType == constants.TRANSTYPE_WALLET_LIXI {
		return constants.TRANSTYPE_WALLET_LIXI, nil
	}

	return "", nil
}
