package config

import "os"

func GetDataSourceURL() string {
	return os.Getenv("DATA_SOURCE_URL")
}

func GetApplicationPort() string {
	return os.Getenv("APPLICATION_PORT")
}

func GetPaymentServiceUrl() string {
	return os.Getenv("PAYMENT_SERVICE_URL")
}
