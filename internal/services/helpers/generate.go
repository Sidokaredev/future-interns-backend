package helper_services

import "github.com/google/uuid"

func GenerateUUID(text string) string {
	namespace := uuid.Must(uuid.NewRandom())
	content := []byte(text)

	SHA1 := uuid.NewSHA1(namespace, content)

	return SHA1.String()
}
