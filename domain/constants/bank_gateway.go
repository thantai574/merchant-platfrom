package constants

const (
	GPAY_VCCB             = "GPVCCB"        // ngân hàng Bản Việt ibft Gpay
	GPNAPASGPBANK         = "GPNAPASGPBANK" // NAPAS BankCode
	VCCB                  = "VCCB"
	MSB                   = "MSB"
	VPB                   = "VPB" // ngân hàng VPBank Credit Payment
	SUCCESS_ERR_CODE      = "000"
	WRONG_OTP_ERR_CODE    = "602"
	FAIL_MAX_OTP_ERR_CODE = "603"
)

var TimeoutErrCode = []string{"306", "502"}

// --------------------- Update 02/02/2021 Bank Gateway ---------------------
type BankGwStatus string

const (
	BankStatusSuccess BankGwStatus = "Success"
	BankStatusPending BankGwStatus = "Pending"
	BankStatusFail    BankGwStatus = "Fail"
	BankOtherStatus   BankGwStatus = "Other"
)

func (status BankGwStatus) IsSuccess() bool {
	if len(status) > 2 && status[0:1] == "2" {
		return true
	}
	return false
}

func (status BankGwStatus) IsVerifying() bool {
	if status == "102" {
		return true
	}
	return false
}

func (status BankGwStatus) IsFail() bool {
	if len(status) > 2 && status[0:1] == "4" {
		return true
	}
	return false
}

func (status BankGwStatus) IsNeedToEnterOTP() bool {
	if status == "453" {
		return true
	}
	return false
}

func (status BankGwStatus) IsNotByPassOTP() bool {
	if status == "452" {
		return true
	}
	return false
}

func (status BankGwStatus) IsMaxAttemptFailOTP() bool {
	if status == "461" {
		return true
	}
	return false
}

func (status BankGwStatus) IsWrongOTP() bool {
	if status == "449" {
		return true
	}
	return false
}
