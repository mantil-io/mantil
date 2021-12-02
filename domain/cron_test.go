package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateAWSCron(t *testing.T) {
	cases := []struct {
		cron  string
		valid bool
	}{
		{"* * * * ? *", true},
		{"*/1 * * * ? *", true},
		{"* */1 * * ? *", true},
		{"* * */1 * ? *", true},
		{"* * * */1 ? *", true},
		{"0 18 ? * MON-FRI *", true},
		{"0 10 * * ? *", true},
		{"15 12 * * ? *", true},
		{"0 8 1 * ? *", true},
		{"0/5 8-17 ? * MON-FRI *", true},
		{"0 9 ? * 2#1 *", true},
		{"0 07/12 ? * * *", true},
		{"10,20,30,40 07/12 ? * * *", true},
		{"10 10,15,20,23 ? * * *", true},
		{"10 10 15,30,31 * ? *", true},
		{"10 10 15 JAN,JUL,DEC ? *", true},
		{"10 10 31 04,09,12 ? *", true},
		{"0,5 07/12 ? * 01,05,7 *", true},
		{"0,5 07/12 ? * 01,05,7 2020,2021,2028,2199", true},
		{"0,5 07/12 ? * 01,05,7 2020,2021,2028,2199", true},
		{"0,5 07/12 ? * 01,05,7 2020-2199", true},

		{"* * * * * *", false},
		{"0 18 ? * MON-FRI", false},
		{"0 18 * * * *", false},
		{"0 65 * * ? *", false},
		{"89 10 * * ? *", false},
		{"15/65 10 * * ? *", false},
		{"15/30 10 * * ? 2400", false},
		{"0 9 ? * 2#6 *", false},
		{"0 9 ? * ? *", false},
		{"10 10 31 04,09,13 ? *", false},
		{"0,5 07/12 ? * 01,05,8 *", false},
		{"0,5 07/12 ? * 01,05,7 2020,2021,2028,1111", false},
		{"0,5 07/12 ? * 01,05,7 2020,2021,2028,1969", false},
		{"0,5 07/12 ? * 01,05,7 1969-2100", false},
	}
	for _, c := range cases {
		require.Equal(t, c.valid, ValidateAWSCron(c.cron), c.cron)
	}
}
