class Api::V1::BeepsController < Api::V1::BaseController
  rescue_from Beep::ScheduleParser::Error do |exception|
    render_json_error(
      status: :unprocessable_entity,
      message: exception.message,
      code: "VALIDATION_ERROR",
      errors: [ exception.message ]
    )
  end

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
                                               ordered_by: index_order
    @run_stats = Beep::Run.stats_by_beep(@beeps.map(&:id))
    @recent_runs = Beep::Run.recent_by_beep(@beeps.map(&:id))
    render :index
  end

  def show
    @beep = Current.account.beeps.includes(beeper: :beeper_app).find(params[:id])
    render :show
  end

  def create
    # Resolved timezone respects user's preference first, falling back to payload timezone, then UTC.
    # Note: request payload timezone does not override an already configured user timezone.
    timezone = beep_timezone
    run_at = resolve_schedule_run_at(timezone: timezone)
    kind = params[:kind].presence || (params[:cron].present? ? "recurring" : "once")

    attrs = beep_params.merge(kind: kind, timezone: timezone)
    attrs[:run_at] = run_at if run_at.present?

    @beep = Current.account.beeps.new(attrs)

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

    # Timezone parameter serves as a transient reference zone to interpret :at wall-clock time if provided,
    # falling back to the beep's existing timezone. The beep's stored timezone itself remains immutable.
    timezone = IanaTimezone.resolve(params[:timezone], @beep.timezone)
    run_at = resolve_schedule_run_at(timezone: timezone)

    attrs = beep_params
    attrs = attrs.merge(run_at: run_at) if run_at.present?

    if @beep.update(attrs)
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
    SORT_COLUMNS = {
      "title" => :title,
      "status" => :status,
      "created_at" => :created_at,
      "next_run_at" => :next_run_at,
      "schedule" => :next_run_at,
      "scheduled_at" => :next_run_at
    }.freeze

    def index_order
      column = SORT_COLUMNS[params[:sort].to_s]
      if column
        direction = params[:dir].to_s == "asc" ? :asc : :desc
        { column => direction, id: direction }
      else
        { created_at: :desc, id: :desc }
      end
    end

    def beep_params
      params.permit(:title, :body, :run_at, :cron, :kind, :intent, notification_channels: [], metadata: {})
    end

    def resolve_schedule_run_at(timezone:)
      if params[:cron].present? && (params[:in].present? || params[:at].present?)
        raise Beep::ScheduleParser::Error, "Cannot specify :cron together with :in or :at"
      end

      Beep::ScheduleParser.resolve_run_at(
        in_val: params[:in],
        at_val: params[:at],
        run_at_val: params[:run_at],
        timezone: timezone
      )
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
