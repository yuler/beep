import { Controller } from "@hotwired/stimulus"

// Connects to data-controller="auto-submit"
export default class extends Controller {
  connect() {
    this.element.requestSubmit()
  }
}
