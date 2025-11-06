package constants

const (
	TopicCreateOrder  = "create-order"
	TopicFailedOrder  = "failed-order"
	TopicSuccessOrder = "success-order"

	TopicMQTTUpdateExpiredOrder = "update-expired-order"
	TopicMQTTUpdateBalanceVA    = "update-balance-va"
	TopicUpdateCancelOrder      = "update-cancel-order"
	TopicMQTTCloseVA            = "close-va"

	TopicUpdateVABalance   = "update-va-balance"
	TopicMQTTFailFraudMPGS = "mpgs-fail-fraud"
	TopicUpdateBankStatus  = "update-bank-status"
)

const TRANSFER_POINT_SUCCESS = "transfer-point-success"
const UPDATE_USER_INFO = "update-user-info"
const DEPOSIT_SUCCESS = "deposit-success"
const WITHDRAW_SUCCESS = "withdraw-success"
const WITHDRAW_CANCEL = "withdraw-cancel"
const DEPOSIT_CANCEL = "deposit-cancel"
const UPDATE_TRANSACTION = "update-transaction"
const REDIRECT_URL = "redirect-url"

type Message struct {
	Event       string      `json:"t"`
	Key         string      `json:"k"`
	MessageData interface{} `json:"d"`
}

const MQTTTransferPointSuccess = "transfer-point-success"
const MQTTUpdateUserInfo = "update-user-info"
const MQTTDepositSuccess = "deposit-success"
const MQTTWithdrawSuccess = "withdraw-success"
const MQTTWithdrawCancel = "withdraw-cancel"
const MQTTDepositCancel = "deposit-cancel"
const MQTTUpdateTransaction = "update-transaction"
const MQTTTransactionSuccess = "success-transaction"
const MQTTLIXISuccess = "success-lixi"
const MQTTTransactionFailed = "failed-transaction"
const MQTTKycVerified = "kyc-verified"

const (
	MQTTEventNotification = "notification"
	MQTTEventBackground   = "background"
)

const STATUS_SUCCESS = "SUCCESS"
const STATUS_VERIFYING = "VERIFYING"
const STATUS_FAILED = "FAILED"
const STATUS_PROCESSING = "PROCESSING"
