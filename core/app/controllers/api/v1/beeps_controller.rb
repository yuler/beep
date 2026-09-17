class Api::V1::BeepsController < Api::V1::BaseController
  def index
    scope = Current.account.beeps.includes(beeper: :beeper_app)
    scope = scope.where(status: params[:status]) if params[:status].present? && Beep.statuses.key?(params[:status])
    scope = scope.where(kind: params[:kind]) if params[:kind].present? && Beep.kinds.key?(params[:kind])

    search_term = params[:q].presence || params[:search].presence
    if search_term.present?
      q_clean = search_term.strip
      q_term = "%#{ActiveRecord::Base.sanitize_sql_like(q_clean.downcase)}%"
      search_scope = scope.left_outer_joins(:beeper)
      text_condition = search_scope.where(
        "LOWER(beeps.title) LIKE :q OR LOWER(beeps.body) LIKE :q OR LOWER(beepers.title) LIKE :q",
        q: q_term
      )
      scope = if valid_uuid_format?(q_clean)
        search_scope.where(id: q_clean).or(text_condition)
      else
        text_condition
      end
    end

    @beeps = set_page_and_extract_portion_from scope,
                                               ordered_by: { created_at: :desc, id: :desc }
    @run_stats = Beep::Run.stats_by_beep(@beeps.map(&:id))
    @recent_runs = Beep::Run.recent_by_beep(@beeps.map(&:id))
    render :index
  end

  def show
    @beep = Current.account.beeps.includes(beeper: :beeper_app).find(params[:id])
    render :show
  end

  def create
    kind = params[:kind].presence || (params[:cron].present? ? "recurring" : "once")
    @beep = Current.account.beeps.new(beep_params.merge(kind: kind, timezone: beep_timezone))

    if @beep.save
      render :create, status: :created
    else
      render_json_error(
        status: :unprocessable_entity,
        message: @beep.errors.full_messages.to_sentence,
        code: "VALIDATION_ERROR",
        errors: @beep.errors.full_messages
      )
    end
  end

  def update
    @beep = Current.account.beeps.find(params[:id])

    if @beep.update(beep_params)
      render :show
    else
      render_json_error(
        status: :unprocessable_entity,
        message: @beep.errors.full_messages.to_sentence,
        code: "VALIDATION_ERROR",
        errors: @beep.errors.full_messages
      )
    end
  end

  def destroy
    Current.account.beeps.find(params[:id]).destroy!
    head :no_content
  end

  private
    def beep_params
      params.permit(:title, :body, :run_at, :cron, :kind, :intent, notification_channels: [], metadata: {})
    end

    def beep_timezone
      IanaTimezone.resolve(Current.user.timezone, params[:timezone])
    end

    UUID_REGEX = /\A[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\z/i
    BASE36_UUID_REGEX = /\A[0-9a-z]{25}\z/i

    def valid_uuid_format?(str)
      return true if str.match?(UUID_REGEX)
      return true if ActiveRecord::Base.connection.adapter_name.downcase.start_with?("sqlite") && str.match?(BASE36_UUID_REGEX)

      false
    end
end
