package ui

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExpandCWDArgs(t *testing.T) {
	t.Parallel()

	require.Equal(t, []string{"--folder=/tmp/project", "/tmp/project"}, expandCWDArgs([]string{"--folder={cwd}", "{cwd}"}, "/tmp/project"))
}

func TestNotificationDBusArgsSetsBtermIdentity(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	require.Equal(t, `'can\'t \\ stop'`, gvariantString(`can't \ stop`))
}
