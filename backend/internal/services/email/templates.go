package email

import (
	"fmt"
	"strings"
)

// template holds the rendered bits of an outgoing email. We deliberately
// keep both the plain-text and HTML versions — most clients prefer HTML
// but the plain-text fallback is what spam filters actually read.
type template struct {
	subject string
	text    string
	html    string
}

// renderWelcome produces the welcome email for the given user. The locale
// picks one of the four supported languages (en/es/ja/id); unknown locales
// fall back to English so we never send a half-translated email.
func renderWelcome(userEmail, locale string) template {
	t := welcomeCopy(locale)
	return template{
		subject: t.subjectWelcome,
		text: fmt.Sprintf(
			"%s\n\n%s\n%s\n\n— %s\n",
			t.subjectWelcome,
			t.line1(userEmail),
			t.line2,
			t.signOff,
		),
		html: fmt.Sprintf(
			`<!doctype html><html><body style="font-family:sans-serif">
<p>%s</p>
<p>%s</p>
<p>%s</p>
<p>— %s</p>
</body></html>`,
			t.subjectWelcome,
			t.line1(userEmail),
			t.line2,
			t.signOff,
		),
	}
}

// renderReset produces the 6-digit code email. The code
// is rendered with monospace and zero width padding so it survives copy-
// paste into a 6-digit form field on desktop or mobile.
func renderReset(code, locale string) template {
	t := welcomeCopy(locale) // same copy object — fields are shared
	return template{
		subject: t.subjectReset,
		text: fmt.Sprintf(
			"%s\n\n%s: %s\n\n%s\n%s\n\n— %s\n",
			t.subjectReset,
			t.codeLabel, code,
			t.lineExpires,
			t.lineIgnore,
			t.signOff,
		),
		html: fmt.Sprintf(
			`<!doctype html><html><body style="font-family:sans-serif">
<h1 style="margin-bottom:0">%s</h1>
<p>%s: <strong style="font-size:24px;letter-spacing:0.4em;font-family:monospace">%s</strong></p>
<p style="color:#666">%s</p>
<p style="color:#666">%s</p>
<p>— %s</p>
</body></html>`,
			t.subjectReset,
			t.codeLabel, code,
			t.lineExpires,
			t.lineIgnore,
			t.signOff,
		),
	}
}

// copy is the localised string bundle. Field naming stays lowercase because
// the strings themselves carry the case they need; this struct is the
// assembly line.
type copy struct {
	subjectWelcome string
	subjectReset   string
	line1          func(userEmail string) string
	line2          string
	codeLabel      string
	lineExpires    string
	lineIgnore     string
	signOff        string
}

// welcomeCopy returns the localised bundle for `locale`. Unknown locales
// fall back to English (Spec 017 §8.4 — supported locales are en/es/ja/id).
func welcomeCopy(locale string) copy {
	if c, ok := welcome[normaliseLocale(locale)]; ok {
		return c
	}
	return welcome["en"]
}

// normaliseLocale trims/lower-cases and accepts the canonical four codes.
// Anything else falls back to English via welcomeCopy.
func normaliseLocale(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "en", "es", "ja", "id":
		return v
	}
	return "en"
}

// welcome is the single source of truth for all four locales. Keep the
// entries in alphabetical order by key.
var welcome = map[string]copy{
	"en": {
		subjectWelcome: "Welcome to Itinera",
		subjectReset:   "Reset your Itinera password",
		line1:          func(email string) string { return "Hi " + email + "," },
		line2:          "We're glad you're here. Start planning your first trip — your data is saved automatically.",
		codeLabel:      "Your code",
		lineExpires:    "The code expires in 1 hour.",
		lineIgnore:     "If you didn't request this, you can ignore the email.",
		signOff:        "The Itinera team",
	},
	"es": {
		subjectWelcome: "Bienvenido a Itinera",
		subjectReset:   "Restablece tu contraseña de Itinera",
		line1:          func(email string) string { return "Hola " + email + "," },
		line2:          "Nos alegra tenerte aquí. Empieza a planificar tu primer viaje — tus datos se guardan automáticamente.",
		codeLabel:      "Tu código",
		lineExpires:    "El código caduca en 1 hora.",
		lineIgnore:     "Si no has solicitado esto, puedes ignorar el correo.",
		signOff:        "El equipo de Itinera",
	},
	"ja": {
		subjectWelcome: "Itineraへようこそ",
		subjectReset:   "Itineraのパスワードを再設定",
		line1:          func(email string) string { return email + " さん、" },
		line2:          "Itineraへようこそ。最初の旅行の計画を立てましょう — データは自動的に保存されます。",
		codeLabel:      "確認コード",
		lineExpires:    "コードの有効期限は1時間です。",
		lineIgnore:     "リクエストしていない場合は、このメールを無視してください。",
		signOff:        "Itinera チーム",
	},
	"id": {
		subjectWelcome: "Selamat datang di Itinera",
		subjectReset:   "Reset kata sandi Itinera Anda",
		line1:          func(email string) string { return "Halo " + email + "," },
		line2:          "Kami senang Anda bergabung. Mulailah merencanakan perjalanan pertama Anda — data Anda tersimpan otomatis.",
		codeLabel:      "Kode Anda",
		lineExpires:    "Kode kedaluwarsa dalam 1 jam.",
		lineIgnore:     "Jika Anda tidak meminta ini, abaikan email tersebut.",
		signOff:        "Tim Itinera",
	},
}
