package signup

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

// https://developer.typeform.com/webhooks/example-payload/
// https://mholt.github.io/json-to-go/
func TestTypeformWithExamplePayload(t *testing.T) {
	buf, err := ioutil.ReadFile("testdata/typeform.json")
	require.NoError(t, err)
	var tf TypeformWebhook
	err = json.Unmarshal(buf, &tf)
	require.NoError(t, err)

	require.Equal(t, "ianic@mantil.com", tf.Email())
}

func TestTypeformWithOurSignupFormPayload(t *testing.T) {
	buf, err := ioutil.ReadFile("testdata/typeform_our_form.json")
	require.NoError(t, err)
	var tf TypeformWebhook
	err = json.Unmarshal(buf, &tf)
	require.NoError(t, err)

	require.Equal(t, "an_account@example.com", tf.Email())

	rec := tf.AsRecord("", nil)
	require.Equal(t, "ianic", rec.Name)
	require.Equal(t, "an_account@example.com", rec.Email)
	require.Equal(t, "PM", rec.Position)
	require.Equal(t, "71+", rec.OrganizationSize)
	require.Equal(t, false, rec.Developer)
	require.True(t, rec.CreatedAt > 0)
	require.Len(t, rec.ActivationCode, 22)
}

func TestTypeformOur3Responses(t *testing.T) {
	buf, err := ioutil.ReadFile("testdata/typeform_our_3_responses.json")
	require.NoError(t, err)
	var tf TypeformWebhook
	err = json.Unmarshal(buf, &tf)
	require.NoError(t, err)

	//fmt.Printf("answers map: %+v\n", tf.AnswersMap())

	m := tf.Survey()
	require.Len(t, m, 3)

	require.Equal(t, "ana@mantil.com", tf.Email())
	rec := tf.AsRecord("", nil)
	require.Equal(t, "Ana", rec.Name)
	require.Equal(t, "ana@mantil.com", rec.Email)
	require.Equal(t, "", rec.Position)
	require.Equal(t, "2-10", rec.OrganizationSize)
}

// func TestBugReport(t *testing.T) {
// 	var suspiciousRequest = `eyJ2ZXJzaW9uIjoiMS4wIiwicmVzb3VyY2UiOiIvc2lnbnVwL3twcm94eSt9IiwicGF0aCI6InR5cGVmb3JtIiwiaHR0cE1ldGhvZCI6IlBPU1QiLCJoZWFkZXJzIjp7IkNvbnRlbnQtTGVuZ3RoIjoiMTMwNiIsIkNvbnRlbnQtVHlwZSI6ImFwcGxpY2F0aW9uL2pzb24iLCJIb3N0IjoieXRnNWdma2c1ay5leGVjdXRlLWFwaS5ldS1jZW50cmFsLTEuYW1hem9uYXdzLmNvbSIsIlVzZXItQWdlbnQiOiJUeXBlZm9ybSBXZWJob29rcyIsIlgtQW16bi1UcmFjZS1JZCI6IlJvb3Q9MS02MTllNTJjMy0zZjEyMjQ2MTdiYmFjNmZiMDlhMjA3YWEiLCJYLUZvcndhcmRlZC1Gb3IiOiIzNS4xNjkuMTUxLjgxIiwiWC1Gb3J3YXJkZWQtUG9ydCI6IjQ0MyIsIlgtRm9yd2FyZGVkLVByb3RvIjoiaHR0cHMiLCJhY2NlcHQtZW5jb2RpbmciOiJnemlwIn0sIm11bHRpVmFsdWVIZWFkZXJzIjp7IkNvbnRlbnQtTGVuZ3RoIjpbIjEzMDYiXSwiQ29udGVudC1UeXBlIjpbImFwcGxpY2F0aW9uL2pzb24iXSwiSG9zdCI6WyJ5dGc1Z2ZrZzVrLmV4ZWN1dGUtYXBpLmV1LWNlbnRyYWwtMS5hbWF6b25hd3MuY29tIl0sIlVzZXItQWdlbnQiOlsiVHlwZWZvcm0gV2ViaG9va3MiXSwiWC1BbXpuLVRyYWNlLUlkIjpbIlJvb3Q9MS02MTllNTJjMy0zZjEyMjQ2MTdiYmFjNmZiMDlhMjA3YWEiXSwiWC1Gb3J3YXJkZWQtRm9yIjpbIjM1LjE2OS4xNTEuODEiXSwiWC1Gb3J3YXJkZWQtUG9ydCI6WyI0NDMiXSwiWC1Gb3J3YXJkZWQtUHJvdG8iOlsiaHR0cHMiXSwiYWNjZXB0LWVuY29kaW5nIjpbImd6aXAiXX0sInF1ZXJ5U3RyaW5nUGFyYW1ldGVycyI6bnVsbCwibXVsdGlWYWx1ZVF1ZXJ5U3RyaW5nUGFyYW1ldGVycyI6bnVsbCwicmVxdWVzdENvbnRleHQiOnsiYWNjb3VudElkIjoiNDc3MzYxODc3NDQ1IiwiYXBpSWQiOiJ5dGc1Z2ZrZzVrIiwiZG9tYWluTmFtZSI6Inl0ZzVnZmtnNWsuZXhlY3V0ZS1hcGkuZXUtY2VudHJhbC0xLmFtYXpvbmF3cy5jb20iLCJkb21haW5QcmVmaXgiOiJ5dGc1Z2ZrZzVrIiwiZXh0ZW5kZWRSZXF1ZXN0SWQiOiJKVUhlbWlOWEZpQUVKeUE9IiwiaHR0cE1ldGhvZCI6IlBPU1QiLCJpZGVudGl0eSI6eyJhY2Nlc3NLZXkiOm51bGwsImFjY291bnRJZCI6bnVsbCwiY2FsbGVyIjpudWxsLCJjb2duaXRvQW1yIjpudWxsLCJjb2duaXRvQXV0aGVudGljYXRpb25Qcm92aWRlciI6bnVsbCwiY29nbml0b0F1dGhlbnRpY2F0aW9uVHlwZSI6bnVsbCwiY29nbml0b0lkZW50aXR5SWQiOm51bGwsImNvZ25pdG9JZGVudGl0eVBvb2xJZCI6bnVsbCwicHJpbmNpcGFsT3JnSWQiOm51bGwsInNvdXJjZUlwIjoiMzUuMTY5LjE1MS44MSIsInVzZXIiOm51bGwsInVzZXJBZ2VudCI6IlR5cGVmb3JtIFdlYmhvb2tzIiwidXNlckFybiI6bnVsbH0sInBhdGgiOiJ0eXBlZm9ybSIsInByb3RvY29sIjoiSFRUUC8xLjEiLCJyZXF1ZXN0SWQiOiJKVUhlbWlOWEZpQUVKeUE9IiwicmVxdWVzdFRpbWUiOiIyNC9Ob3YvMjAyMToxNDo1NzowNyArMDAwMCIsInJlcXVlc3RUaW1lRXBvY2giOjE2Mzc3NjU4Mjc2NDUsInJlc291cmNlSWQiOiJQT1NUIC9zaWdudXAve3Byb3h5K30iLCJyZXNvdXJjZVBhdGgiOiIvc2lnbnVwL3twcm94eSt9Iiwic3RhZ2UiOiIkZGVmYXVsdCJ9LCJwYXRoUGFyYW1ldGVycyI6eyJwcm94eSI6InR5cGVmb3JtIn0sInN0YWdlVmFyaWFibGVzIjpudWxsLCJib2R5Ijoie1wiZXZlbnRfaWRcIjpcIjAxRk45NzZKSlQwSjUyS0paNVZXRjBUWEg3XCIsXCJldmVudF90eXBlXCI6XCJmb3JtX3Jlc3BvbnNlXCIsXCJmb3JtX3Jlc3BvbnNlXCI6e1wiZm9ybV9pZFwiOlwiUVU1d2Q3bFFcIixcInRva2VuXCI6XCJtdHRiZG5waXQzdjJ4enZwNHRrbG10dGJkbjl6NjRxZ1wiLFwibGFuZGVkX2F0XCI6XCIyMDIxLTExLTI0VDE0OjU2OjM4WlwiLFwic3VibWl0dGVkX2F0XCI6XCIyMDIxLTExLTI0VDE0OjU3OjA3WlwiLFwiaGlkZGVuXCI6e1wic291cmNlXCI6XCJcIn0sXCJkZWZpbml0aW9uXCI6e1wiaWRcIjpcIlFVNXdkN2xRXCIsXCJ0aXRsZVwiOlwiQmV0YSBTaWdudXAgRm9ybSB2MVwiLFwiZmllbGRzXCI6W3tcImlkXCI6XCIzNXhkU2t6Q3Y5cTlcIixcInRpdGxlXCI6XCJGaXJzdCB0aGluZ3MgZmlyc3QsIHdoYXQgaXMgeW91ciBuYW1lP1wiLFwidHlwZVwiOlwic2hvcnRfdGV4dFwiLFwicmVmXCI6XCIyMGJlY2I1YTA3ODBiZThmXCIsXCJwcm9wZXJ0aWVzXCI6e319LHtcImlkXCI6XCIybk5ldTd4YnNlbXhcIixcInRpdGxlXCI6XCJBbmQgeW91ciBlbWFpbCBhZGRyZXNzP1wiLFwidHlwZVwiOlwiZW1haWxcIixcInJlZlwiOlwiNzcwYzA1NzViNzdiMWNiOFwiLFwicHJvcGVydGllc1wiOnt9fSx7XCJpZFwiOlwiOWpkeHF5c2FuVEc5XCIsXCJ0aXRsZVwiOlwiTGFzdGx5LCBob3cgYmlnIGlzIHlvdXIgZGV2ZWxvcG1lbnQgb3JnYW5pc2F0aW9uP1wiLFwidHlwZVwiOlwibXVsdGlwbGVfY2hvaWNlXCIsXCJyZWZcIjpcIjdiZWQyZjQyLWY0MmQtNGFmZS05ODU0LWY5ZmJiOTMyZTViNFwiLFwicHJvcGVydGllc1wiOnt9LFwiY2hvaWNlc1wiOlt7XCJpZFwiOlwiUzFsS1JBbW1ZUm9qXCIsXCJsYWJlbFwiOlwiSnVzdCBtZVwifSx7XCJpZFwiOlwiMDVQSUNGcVVId2xkXCIsXCJsYWJlbFwiOlwiMi0xMFwifSx7XCJpZFwiOlwiNlJjbW9JSWpIbTNLXCIsXCJsYWJlbFwiOlwiMTEtMzBcIn0se1wiaWRcIjpcIm50NlE1ZzNzVGlNcFwiLFwibGFiZWxcIjpcIjMxLTcwXCJ9LHtcImlkXCI6XCIya0dFb3ZIR0lsSDNcIixcImxhYmVsXCI6XCI3MStcIn1dfV19LFwiYW5zd2Vyc1wiOlt7XCJ0eXBlXCI6XCJ0ZXh0XCIsXCJ0ZXh0XCI6XCJBbmFcIixcImZpZWxkXCI6e1wiaWRcIjpcIjM1eGRTa3pDdjlxOVwiLFwidHlwZVwiOlwic2hvcnRfdGV4dFwiLFwicmVmXCI6XCIyMGJlY2I1YTA3ODBiZThmXCJ9fSx7XCJ0eXBlXCI6XCJlbWFpbFwiLFwiZW1haWxcIjpcImFuYUBtYW50aWwuY29tXCIsXCJmaWVsZFwiOntcImlkXCI6XCIybk5ldTd4YnNlbXhcIixcInR5cGVcIjpcImVtYWlsXCIsXCJyZWZcIjpcIjc3MGMwNTc1Yjc3YjFjYjhcIn19LHtcInR5cGVcIjpcImNob2ljZVwiLFwiY2hvaWNlXCI6e1wibGFiZWxcIjpcIjItMTBcIn0sXCJmaWVsZFwiOntcImlkXCI6XCI5amR4cXlzYW5URzlcIixcInR5cGVcIjpcIm11bHRpcGxlX2Nob2ljZVwiLFwicmVmXCI6XCI3YmVkMmY0Mi1mNDJkLTRhZmUtOTg1NC1mOWZiYjkzMmU1YjRcIn19XX19XG4iLCJpc0Jhc2U2NEVuY29kZWQiOmZhbHNlfQ==`
// 	dst := make([]byte, base64.StdEncoding.DecodedLen(len(suspiciousRequest)))
// 	_, err := base64.StdEncoding.Decode(dst, []byte(suspiciousRequest))
// 	require.NoError(t, err)

// 	dst = bytes.Trim(dst, "\x00")
// 	//fmt.Printf("%v", dst)
// 	var req events.APIGatewayProxyRequest
// 	err = json.Unmarshal(dst, &req)
// 	require.NoError(t, err)

// 	var tw TypeformWebhook
// 	err = json.Unmarshal([]byte(req.Body), &tw)
// 	require.NoError(t, err)

// 	pp, _ := json.MarshalIndent(tw, "", "  ")
// 	fmt.Printf("%s \n", pp)

// 	t.Logf("%#v", tw.AsRecord("", nil))

// }
