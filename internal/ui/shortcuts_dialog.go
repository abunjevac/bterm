package ui

import (
	"strings"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/abunjevac/bterm/internal/keymap"
)

func showShortcutsDialog(parent *gtk.ApplicationWindow, layout *keymap.Layout) {
	win := gtk.NewWindow()

	win.SetTitle("Keyboard Shortcuts")
	win.SetTransientFor(&parent.Window)
	win.SetModal(true)

	applyDialogSpec(win, shortcutsDialogSpec())

	content, closeBtn := shortcutsDialogContent(buildShortcutCards(layout))

	win.SetChild(content)

	closeBtn.ConnectClicked(func() { win.Close() })

	addEscapeClose(win)

	win.Present()
}

func buildShortcutCards(layout *keymap.Layout) *gtk.Grid {
	groups := make(map[string][]keymap.LayoutEntry)

	for _, entry := range layout.Entries() {
		group := shortcutGroup(entry.Action)

		groups[group] = append(groups[group], entry)
	}

	cards := gtk.NewGrid()

	cards.SetRowSpacing(12)
	cards.SetColumnSpacing(12)
	cards.SetMarginTop(16)
	cards.SetMarginBottom(16)
	cards.SetMarginStart(20)
	cards.SetMarginEnd(20)
	cards.SetColumnHomogeneous(true)

	groupOrder := []string{"Tabs", "Panes", "Clipboard", "Appearance", "Application", "Input", "Other"} //nolint:goconst // shortcut group labels intentionally remain local.
	cardIndex := 0

	for _, group := range groupOrder {
		entries := groups[group]

		if len(entries) == 0 {
			continue
		}

		cards.Attach(shortcutCard(group, entries), cardIndex%2, cardIndex/2, 1, 1)

		cardIndex++
	}

	return cards
}

func shortcutsDialogContent(cards *gtk.Grid) (*gtk.Box, *gtk.Button) {
	title := gtk.NewLabel("Keyboard Shortcuts")

	title.SetHAlign(gtk.AlignStart)
	title.AddCSSClass("bterm-shortcuts-title")

	description := gtk.NewLabel("Current key bindings")

	description.SetHAlign(gtk.AlignStart)
	description.AddCSSClass("bterm-shortcuts-description")

	header := gtk.NewBox(gtk.OrientationVertical, 2)

	header.SetMarginTop(20)
	header.SetMarginStart(20)
	header.SetMarginEnd(20)
	header.Append(title)
	header.Append(description)

	closeBtn := gtk.NewButtonWithLabel("Close")

	closeBtn.AddCSSClass("suggested-action")

	footer := gtk.NewBox(gtk.OrientationHorizontal, 8)

	footer.SetHAlign(gtk.AlignEnd)
	footer.SetMarginTop(4)
	footer.SetMarginBottom(12)
	footer.SetMarginStart(12)
	footer.SetMarginEnd(12)
	footer.Append(closeBtn)

	scroll := gtk.NewScrolledWindow()

	scroll.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scroll.SetVExpand(true)
	scroll.SetChild(cards)

	content := gtk.NewBox(gtk.OrientationVertical, 0)

	content.Append(header)
	content.Append(scroll)
	content.Append(footer)

	return content, closeBtn
}

func shortcutCard(title string, entries []keymap.LayoutEntry) *gtk.Box {
	card := gtk.NewBox(gtk.OrientationVertical, 4)

	card.AddCSSClass("bterm-shortcut-card")

	heading := gtk.NewLabel(title)

	heading.SetHAlign(gtk.AlignStart)
	heading.AddCSSClass("bterm-shortcut-card-title")
	card.Append(heading)

	for _, entry := range entries {
		row := gtk.NewBox(gtk.OrientationHorizontal, 12)

		row.AddCSSClass("bterm-shortcut-row")

		nameLabel := gtk.NewLabel(humanizeAction(entry.Action))

		nameLabel.SetHAlign(gtk.AlignStart)
		nameLabel.SetHExpand(true)
		nameLabel.AddCSSClass("bterm-shortcut-action")

		keys := gtk.NewBox(gtk.OrientationHorizontal, 5)

		keys.SetHAlign(gtk.AlignEnd)

		for _, key := range entry.Keys {
			keyLabel := gtk.NewLabel(key)

			keyLabel.AddCSSClass("bterm-shortcut-key")
			keys.Append(keyLabel)
		}

		row.Append(nameLabel)
		row.Append(keys)
		card.Append(row)
	}

	return card
}

func shortcutGroup(action keymap.Action) string {
	switch action {
	case keymap.ActionNewTabEnd, keymap.ActionNewTabAfter, keymap.ActionCloseTab,
		keymap.ActionTab1, keymap.ActionTab2, keymap.ActionTab3, keymap.ActionTab4,
		keymap.ActionTab5, keymap.ActionTab6, keymap.ActionTab7, keymap.ActionTab8, keymap.ActionTab9:
		return "Tabs"
	case keymap.ActionSplitLeftRight, keymap.ActionSplitTopBottom, keymap.ActionResizeLeft,
		keymap.ActionResizeRight, keymap.ActionResizeUp, keymap.ActionResizeDown,
		keymap.ActionFocusLeft, keymap.ActionFocusRight, keymap.ActionFocusUp,
		keymap.ActionFocusDown, keymap.ActionClosePane:
		return "Panes"
	case keymap.ActionCopy, keymap.ActionPaste:
		return "Clipboard"
	case keymap.ActionFontInc, keymap.ActionFontDec, keymap.ActionFontReset:
		return "Appearance"
	case keymap.ActionOpenConfig, keymap.ActionNewWindow, keymap.ActionCloseWindow,
		keymap.ActionOpenEditor, keymap.ActionOpenFileBrowser:
		return "Application"
	case keymap.ActionClear, keymap.ActionReset, keymap.ActionSendNewline:
		return "Input"
	default:
		return "Other"
	}
}

func humanizeAction(a keymap.Action) string {
	words := strings.Split(a.String(), "_")

	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}

	return strings.Join(words, " ")
}
