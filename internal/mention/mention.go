package mention

import (
	"fmt"
	"html"

	"github.com/mymmrac/telego"
)

// HTML returns an inline mention link rendered for HTML parse mode.
// The displayed text is "FirstName (@username)" if a username is present, otherwise just FirstName.
func HTML(id int64, firstName, username string) string {
	name := html.EscapeString(firstName)
	if username != "" {
		name += " (@" + html.EscapeString(username) + ")"
	}

	return fmt.Sprintf(`<a href="tg://user?id=%d">%s</a>`, id, name)
}

// User is a convenience wrapper around HTML for a *telego.User.
func User(u *telego.User) string {
	return HTML(u.ID, u.FirstName, u.Username)
}

// ByID renders a mention when only the user ID is known. The fallback display
// is "user#<id>" if name is empty.
func ByID(id int64, name string) string {
	if name == "" {
		name = fmt.Sprintf("user#%d", id)
	}

	return fmt.Sprintf(`<a href="tg://user?id=%d">%s</a>`, id, html.EscapeString(name))
}

// DisplayName returns the conventional short label for a user — "@username" if set, else FirstName.
func DisplayName(u *telego.User) string {
	if u.Username != "" {
		return "@" + u.Username
	}

	return u.FirstName
}
