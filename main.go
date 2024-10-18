// ---------------------------------------------------------------------------------------------------------------------
// (w) 2024 by Jan Buchholz
// App startup, using Unison library (c) Richard A. Wilkes
// https://github.com/richardwilkes/unison
// ---------------------------------------------------------------------------------------------------------------------

package main

import (
	"Dropbox_REST_Client/ui"
	"github.com/richardwilkes/unison"
)

func main() {
	unison.Start(
		unison.StartupFinishedCallback(func() {
			err := ui.NewMainWindow()
			if err != nil {
				panic(err)
			}
		}),
		unison.QuitAfterLastWindowClosedCallback(func() bool {
			return true
		}),
		unison.AllowQuitCallback(func() bool {
			return ui.AllowQuitCallback()
		}),
	)
}
