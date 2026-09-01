import { Controller } from "@hotwired/stimulus"

// Handles switching the user's preferred language between en and zh.
// Sets the shared PARAGLIDE_LOCALE cookie across the application.
export default class extends Controller {
  connect() {
    this.closeOnOutsideClick = this.closeOnOutsideClick.bind(this)
    document.addEventListener("click", this.closeOnOutsideClick)
  }

  disconnect() {
    document.removeEventListener("click", this.closeOnOutsideClick)
  }

  closeOnOutsideClick(event) {
    if (this.element.hasAttribute("open") && !this.element.contains(event.target)) {
      this.element.removeAttribute("open")
    }
  }

  switch(event) {
    event.preventDefault()
    const locale = event.currentTarget.dataset.locale
    if (!locale) return

    const maxAge = 34560000
    document.cookie = `PARAGLIDE_LOCALE=${locale}; path=/; max-age=${maxAge}; SameSite=Lax`

    const url = new URL(window.location.href)
    url.searchParams.delete("locale")
    window.location.assign(url.toString())
  }
}
