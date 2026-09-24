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

  Process {
    id: watcherProc
    command: ["python3", "-u", root.pluginDir + "/watcher.py"]
    running: true
    stdout: SplitParser {
      onRead: function(line) {
        var trimmed = String(line || "").trim()
        if (!trimmed) return
        var parts = trimmed.split(":")
        if (parts.length === 2) {
          var svc = parts[0]
          var active = parts[1] === "active"
          if (svc === "runner") root.runnerRunning = active
          else if (svc === "channel") root.channelRunning = active
        }
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

    function open(): void { root.open() }
    function close(): void { root.close() }
    function show(): void { root.open() }
    function hide(): void { root.close() }
    function toggle(): void { root.togglePanel() }
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
    tooltipText: "Beep\nRunner: " + (root.runnerRunning ? "running" : "stopped") + "\nChannel: " + (root.channelRunning ? "running" : "stopped")
    onPressed: function(b) {
      root.togglePanel()
    }
  }
}
