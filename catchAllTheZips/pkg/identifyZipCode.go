package pkg

import "regexp"

func IdentifyZipCode(zipCode string) bool {
	cepRegex := regexp.MustCompile(`^\d{8}$`)
	return cepRegex.MatchString(zipCode)
}
