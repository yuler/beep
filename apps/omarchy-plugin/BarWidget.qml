import QtQuick
import Quickshell
import Quickshell.Io
import qs.Commons
import qs.Ui

BarWidget {
  id: root
  moduleName: "beep"

  property bool runnerRunning: false
  property bool channelRunning: false
  property string watcherError: ""

  readonly property bool opened: panelLoader.item ? panelLoader.item.opened === true : false
  readonly property bool popoutSwitchClosing: panelLoader.item ? panelLoader.item.popoutSwitchClosing === true : false

  function open() {
    if (panelLoader.item) panelLoader.item.open()
  }

  function close() {
    if (panelLoader.item) panelLoader.item.close()
  }

  function togglePanel() {
    if (panelLoader.item) panelLoader.item.toggle()
  }

  function closeForPopoutSwitch() {
    if (panelLoader.item) panelLoader.item.closeForPopoutSwitch()
  }

  function injectPanel() {
    var target = panelLoader.item
    if (!target) return
    if ("bar" in target) target.bar = root.bar
    if ("settings" in target) target.settings = root.settings
    if ("anchorItem" in target) target.anchorItem = button
    if ("hostWidget" in target) target.hostWidget = root
    if ("runnerRunning" in target) target.runnerRunning = root.runnerRunning
    if ("channelRunning" in target) target.channelRunning = root.channelRunning
  }

  implicitWidth: button.implicitWidth
  implicitHeight: button.implicitHeight

  onBarChanged: injectPanel()
  onSettingsChanged: injectPanel()
  onRunnerRunningChanged: injectPanel()
  onChannelRunningChanged: injectPanel()

  readonly property string pluginDir: {
    var here = Qt.resolvedUrl(".").toString()
    return here.replace(/^file:\/\//, "").replace(/\/$/, "")
  }

  Timer {
    id: restartTimer
    interval: 5000
    repeat: false
    onTriggered: {
      if (!watcherProc.running) {
        watcherProc.running = true
      }
    }
  }

  Process {
    id: watcherProc
    command: ["python3", "-u", root.pluginDir + "/watcher.py"]
    running: true
    stdout: SplitParser {
      onRead: function(line) {
        var trimmed = String(line || "").trim()
        if (!trimmed) return
        if (trimmed.startsWith("error:")) {
          root.watcherError = trimmed.substring(6).trim()
          return
        }
        var parts = trimmed.split(":")
        if (parts.length === 2) {
          root.watcherError = ""
          var svc = parts[0]
          var active = parts[1] === "active"
          if (svc === "runner") root.runnerRunning = active
          else if (svc === "channel") root.channelRunning = active
        }
      }
    }
    onExited: function(exitCode, exitStatus) {
      if (exitCode !== 0) {
        if (!root.watcherError) {
          root.watcherError = "Watcher exited (code " + exitCode + ")"
        }
        restartTimer.start()
      }
    }
  }

  Loader {
    id: panelLoader
    active: true
    source: Qt.resolvedUrl("Panel.qml")
    visible: false
    onLoaded: {
      root.injectPanel()
      Qt.callLater(root.injectPanel)
    }
  }

  IpcHandler {
    target: "beep"

    function open() { root.open() }
    function close() { root.close() }
    function show() { root.open() }
    function hide() { root.close() }
    function toggle() { root.togglePanel() }
  }

  BarIconButton {
    id: button
    anchors.fill: parent
    bar: root.bar
    text: ""
    active: false
    iconComponent: Component {
      BeepIcon {
        iconSize: Style.bar.iconCanvas
        color: button.foreground
        runnerRunning: root.runnerRunning
        channelRunning: root.channelRunning
      }
    }
    tooltipText: {
      var base = "Beep\nRunner: " + (root.runnerRunning ? "running" : "stopped") + "\nChannel: " + (root.channelRunning ? "running" : "stopped")
      if (root.watcherError) {
        base += "\nError: " + root.watcherError
      }
      return base
    }
    onPressed: function(b) {
      root.togglePanel()
    }
  }
}
