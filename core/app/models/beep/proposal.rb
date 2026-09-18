class Beep::Proposal
  MODEL = "deepseek-chat"
  ACTIONS = %w[ create other ].freeze
  INTENTS = ACTIONS

  class Error < StandardError; end

  class Result
    attr_reader :action, :intent, :metadata, :kind, :title, :body, :run_at, :cron, :timezone, :notification_channels, :errors, :message

    def initialize(action: nil, intent: nil, metadata: nil, kind: "once", title:, body:, run_at:, cron: nil, timezone:, notification_channels: nil, errors:, message:)
      if action.nil? && intent.in?(ACTIONS)
        @action = intent
        @intent = nil
      else
        @action = action || "create"
        @intent = intent
      end
      @metadata = metadata
      @kind = kind
      @title = title
      @body = body
      @run_at = run_at
      @cron = cron
      @timezone = timezone
      @notification_channels = notification_channels
      @errors = errors
      @message = message
    end

    def confirmable?
      action == "create" && title.present? && errors.blank?
    end
  end

  def self.create(prompt, timezone: IanaTimezone::DEFAULT, chat: nil)
    new(prompt, timezone: timezone, chat: chat).create
  end

  def initialize(prompt, timezone: IanaTimezone::DEFAULT, chat: nil)
    @prompt = prompt.to_s
    @timezone = IanaTimezone.resolve(timezone)
    @chat = chat
  end

  def create
    payload = parse_model_json(ask_model)
    build_result(payload)
  end

  private
    def ask_model
      chat = @chat || default_chat
      chat.with_instructions(instructions)
        .with_temperature(0)
        .with_params(response_format: { type: "json_object" })
        .ask(@prompt)
        .content
    end

    def default_chat
      RubyLLM.chat(model: MODEL, provider: :deepseek, assume_model_exists: true)
    end

    def parse_model_json(content)
      json = content.to_s.strip
      if json.start_with?("```")
        json = json.sub(/\A```(?:json)?\s*/i, "").sub(/```\s*\z/, "").strip
      end

      parsed = JSON.parse(json)
      if parsed.is_a?(Hash)
        parsed
      else
        raise Error, "Model returned invalid JSON"
      end
    rescue JSON::ParserError
      raise Error, "Model returned invalid JSON"
    end

    def build_result(payload)
      action = if payload["action"].to_s.in?(ACTIONS)
        payload["action"].to_s
      elsif payload["intent"].to_s.in?(ACTIONS)
        payload["intent"].to_s
      else
        inferred_action(payload)
      end

      raw_intent = payload["intent"].to_s.strip.presence
      business_intent = if raw_intent.present? && !raw_intent.in?(ACTIONS)
        normalized = raw_intent.underscore.tr(" -", "_").gsub(/[^a-zA-Z0-9_]/, "").downcase
        truncate(normalized.presence || raw_intent, 100)
      end

      metadata = parse_metadata(payload["metadata"])

      raw_kind = payload["kind"].to_s.downcase
      raw_cron = payload["cron"].to_s.strip.presence
      kind = (raw_kind == "recurring" || raw_cron.present?) ? "recurring" : "once"

      title = truncate(payload["title"], Beep::TITLE_MAX_LENGTH)
      body = truncate(payload["body"], Beep::BODY_MAX_LENGTH)
      errors = {}

      run_at = nil
      cron = nil

      if kind == "recurring"
        cron = raw_cron
        if cron.blank?
          errors["cron"] = "can't be blank"
        else
          parsed_cron = Fugit.parse(cron) rescue nil
          unless parsed_cron.is_a?(Fugit::Cron)
            errors["cron"] = "is not a valid cron expression"
          end
        end
      else
        run_at, run_at_error = parse_run_at(payload["run_at"])
        if run_at_error.present?
          errors["run_at"] = run_at_error
        end
      end

      if action == "create" && title.blank?
        errors["title"] = "can't be blank"
      end

      message = if action == "other"
        "Describe what to be reminded of."
      end

      raw_channels = payload["notification_channels"]
      channels = if raw_channels.is_a?(Array)
        allowed = raw_channels.map(&:to_s).select { |c| User::NOTIFICATION_CHANNELS.include?(c) }.uniq
        allowed.presence
      end

      Result.new(
        action: action,
        intent: business_intent,
        metadata: metadata,
        kind: kind,
        title: title,
        body: body,
        run_at: run_at,
        cron: cron,
        timezone: @timezone,
        notification_channels: channels,
        errors: errors,
        message: message
      )
    end

    def inferred_action(payload)
      if payload["title"].present? || payload["run_at"].present? || payload["cron"].present?
        "create"
      else
        "other"
      end
    end

    def parse_metadata(value)
      case value
      when Hash
        value.presence
      when String
        parsed = JSON.parse(value) rescue nil
        parsed.is_a?(Hash) ? parsed.presence : nil
      else
        nil
      end
    end

    def parse_run_at(value)
      if value.blank?
        [ nil, nil ]
      else
        time = Time.use_zone(@timezone) { Time.zone.parse(value.to_s) } || Time.iso8601(value.to_s)
        if time.future?
          [ time, nil ]
        else
          [ time, "must be in the future" ]
        end
      end
    rescue ArgumentError
      [ nil, "is invalid" ]
    end

    def truncate(value, limit)
      text = value.to_s.strip.presence
      if text && text.length > limit
        text[0, limit]
      else
        text
      end
    end

    def instructions
      now = Time.current.in_time_zone(@timezone)

      <<~PROMPT
        You extract a reminder from the user message.
        Timezone is #{@timezone}. Current datetime is #{now.iso8601}.
        Reply with JSON only:
        {"action":"create"|"other","intent":string|null,"metadata":object|null,"kind":"once"|"recurring","title":string|null,"body":string|null,"run_at":string|null,"cron":string|null,"notification_channels":string[]|null}
        action is "create" when the user wants a reminder or alert, otherwise "other".
        intent: short lowercase English snake_case slug representing the business purpose or intent if mentioned (e.g. '目的是 用来测试' -> 'test', '目的是 午休' -> 'lunch_break', '下班' -> 'get_off_work', '测试' -> 'test'). Must always be translated/normalized into English snake_case for programmatic handling in code. If not mentioned, set to null.
        metadata: JSON object if structured data, key-value attributes, or JSON payload is mentioned (e.g. 'json 数据"{"a": "b"}""' -> {"a": "b"}), otherwise null. Must be a JSON object, not a string or array.
        kind: "recurring" when the reminder repeats on a schedule or interval, otherwise "once".
        title: concise title, max #{Beep::TITLE_MAX_LENGTH} characters. Do not include intent, metadata, or schedule commands in the title.
        body: optional extra detail as markdown, max #{Beep::BODY_MAX_LENGTH} characters.
        run_at: future datetime as UTC ISO8601 if kind is "once" and a specific time is mentioned, otherwise null. Convert relative times using the timezone.
        cron: 5-part standard cron string ("min hour day month weekday") if kind is "recurring" (e.g., "*/5 14-20 * * *" for every 5 min from 14:00 to 20:59, "0 9 * * *" for daily 9am, "0 9 * * 1-5" for weekdays 9am), otherwise null.
        notification_channels: array of channel names or null if not mentioned. Allowed channels: email, web_push, cli.
        Channel mapping:
        - "email": email / e-mail / 邮件 / 邮箱
        - "web_push": web push / webpush / 浏览器推送 / 网页推送 / push
        - "cli": cli / 终端 / 命令行
        If specific channels are requested (e.g., "只/仅/only ..."), return only the requested channels.
        If user specifies not to use certain channels (e.g., "不要/别/no/without ..."), exclude them.
        If no channel is mentioned, set notification_channels to null. Only use allowed channels ("email", "web_push", "cli"); ignore unknown channels.
        Do not invent a reminder when the message is not a create request.
        Only one reminder.
      PROMPT
    end
end
