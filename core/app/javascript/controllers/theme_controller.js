import { Controller } from "@hotwired/stimulus"

// Toggles between light and dark themes by setting [data-theme] on <html>.
// The choice is persisted to localStorage. When nothing is stored the theme
// follows the OS preference (handled by the pre-paint script in the <head>).
export default class extends Controller {
  static targets = [ "light", "dark" ]

  static KEY = "theme"

  connect() {
    this.#reflect(this.#effective())
  }

  toggle() {
    const next = this.#current() === "dark" ? "light" : "dark"
    localStorage.setItem(this.constructor.KEY, next)
    document.documentElement.setAttribute("data-theme", next)
    this.#reflect(next)
  }

  #current() {
    const stored = localStorage.getItem(this.constructor.KEY)
    if (stored === "light" || stored === "dark") return stored

    return null
  }

  #effective() {
    return this.#current() || (window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light")
  }

  #reflect(theme) {
    if (this.hasLightTarget) this.lightTarget.hidden = theme === "dark"
    if (this.hasDarkTarget) this.darkTarget.hidden = theme !== "dark"
    this.#syncFavicon(theme)
  }

  #syncFavicon(theme) {
    const link = document.getElementById("favicon-theme")
    if (link) link.href = theme === "dark" ? "/icon-dark.svg" : "/icon-light.svg"
  }
}
