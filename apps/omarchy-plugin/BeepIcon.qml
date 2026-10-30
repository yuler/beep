import QtQuick
import QtQuick.Shapes
import qs.Commons
import qs.Ui

Item {
  id: root

  property real iconSize: Style.bar.iconCanvas
  property color color: Color.foreground
  property bool runnerRunning: false
  property bool channelRunning: false

  readonly property color statusColor: "#22c55e"

  width: iconSize
  height: iconSize
  implicitWidth: iconSize
  implicitHeight: iconSize

  Item {
    id: glyphContainer
    anchors.centerIn: parent
    width: 512
    height: 512
    scale: (root.iconSize * 0.95) / 512
    transformOrigin: Item.Center

    Shape {
      anchors.fill: parent
      layer.enabled: true
      layer.samples: 4

      // Bell body
      ShapePath {
        strokeColor: root.color
        strokeWidth: 30
        fillColor: "transparent"
        capStyle: ShapePath.RoundCap
        joinStyle: ShapePath.RoundJoin
        PathSvg {
          path: "M256 116 C214 116 181 148 181 190 L181 226 C181 252 170 274 151 294 C143 303 149 318 162 318 L350 318 C363 318 369 303 361 294 C342 274 331 252 331 226 L331 190 C331 148 298 116 256 116 Z"
        }
      }

      // Clapper
      ShapePath {
        strokeColor: root.color
        strokeWidth: 30
        fillColor: "transparent"
        capStyle: ShapePath.RoundCap
        PathSvg {
          path: "M232 318 C232 337 242 348 256 348 C270 348 280 337 280 318"
        }
      }

      // Eyes
      ShapePath {
        strokeColor: root.color
        strokeWidth: 30
        fillColor: "transparent"
        capStyle: ShapePath.RoundCap
        PathSvg {
          path: "M225 202 L225 232 M287 202 L287 232"
        }
      }

      // Sound wave inner
      ShapePath {
        strokeColor: root.color
        strokeWidth: 26
        fillColor: "transparent"
        capStyle: ShapePath.RoundCap
        PathSvg {
          path: "M365 151 C386 166 398 188 398 214"
        }
      }

      // Sound wave outer
      ShapePath {
        strokeColor: root.color
        strokeWidth: 26
        fillColor: "transparent"
        capStyle: ShapePath.RoundCap
        PathSvg {
          path: "M390 119 C424 141 444 174 444 214"
        }
      }

      // Base bar
      ShapePath {
        strokeColor: root.color
        strokeWidth: 30
        fillColor: "transparent"
        capStyle: ShapePath.RoundCap
        PathSvg {
          path: "M226 388 L286 388"
        }
      }
    }
  }

  Rectangle {
    readonly property real markSize: Math.max(4, Math.round(root.iconSize * 0.22))

    anchors.top: parent.top
    anchors.right: parent.right
    width: markSize
    height: markSize
    radius: markSize / 2
    color: root.statusColor
  }
}
