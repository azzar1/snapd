// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2018 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package builtin

import (
	"fmt"
	"os"
	"strings"

	"github.com/snapcore/snapd/interfaces"
	"github.com/snapcore/snapd/interfaces/apparmor"
)

const serviceWatchdogSummary = `allows use of systemd service watchdog`

const serviceWatchdogBaseDeclarationSlots = `
  service-watchdog:
    allow-installation:
      slot-snap-type:
        - core
    deny-auto-connection: true
`

const serviceWatchdogConnectedPlugAppArmorTemplate = `
# Allow sending notification messages to systemd through the notify socket
"{{notify-socket}}" w,
`

type serviceWatchdogInterface struct {
	commonInterface
}

var osGetenv = os.Getenv

func (iface *serviceWatchdogInterface) AppArmorConnectedPlug(spec *apparmor.Specification, plug *interfaces.ConnectedPlug, slot *interfaces.ConnectedSlot) error {
	notifySocket := osGetenv("NOTIFY_SOCKET")
	if notifySocket == "" {
		notifySocket = "/run/systemd/notify"
	} else {
		// must be an absolute path without any AppArmor regular
		// expression (AARE) characters or double quotes
		if !strings.HasPrefix(notifySocket, "/") || strings.ContainsAny(notifySocket, `?*[]{}^"`) {
			return fmt.Errorf("cannot use %q as notify socket path", notifySocket)
		}
	}
	snippet := strings.Replace(serviceWatchdogConnectedPlugAppArmorTemplate,
		"{{notify-socket}}", notifySocket, 1)
	spec.AddSnippet(snippet)
	return nil
}

func init() {
	registerIface(&serviceWatchdogInterface{commonInterface: commonInterface{
		name:                 "service-watchdog",
		summary:              serviceWatchdogSummary,
		implicitOnCore:       true,
		implicitOnClassic:    true,
		baseDeclarationSlots: serviceWatchdogBaseDeclarationSlots,
		reservedForOS:        true,
	}})
}