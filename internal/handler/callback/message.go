package callback

import "fmt"

const banMessageFormat = "Ваш запит відхилено. У разі помилки зверніться до адміністратора (@%s)."

func getBanMessage(admin string) string {
	return fmt.Sprintf(banMessageFormat, admin)
}
