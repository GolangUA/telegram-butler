package callback

import "fmt"

const banMessageFormat = "Ваш запит відхилено. У разі помилки зверніться до адміністратора (@%s)."

func getBanMessage(admin string) string {
	return fmt.Sprintf(banMessageFormat, admin)
}

//nolint:lll
const welcomeMessageFormat = `
%s, вітаємо у спільноті <b>GolangUA</b> 🇺🇦!

Правила групи:
Code of Conduct: https://go.dev/conduct

Продуктивне спілкування вимагає зусиль, подумайте, як ваші слова будуть сприйняті. Зокрема:
	- Поважайте різницю у поглядах.
	- Користуйтеся принципами взаємоповаги.
	- Уникайте неконструктивної критики, обговорення потенційно образливих або чутливих питань. 

За порушення Code of Conduct адміни групу можуть дати попередження, відправити в мовчанку на декілька днів або заблокувати назавжди, в залежності від ситуації. 

Чекаємо тебе за посиланням @%s :)
`

func getWelcomeMessage(username, group string) string {
	return fmt.Sprintf(welcomeMessageFormat, username, group)
}
