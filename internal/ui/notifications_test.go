package ui

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNotificationDBusArgsSetsBtermIdentity(t *testing.T) {
	args := notificationDBusArgs("Title", "Message")

	require.Equal(t, []string{
		"call",
		"--session",
		"--dest", "org.freedesktop.Notifications",
		"--object-path", "/org/freedesktop/Notifications",
		"--method", "org.freedesktop.Notifications.Notify",
		"'bterm'",
		"0",
		"'io.github.abunjevac.bterm'",
		"'Title'",
		"'Message'",
		"[]",
		"{'desktop-entry': <'io.github.abunjevac.bterm'>}",
		"int32 -1",
	}, args)
}

func TestGVariantStringEscapesValues(t *testing.T) {
	require.Equal(t, `'can\'t \\ stop'`, gvariantString(`can't \ stop`))
}
