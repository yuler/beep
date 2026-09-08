class Api::V1::Runner::BaseController < ActionController::API
  include ActionController::Cookies
  include Api::V1::Responses

  before_action :authenticate_runner!

  private
    def authenticate_runner!
      token = extract_runner_token
      if token.blank?
        render_json_error(
          status: :unauthorized,
          message: "Missing runner token",
          code: "UNAUTHORIZED"
        )
        return
      end

      @current_runner = Runner.find_by(token:)
      unless @current_runner
        render_json_error(
          status: :unauthorized,
          message: "Invalid runner token",
          code: "UNAUTHORIZED"
        )
      end
    end

    def extract_runner_token
      request.headers["X-Runner-Token"].to_s.strip.presence
    end

    def set_run
      @run = @current_runner.runs.find(params[:task_id])
    end

    # Prefer id when present so a local filename rename updates the same
    # Runner::Job (slug change) instead of inserting a duplicate under the new slug.
    def find_or_build_job(id:, slug:)
      if id.present? && (job = @current_runner.jobs.find_by(id: id))
        job
      else
        @current_runner.jobs.find_or_initialize_by(slug: slug)
      end
    end

    def job_attrs_from(data)
      slug = data[:slug].to_s.strip.downcase
      name = data[:name].presence || (slug.present? ? slug.tr("_-", " ").titleize : nil)
      cron = data.key?(:cron) ? data[:cron].presence : "*/5 * * * *"
      timezone = IanaTimezone.resolve(data[:timezone])
      timeout_seconds = data[:timeout_seconds].presence || 30
      config = data[:config].respond_to?(:to_unsafe_h) ? data[:config].to_unsafe_h : (data[:config] || {})
      if data[:description].present? && !config.key?("description")
        config = config.merge("description" => data[:description].to_s)
      end

      {
        account: @current_runner.account,
        slug: slug,
        name: name,
        cron: cron,
        timezone: timezone,
        timeout_seconds: timeout_seconds,
        config: config
      }
    end
end
