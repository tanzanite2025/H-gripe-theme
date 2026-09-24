package payment

import domainmoney "commerce-platform/internal/domain/money"

func paymentMoneyFromMinor(amount int64, code string) (domainmoney.Money, error) {
	return domainmoney.New(amount, code)
}

func paymentMajorStringFromMinor(amount int64, code string) (string, error) {
	money, err := paymentMoneyFromMinor(amount, code)
	if err != nil {
		return "", err
	}
	return money.FormatMajor()
}
