package bcrypt

import cryptoBcrypt "golang.org/x/crypto/bcrypt"

const DefaultCost = cryptoBcrypt.DefaultCost

func Hash(password string) (string, error) {
	return HashWithCost(password, DefaultCost)
}

func HashWithCost(password string, cost int) (string, error) {
	returnString, err := cryptoBcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(returnString), nil
}

func Compare(hash, password string) error {
	return cryptoBcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
