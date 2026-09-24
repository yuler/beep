#!/usr/bin/env python3
"""Watch beep-runner.service and beep-channel.service state changes via systemd D-Bus.

Emits on stdout:
  runner:active | runner:inactive
  channel:active | channel:inactive
on initial start and whenever systemd reports a state change.
No polling or timers are used.
"""

import os
import signal
import sys

try:
    import dbus
    import dbus.mainloop.glib
    from gi.repository import GLib
except ImportError as err:
    print(f"error:{err}", flush=True)
    sys.exit(1)

UNITS = {
    "beep-runner.service": "runner",
    "beep-channel.service": "channel",
}

PATH_TO_NAME = {}
for u, short in UNITS.items():
    escaped = u.replace("-", "_2d").replace(".", "_2e")
    PATH_TO_NAME["/org/freedesktop/systemd1/unit/" + escaped] = (u, short)

dbus.mainloop.glib.DBusGMainLoop(set_as_default=True)

try:
    bus = dbus.SessionBus()
    manager_obj = bus.get_object("org.freedesktop.systemd1", "/org/freedesktop/systemd1")
    manager = dbus.Interface(manager_obj, "org.freedesktop.systemd1.Manager")
    manager.Subscribe()
except Exception as err:
    print(f"error:{err}", flush=True)
    sys.exit(1)

current_states = {}

def get_unit_state(unit_name):
    try:
        unit_path = manager.GetUnit(unit_name)
        unit_obj = bus.get_object("org.freedesktop.systemd1", unit_path)
        props = dbus.Interface(unit_obj, "org.freedesktop.DBus.Properties")
        state = str(props.Get("org.freedesktop.systemd1.Unit", "ActiveState"))
        return state
    except Exception:
        return "inactive"

def emit(short, state):
    simplified = "active" if state == "active" else "inactive"
    if current_states.get(short) != simplified:
        current_states[short] = simplified
        print(f"{short}:{simplified}", flush=True)

# Initial state query
for u, short in UNITS.items():
    emit(short, get_unit_state(u))

def on_properties_changed(interface, changed_props, invalidated_props, path=None):
    if path in PATH_TO_NAME and "ActiveState" in changed_props:
        unit, short = PATH_TO_NAME[path]
        emit(short, str(changed_props["ActiveState"]))

bus.add_signal_receiver(
    on_properties_changed,
    signal_name="PropertiesChanged",
    dbus_interface="org.freedesktop.DBus.Properties",
    path_keyword="path",
)

def on_job_removed(job_id, job, unit, result):
    if unit in UNITS:
        short = UNITS[unit]
        emit(short, get_unit_state(unit))

bus.add_signal_receiver(
    on_job_removed,
    signal_name="JobRemoved",
    dbus_interface="org.freedesktop.systemd1.Manager",
)

def on_unit_files_changed():
    for u, short in UNITS.items():
        emit(short, get_unit_state(u))

bus.add_signal_receiver(
    on_unit_files_changed,
    signal_name="UnitFilesChanged",
    dbus_interface="org.freedesktop.systemd1.Manager",
)

loop = GLib.MainLoop()

def on_sigterm(signum, frame):
    loop.quit()

signal.signal(signal.SIGTERM, on_sigterm)
signal.signal(signal.SIGINT, on_sigterm)

try:
    loop.run()
except (KeyboardInterrupt, SystemExit):
    pass
