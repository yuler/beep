import QtQuick
import QtQuick.Controls
import QtQuick.Layouts
import Quickshell
import Quickshell.Io
import qs.Commons
import qs.Ui

Panel {
  id: root
  moduleName: "beep"
  ipcTarget: "beep"
  manageIpc: false

  property var anchorItem: null
  property var hostWidget: null
  property bool runnerRunning: false
  property bool channelRunning: false

  property var userData: ({ userName: "", userEmail: "", accountSlug: "", serverUrl: "", isLoggedIn: false })
  property var beeps: []
  property int beepsTotal: 0
  property bool beepsLoading: false
  property var jobs: []
  property int jobsTotal: 0
  property bool jobsLoading: false

  readonly property var barIdentity: hostWidget || root
  readonly property color foreground: bar ? bar.foreground : Color.foreground
  readonly property color dim: Qt.darker(foreground, 1.45)
  readonly property string fontFamily: bar ? bar.fontFamily : Style.font.family

  function alpha(c, a) {
    return Qt.rgba(c.r, c.g, c.b, a)
  }

  function open() {
    fetchData()
    root.controller.show()
  }

  function close() {
    root.controller.hide()
  }

  function toggle() {
    if (root.opened) root.close()
    else root.open()
  }

  function switchPanel(direction) {
    if (root.bar && typeof root.bar.switchPanelFrom === "function")
      return root.bar.switchPanelFrom(root.barIdentity, direction)
    return false
  }

  readonly property string pluginDir: {
    var here = Qt.resolvedUrl(".").toString()
    return here.replace(/^file:\/\//, "").replace(/\/$/, "")
  }

  function webOrigin() {
    var server = String(root.userData.serverUrl || "").replace(/\/$/, "")
    var match = server.match(/^(https?:\/\/)core\.([^/:]+)(?::(\d+))?/i)
    if (!match) return server
    var port = match[3] === "3001" ? "3000" : match[3]
    return match[1] + "web." + match[2] + (port ? ":" + port : "")
  }

  function accountWebUrl(path) {
    var origin = root.webOrigin()
    var slug = root.userData.accountSlug || ""
    if (!origin || !slug || !path) return ""
    return origin + "/" + slug + path
  }

  function beepUrl(beep) {
    var id = beep && beep.id ? beep.id : ""
    return root.accountWebUrl(id ? "/beeps/" + id : "")
  }

  function jobUrl(job) {
    if (!job) return ""
    if (job.runnerId) return root.accountWebUrl("/runners/" + job.runnerId)
    return root.accountWebUrl("/runners")
  }

  function notifyAt(beep) {
    if (!beep) return ""
    return beep.next_run_at || beep.run_at || ""
  }

  function formatNotifyTime(iso) {
    if (!iso) return "—"
    var date = new Date(iso)
    if (isNaN(date.getTime())) return "—"
    return Qt.formatDateTime(date, "MM-dd HH:mm")
  }

  function linkBeeps(items) {
    var out = []
    for (var i = 0; i < items.length; i++) {
      var beep = items[i] || {}
      out.push({
        id: beep.id || "",
        title: beep.title || "",
        next_run_at: beep.next_run_at || "",
        run_at: beep.run_at || "",
        url: root.beepUrl(beep)
      })
    }
    return out
  }

  function linkJobs(items) {
    var out = []
    for (var i = 0; i < items.length; i++) {
      var job = items[i] || {}
      var linked = {
        slug: job.slug || "",
        id: job.id || "",
        runnerId: job.runnerId || "",
        name: job.name || job.slug || "",
        url: ""
      }
      linked.url = root.jobUrl(linked)
      out.push(linked)
    }
    return out
  }

  function openExternal(url) {
    var target = String(url || "").trim()
    if (!target) return
    Util.execArgv(["xdg-open", target])
    root.close()
  }

  function fetchData() {
    userProc.running = false
    userProc.running = true

    // 2. Beeps (active list)
    root.beepsLoading = true
    beepsProc.running = false
    beepsProc.running = true

    // 3. Runner jobs
    root.jobsLoading = true
    jobsProc.running = false
    jobsProc.running = true
  }

  onOpenedChanged: {
    if (root.opened) {
      root.fetchData()
    }
  }

  Process {
    id: userProc
    command: ["python3", "-u", root.pluginDir + "/session.py"]
    running: false
    stdout: StdioCollector {
      waitForEnd: true
      onStreamFinished: {
        var trimmed = String(text || "").trim()
        if (!trimmed) return
        try {
          var data = JSON.parse(trimmed)
          root.userData = {
            userName: data.UserName || "",
            userEmail: data.UserEmail || "",
            accountSlug: data.AccountSlug || "",
            serverUrl: data.ServerURL || "",
            isLoggedIn: data.LoggedIn === true
          }
          if (root.beeps.length > 0) root.beeps = root.linkBeeps(root.beeps)
          if (root.jobs.length > 0) root.jobs = root.linkJobs(root.jobs)
        } catch (e) {
          root.userData = { userName: "", userEmail: "", accountSlug: "", serverUrl: "", isLoggedIn: false }
        }
      }
    }
  }

  Process {
    id: beepsProc
    command: ["beep", "list", "--json"]
    running: false
    stdout: StdioCollector {
      waitForEnd: true
      onStreamFinished: {
        var trimmed = String(text || "").trim()
        if (!trimmed) return
        root.beepsLoading = false
        try {
          var items = JSON.parse(trimmed)
          if (Array.isArray(items)) {
            root.beepsTotal = items.length
            var sorted = items.slice().sort(function(a, b) {
              var ta = root.notifyAt(a)
              var tb = root.notifyAt(b)
              var da = ta ? new Date(ta).getTime() : Number.POSITIVE_INFINITY
              var db = tb ? new Date(tb).getTime() : Number.POSITIVE_INFINITY
              if (isNaN(da)) da = Number.POSITIVE_INFINITY
              if (isNaN(db)) db = Number.POSITIVE_INFINITY
              return da - db
            })
            root.beeps = root.linkBeeps(sorted.slice(0, 5))
          } else {
            root.beepsTotal = 0
            root.beeps = []
          }
        } catch (e) {
          root.beepsTotal = 0
          root.beeps = []
        }
      }
    }
  }

  Process {
    id: jobsProc
    command: ["beep", "runner", "job", "list", "--json"]
    running: false
    stdout: StdioCollector {
      waitForEnd: true
      onStreamFinished: {
        var trimmed = String(text || "").trim()
        if (!trimmed) return
        root.jobsLoading = false
        try {
          var items = JSON.parse(trimmed)
          if (Array.isArray(items)) {
            root.jobsTotal = items.length
            var top = items.slice(0, 5)
            var jobs = []
            for (var i = 0; i < top.length; i++) {
              var item = top[i]
              var serverJob = item.ServerJob || null
              var localJob = item.LocalJob || null
              var slug = item.Slug || (localJob && localJob.slug) || (serverJob && serverJob.slug) || ""
              if (!slug) continue
              jobs.push({
                slug: slug,
                id: item.ID || (serverJob && serverJob.id) || (localJob && localJob.id) || "",
                runnerId: (serverJob && serverJob.runner_id) || "",
                name: (serverJob && serverJob.name) || (localJob && localJob.name) || slug
              })
            }
            root.jobs = root.linkJobs(jobs)
          } else {
            root.jobsTotal = 0
            root.jobs = []
          }
        } catch (e) {
          root.jobsTotal = 0
          root.jobs = []
        }
      }
    }
  }

  KeyboardPanel {
    id: panel
    anchorItem: root.anchorItem
    owner: root.barIdentity
    bar: root.bar
    open: root.opened
    focusTarget: keyCatcher
    contentWidth: panel.fittedContentWidth(Style.space(380))
    contentHeight: panel.fittedContentHeight(contentCol.implicitHeight, Style.space(560))

    PanelKeyCatcher {
      id: keyCatcher
      anchors.fill: parent
      onCloseRequested: root.close()
      onTabRequested: function(direction) { root.switchPanel(direction) }

      Flickable {
        id: flickable
        anchors.fill: parent
        contentWidth: width
        contentHeight: contentCol.implicitHeight
        clip: true
        boundsBehavior: Flickable.StopAtBounds
        ScrollBar.vertical: ScrollBar { policy: ScrollBar.AsNeeded }

        Column {
          id: contentCol
          width: flickable.width
          spacing: Style.space(12)

          // ---------- Hero: App / User Overview ----------
          PanelHero {
            width: parent.width
            title: root.userData.userName || (root.userData.isLoggedIn ? "Beep User" : "Beep")
            meta: root.userData.userEmail || (root.userData.isLoggedIn ? "" : "Not logged in")
            detail: root.userData.accountSlug || ""
            foreground: root.foreground
            fontFamily: root.fontFamily
            iconComponent: Component {
              BeepIcon {
                iconSize: Style.font.display
                color: root.foreground
                runnerRunning: root.runnerRunning
                channelRunning: root.channelRunning
              }
            }
          }

          // ---------- User Block ----------
          PanelSeparator {
            foreground: root.foreground
          }

          PanelSectionHeader {
            text: "USER"
            foreground: root.foreground
            fontFamily: root.fontFamily
          }

          Column {
            width: parent.width
            spacing: Style.space(6)

            Row {
              width: parent.width
              Text {
                text: "Name"
                color: root.dim
                font.family: root.fontFamily
                font.pixelSize: Style.font.bodySmall
                font.bold: true
                width: Style.space(70)
              }
              Text {
                text: root.userData.userName || "—"
                color: root.foreground
                font.family: root.fontFamily
                font.pixelSize: Style.font.bodySmall
                elide: Text.ElideRight
                width: parent.width - Style.space(70)
              }
            }

            Row {
              width: parent.width
              Text {
                text: "Email"
                color: root.dim
                font.family: root.fontFamily
                font.pixelSize: Style.font.bodySmall
                font.bold: true
                width: Style.space(70)
              }
              Text {
                text: root.userData.userEmail || "—"
                color: root.foreground
                font.family: root.fontFamily
                font.pixelSize: Style.font.bodySmall
                elide: Text.ElideRight
                width: parent.width - Style.space(70)
              }
            }

            Row {
              width: parent.width
              Text {
                text: "Account"
                color: root.dim
                font.family: root.fontFamily
                font.pixelSize: Style.font.bodySmall
                font.bold: true
                width: Style.space(70)
              }
              Text {
                text: root.userData.accountSlug || "—"
                color: root.foreground
                font.family: root.fontFamily
                font.pixelSize: Style.font.bodySmall
                elide: Text.ElideRight
                width: parent.width - Style.space(70)
              }
            }
          }

          // ---------- Services Block ----------
          PanelSeparator {
            foreground: root.foreground
          }

          PanelSectionHeader {
            text: "SERVICES"
            foreground: root.foreground
            fontFamily: root.fontFamily
          }

          Column {
            width: parent.width
            spacing: Style.space(8)

            // Runner row
            Row {
              width: parent.width
              spacing: Style.space(10)

              Item {
                width: Style.space(16)
                height: Style.space(16)
                anchors.verticalCenter: parent.verticalCenter

                Rectangle {
                  visible: root.runnerRunning
                  anchors.centerIn: parent
                  width: Style.space(7)
                  height: Style.space(7)
                  radius: width / 2
                  color: "#22c55e"
                }

                Rectangle {
                  visible: !root.runnerRunning
                  anchors.centerIn: parent
                  width: Style.space(10)
                  height: Style.space(2)
                  radius: height / 2
                  color: "#ef4444"
                  rotation: -45
                }
              }

              Text {
                text: "Runner"
                color: root.foreground
                font.family: root.fontFamily
                font.pixelSize: Style.font.body
                font.bold: true
                anchors.verticalCenter: parent.verticalCenter
              }

              Item {
                width: Math.max(1, parent.width - parent.children[0].width - parent.children[1].width - runnerStatusText.width - Style.space(30))
                height: 1
              }

              Text {
                id: runnerStatusText
                text: root.runnerRunning ? "Running" : "Not running"
                color: root.runnerRunning ? "#22c55e" : "#ef4444"
                font.family: root.fontFamily
                font.pixelSize: Style.font.bodySmall
                font.bold: true
                anchors.verticalCenter: parent.verticalCenter
              }
            }

            // Channel row
            Row {
              width: parent.width
              spacing: Style.space(10)

              Item {
                width: Style.space(16)
                height: Style.space(16)
                anchors.verticalCenter: parent.verticalCenter

                Rectangle {
                  visible: root.channelRunning
                  anchors.centerIn: parent
                  width: Style.space(7)
                  height: Style.space(7)
                  radius: width / 2
                  color: "#22c55e"
                }

                Rectangle {
                  visible: !root.channelRunning
                  anchors.centerIn: parent
                  width: Style.space(10)
                  height: Style.space(2)
                  radius: height / 2
                  color: "#ef4444"
                  rotation: -45
                }
              }

              Text {
                text: "Channel"
                color: root.foreground
                font.family: root.fontFamily
                font.pixelSize: Style.font.body
                font.bold: true
                anchors.verticalCenter: parent.verticalCenter
              }

              Item {
                width: Math.max(1, parent.width - parent.children[0].width - parent.children[1].width - channelStatusText.width - Style.space(30))
                height: 1
              }

              Text {
                id: channelStatusText
                text: root.channelRunning ? "Running" : "Not running"
                color: root.channelRunning ? "#22c55e" : "#ef4444"
                font.family: root.fontFamily
                font.pixelSize: Style.font.bodySmall
                font.bold: true
                anchors.verticalCenter: parent.verticalCenter
              }
            }
          }

          // ---------- Beeps Block ----------
          PanelSeparator {
            foreground: root.foreground
          }

          PanelSectionHeader {
            text: "BEEPS (" + root.beepsTotal + ")"
            foreground: root.foreground
            fontFamily: root.fontFamily
          }

          Text {
            visible: root.beepsLoading
            text: "Loading beeps..."
            color: root.dim
            font.family: root.fontFamily
            font.pixelSize: Style.font.bodySmall
          }

          Text {
            visible: !root.beepsLoading && root.beeps.length === 0
            text: "No active beeps"
            color: root.dim
            font.family: root.fontFamily
            font.pixelSize: Style.font.bodySmall
          }

          Column {
            visible: !root.beepsLoading && root.beeps.length > 0
            width: parent.width
            spacing: Style.space(6)

            Repeater {
              model: root.beeps

              Rectangle {
                required property var modelData
                width: parent.width
                height: beepRow.implicitHeight
                radius: Style.cornerRadius
                color: beepLink.containsMouse ? root.alpha(root.foreground, 0.08) : "transparent"

                Row {
                  id: beepRow
                  width: parent.width
                  spacing: Style.space(8)

                  Text {
                    text: modelData.title || "Untitled"
                    color: root.foreground
                    font.family: root.fontFamily
                    font.pixelSize: Style.font.bodySmall
                    elide: Text.ElideRight
                    width: Math.max(0, parent.width - beepTime.width - parent.spacing)
                    anchors.verticalCenter: parent.verticalCenter
                  }

                  Text {
                    id: beepTime
                    text: root.formatNotifyTime(modelData.next_run_at || modelData.run_at || "")
                    color: root.dim
                    font.family: root.fontFamily
                    font.pixelSize: Style.font.caption
                    anchors.verticalCenter: parent.verticalCenter
                  }
                }

                MouseArea {
                  id: beepLink
                  anchors.fill: parent
                  hoverEnabled: true
                  preventStealing: true
                  cursorShape: Qt.PointingHandCursor
                  onClicked: root.openExternal(modelData.url)
                }
              }
            }
          }

          // ---------- Jobs Block ----------
          PanelSeparator {
            foreground: root.foreground
          }

          PanelSectionHeader {
            text: "JOBS (" + root.jobsTotal + ")"
            foreground: root.foreground
            fontFamily: root.fontFamily
          }

          Text {
            visible: root.jobsLoading
            text: "Loading jobs..."
            color: root.dim
            font.family: root.fontFamily
            font.pixelSize: Style.font.bodySmall
          }

          Text {
            visible: !root.jobsLoading && root.jobs.length === 0
            text: "No runner jobs"
            color: root.dim
            font.family: root.fontFamily
            font.pixelSize: Style.font.bodySmall
          }

          Column {
            visible: !root.jobsLoading && root.jobs.length > 0
            width: parent.width
            spacing: Style.space(6)

            Repeater {
              model: root.jobs

              Rectangle {
                required property var modelData
                width: parent.width
                height: jobRow.implicitHeight
                radius: Style.cornerRadius
                color: jobLink.containsMouse ? root.alpha(root.foreground, 0.08) : "transparent"

                Row {
                  id: jobRow
                  width: parent.width
                  spacing: Style.space(8)

                  Text {
                    text: "•"
                    color: root.dim
                    font.family: root.fontFamily
                    font.pixelSize: Style.font.caption
                    anchors.verticalCenter: parent.verticalCenter
                  }

                  Text {
                    text: modelData.name || modelData.slug
                    color: root.foreground
                    font.family: root.fontFamily
                    font.pixelSize: Style.font.caption
                    elide: Text.ElideRight
                    width: Math.max(0, parent.width - Style.space(16))
                    anchors.verticalCenter: parent.verticalCenter
                  }
                }

                MouseArea {
                  id: jobLink
                  anchors.fill: parent
                  hoverEnabled: true
                  preventStealing: true
                  cursorShape: Qt.PointingHandCursor
                  onClicked: root.openExternal(modelData.url)
                }
              }
            }
          }
        }
      }
    }
  }
}
