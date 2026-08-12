import { Controller } from "@hotwired/stimulus"

export default class extends Controller {
  static targets = [ "text", "loading" ]

  connect() {
    // Find the form that contains this button
    this.form = this.element.closest("form")
    if (this.form) {
      this.boundHandleSubmit = this.handleSubmit.bind(this)
      this.form.addEventListener("submit", this.boundHandleSubmit)
    }
  }

  disconnect() {
    if (this.form && this.boundHandleSubmit) {
      this.form.removeEventListener("submit", this.boundHandleSubmit)
    }
  }

  handleSubmit(event) {
    // Disable the button
    this.element.disabled = true

    // Hide text and show loading
    if (this.hasTextTarget) {
      this.textTarget.classList.add("hidden")
    }
    if (this.hasLoadingTarget) {
      this.loadingTarget.classList.remove("hidden")
    }
  }
}
