package validator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// futuresTickerRegex matches 1-5 uppercase product letters one month code letter
// and exactly 2 digit year suffix for example: CLJ26, ZCJ26, MCLM26
var futuresTickerRegex = regexp.MustCompile(`^[A-Z]{1,5}[FGHJKMNQUVXZ]\d{2}$`)

func validateFuturesTicker(fl validator.FieldLevel) bool {
	return futuresTickerRegex.MatchString(fl.Field().String())
}
