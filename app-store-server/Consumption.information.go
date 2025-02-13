package AppStoreServer

import (
	Apple "github.com/godrealms/go-apple-sdk"
	"github.com/godrealms/go-apple-sdk/types"
)

// ConsumptionRequest The request body containing consumption information.
type ConsumptionRequest struct {
	// (Required) The age of the customer’s account.
	// The age of the customer’s account.
	// 0: Account age is undeclared. Use this value to avoid providing information for this field.
	// 1: Account age is between 0–3 days.
	// 2: Account age is between 3–10 days.
	// 3: Account age is between 10–30 days.
	// 4: Account age is between 30–90 days.
	// 5: Account age is between 90–180 days.
	// 6: Account age is between 180–365 days.
	// 7: Account age is over 365 days.
	AccountTenure types.AccountTenure `json:"accountTenure"`

	// (Required) The UUID of the in-app user account that completed the in-app purchase transaction.
	AppAccountToken types.UUID `json:"appAccountToken"`

	// (Required) A value that indicates the extent to which the customer consumed the in-app purchase.
	// 0: The consumption status is undeclared. Use this value to avoid providing information for this field.
	// 1: The in-app purchase is not consumed.
	// 2: The in-app purchase is partially consumed.
	// 3: The in-app purchase is fully consumed.
	ConsumptionStatus types.ConsumptionStatus `json:"consumptionStatus"`

	// (Required) A Boolean value of true or false that indicates whether the customer consented to provide consumption data.
	//Note: The App Store server rejects requests that have a customerConsented value other than true by returning an HTTP 400 error with an InvalidCustomerConsentError.
	CustomerConsented types.CustomerConsented `json:"customerConsented"`

	// (Required) A value that indicates whether the app successfully delivered an in-app purchase that works properly.
	// 0: The app delivered the consumable in-app purchase and it’s working properly.
	// 1: The app didn’t deliver the consumable in-app purchase due to a quality issue.
	// 2: The app delivered the wrong item.
	// 3: The app didn’t deliver the consumable in-app purchase due to a server outage.
	// 4: The app didn’t deliver the consumable in-app purchase due to an in-game currency change.
	// 5: The app didn’t deliver the consumable in-app purchase for other reasons.
	DeliveryStatus types.DeliveryStatus `json:"deliveryStatus"`

	// (Required) A value that indicates the total amount, in USD, of in-app purchases the customer has made in your app, across all platforms.
	// 0: Lifetime purchase amount is undeclared. Use this value to avoid providing information for this field.
	// 1: Lifetime purchase amount is 0 USD.
	// 2: Lifetime purchase amount is between 0.01–49.99 USD.
	// 3: Lifetime purchase amount is between 50–99.99 USD.
	// 4: Lifetime purchase amount is between 100–499.99 USD.
	// 5: Lifetime purchase amount is between 500–999.99 USD.
	// 6: Lifetime purchase amount is between 1000–1999.99 USD.
	// 7: Lifetime purchase amount is over 2000 USD.
	LifetimeDollarsPurchased types.LifetimeDollarsPurchased `json:"lifetimeDollarsPurchased"`

	// (Required) A value that indicates the total amount, in USD, of refunds the customer has received, in your app,
	// across all platforms.
	// 0: Lifetime refund amount is undeclared. Use this value to avoid providing information for this field.
	// 1: Lifetime refund amount is 0 USD.
	// 2: Lifetime refund amount is between 0.01–49.99 USD.
	// 3: Lifetime refund amount is between 50–99.99 USD.
	// 4: Lifetime refund amount is between 100–499.99 USD.
	// 5: Lifetime refund amount is between 500–999.99 USD.
	// 6: Lifetime refund amount is between 1000–1999.99 USD.
	// 7: Lifetime refund amount is over 2000 USD.
	LifetimeDollarsRefunded types.LifetimeDollarsRefunded `json:"lifetimeDollarsRefunded"`

	// (Required) A value that indicates the platform on which the customer consumed the in-app purchase.
	// 0: The platform is undeclared. Use this value to avoid providing information for this field.
	// 1: An Apple platform.
	// 2: Non-Apple platform.
	Platform types.Platform `json:"platform"`

	// (Required) A value that indicates the amount of time that the customer used the app.
	// 0: The engagement time is undeclared. Use this value to avoid providing information for this field.
	// 1: The engagement time is between 0–5 minutes.
	// 2: The engagement time is between 5–60 minutes.
	// 3: The engagement time is between 1–6 hours.
	// 4: The engagement time is between 6–24 hours.
	// 5: The engagement time is between 1–4 days.
	// 6: The engagement time is between 4–16 days.
	// 7: The engagement time is over 16 days.
	PlayTime types.PlayTime `json:"playTime"`

	// A value that indicates your preference, based on your operational logic, as to whether Apple should grant the refund.
	// 0: The refund preference is undeclared. Use this value to avoid providing information for this field.
	// 1: You prefer that Apple grants the refund.
	// 2: You prefer that Apple declines the refund.
	// 3: You have no preference whether Apple grants or declines the refund.
	RefundPreference types.RefundPreference `json:"refundPreference"`

	// (Required) A Boolean value of true or false that indicates whether you provided, prior to its purchase,
	//a free sample or trial of the content, or information about its functionality.
	SampleContentProvided types.SampleContentProvided `json:"sampleContentProvided"`

	// (Required) The status of the customer’s account.
	// 0: Account status is undeclared. Use this value to avoid providing information for this field.
	// 1: The customer’s account is active.
	// 2: The customer’s account is suspended.
	// 3: The customer’s account is terminated.
	// 4: The customer’s account has limited access.
	UserStatus types.UserStatus `json:"userStatus"`
}

// SendConsumptionInformation
// Send consumption information about a consumable in-app purchase or auto-renewable subscription
// to the App Store after your server receives a consumption request notification.
func SendConsumptionInformation(client *Apple.Client, transactionId string, body *ConsumptionRequest) error {
	var result = ""
	client.SetService(Apple.AppStoreServerClient)
	params := Apple.RequestParams{
		Method: "PUT",
		Path:   "/inApps/v1/transactions/consumption/{transactionId}",
		Result: &result,
		Headers: map[string]string{
			"Accept": "application/json",
		},
		PathParams: map[string]string{
			"transactionId": transactionId,
		},
		Body: body,
	}

	if err := client.Request(params); err != nil {
		return err
	}
	return nil
}
