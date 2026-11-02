import { Controller } from "@hotwired/stimulus"

// Toggles between dark (default) and light themes by setting [data-theme] on <html>.
// The choice is persisted to localStorage with key "theme".
export default class extends Controller {
  static targets = [ "light", "dark" ]

  static KEY = "theme"

  connect() {
    this.#reflect(this.#effective())
  }

  toggle() {
    const next = this.#effective() === "dark" ? "light" : "dark"
    try {
      localStorage.setItem(this.constructor.KEY, next)
    } catch (e) {}
    document.documentElement.setAttribute("data-theme", next)
    this.#reflect(next)
  }

  #current() {
    try {
      const stored = localStorage.getItem(this.constructor.KEY)
      if (stored === "light" || stored === "dark") return stored
    } catch (e) {}

    return null
  }

  #effective() {
    return this.#current() || "dark"
  }

  #reflect(theme) {
    if (this.hasLightTarget) this.lightTarget.hidden = theme === "dark"
    if (this.hasDarkTarget) this.darkTarget.hidden = theme !== "dark"
    this.#syncFavicon(theme)
    this.#syncMetaTheme(theme)
  }

  #syncFavicon(theme) {
    const link = document.getElementById("favicon-theme")
    if (link) link.href = theme === "dark" ? "/icon-dark.svg" : "/icon-light.svg"
  }

  #syncMetaTheme(theme) {
    const meta = document.querySelector('meta[name="theme-color"]')
    if (meta) meta.setAttribute("content", theme === "dark" ? "#000000" : "#ffffff")
  }
}
