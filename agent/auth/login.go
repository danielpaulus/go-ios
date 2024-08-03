package auth

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func LoginIfNeeded() error {
	port := os.Getenv("API_KEY")
	accountID := os.Getenv("ACCOUNT_ID")
	if port == "" || accountID == "" {
		print("logging in, please click the link below")
		account, err := AuthorizeUser()
		if err != nil {
			return err
		}
		persist(account)
		os.Setenv("API_KEY", account.ApiKey.String())
		os.Setenv("ACCOUNT_ID", account.Id)
		print("log in ok, account:" + account.Id)
	}
	return nil
}

func persist(account Account) {
	// Read the existing .env file
	envMap := make(map[string]string)
	file, err := os.Open(".env")
	if err == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				envMap[parts[0]] = parts[1]
			}
		}
		file.Close()
	}

	// Update the map with new values
	envMap["API_KEY"] = account.ApiKey.String()
	envMap["ACCOUNT_ID"] = account.Id

	// Write the updated map back to the .env file
	file, err = os.OpenFile(".env", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("Error opening .env file: %v\n", err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for key, value := range envMap {
		_, err := writer.WriteString(fmt.Sprintf("%s=%s\n", key, value))
		if err != nil {
			fmt.Printf("Error writing to .env file: %v\n", err)
			return
		}
	}
	writer.Flush()
}
