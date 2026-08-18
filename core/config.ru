# This file is used by Rack-based servers to start the application.

require_relative "config/environment"

# Direct `bin/rails server` (outside mise) skips the root `.env`/`.env.local`.
# mise sets `MISE_TASK_NAME` for task runs; its absence means env was never
# loaded. Fail fast instead of booting broken.
if Rails.env.development? && ENV["MISE_TASK_NAME"].to_s.empty?
  abort <<~MSG
    error: not started via mise — run `mise dev` (or `mise run core:dev`).
    Direct `bin/rails server` does not load .env/.env.local.
  MSG
end

run Rails.application
Rails.application.load_server
