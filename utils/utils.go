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
	str := base64.StdEncoding.EncodeToString(b)
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")
	processedString := reg.ReplaceAllString(str, "")
	return processedString, nil
}

func SendEmail(email, content string) error {
	smtpServer := os.Getenv("SMTP_SERVER")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("UR_MAIL")
	smtpPass := os.Getenv("PW_MAIL")

	return smtp.SendMail(smtpServer+":"+smtpPort,
		smtp.PlainAuth("", smtpUser, smtpPass, smtpServer),
		os.Getenv("HOST_MAIL"), []string{email}, []byte(content))
}

func ConvertFloat64ToInt64(value float64) int64 {
	return int64(value)
}

func ConvertFloat64ToInt32(value float64) int32 {
	return int32(value)
}

func Filter[T any](slice []T, condition func(T) bool) []T {
	var result []T
	for _, v := range slice {
		if condition(v) {
			result = append(result, v)
		}
	}
	return result
}
