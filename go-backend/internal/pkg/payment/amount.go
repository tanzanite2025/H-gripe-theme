package payment

import domainmoney "commerce-platform/internal/domain/money"

func paymentMoneyFromMajor(amount float64, code string) (domainmoney.Money, error) {
	return domainmoney.FromMajorFloat(amount, code)
}

func paymentMoneyFromMinor(amount int64, code string) (domainmoney.Money, error) {
	return domainmoney.New(amount, code)
}

func paymentMajorFloatFromMinor(amount int64, code string) (float64, error) {
	money, err := paymentMoneyFromMinor(amount, code)
	if err != nil {
		return 0, err
	}
	return money.MajorFloat()
}

func paymentMajorString(amount float64, code string) (string, error) {
	money, err := paymentMoneyFromMajor(amount, code)
	if err != nil {
		return "", err
	}
	return money.FormatMajor()
}
