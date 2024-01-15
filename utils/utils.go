package utils

import (
	"crypto/rand"
	"encoding/base64"
	"net/smtp"
	"os"
	"regexp"
	"strings"
)

// GetShopIdFromString extracts the shop ID from a given URL.
// It takes a string parameter 'url' representing the URL from which the shop ID needs to be extracted.
// It returns a string representing the extracted shop ID.
func GetShopIdFromString(url string) string {
	// Create a regular expression to find the number
	re := regexp.MustCompile(`i\.(\d+)\.`)

	// Find the number in the URL
	match := re.FindStringSubmatch(url)

	if len(match) > 1 {
		return match[1]
	} else {
		return ""
	}
}

// CreateUrlFromIdImage creates a URL from the given ID.
// It takes a string parameter 'id' representing the ID from which the URL needs to be created.
// It returns a string representing the created URL.
func CreateUrlFromIdImage(id string) string {
	return "https://down-vn.img.susercontent.com/file/" + id
}

func CreateListImageFromIds(ids []interface{}) []string {
	listImage := []string{}
	for _, id := range ids {
		listImage = append(listImage, CreateUrlFromIdImage(id.(string)))
	}
	return listImage
}

func GetProductIDFromUrl(url string) string {
	re := regexp.MustCompile(`i\.(\d+\.\d+)`)

	// Find the match.
	matches := re.FindStringSubmatch(url)

	both := matches[len(matches)-1]
	str := strings.Split(both, ".")
	if len(str) > 1 {
		return str[1]
	} else {
		return ""
	}
}

func GenerateTokenVerifyEmail() (string, error) {
	b := make([]byte, 32) //change the size to make the token longer or shorter
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func SendVerificationEmail(email, token string) error {
	smtpServer := os.Getenv("SMTP_SERVER")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("UR_MAIL")
	smtpPass := os.Getenv("PW_MAIL")

	hostMail := os.Getenv("HOST_MAIL")

	msg := "From: " + hostMail + "\n" +
		"To: " + email + "\n" +
		"Subject: Email Verification\n\n" +
		"Click on the following link to verify your email address: https://shopeetrackings.app/verify?token=" + token

	return smtp.SendMail(smtpServer+":"+smtpPort,
		smtp.PlainAuth("", smtpUser, smtpPass, smtpServer),
		os.Getenv("HOST_MAIL"), []string{email}, []byte(msg))
}

func ConvertFloat64ToInt(value float64) int {
	return int(value)
}
