class Beep::Proposal
  MODEL = "deepseek-chat"
  INTENTS = %w[ create other ].freeze

  class Error < StandardError; end

  class Result
    attr_reader :intent, :title, :body, :run_at, :timezone, :errors, :message

    def initialize(intent:, title:, body:, run_at:, timezone:, errors:, message:)
      @intent = intent
      @title = title
      @body = body
      @run_at = run_at
      @timezone = timezone
      @errors = errors
      @message = message
    end

    def confirmable?
      intent == "create" && title.present? && errors.blank?
    end
  end

  def self.create(prompt, chat: nil)
    new(prompt, chat: chat).create
  end

  def initialize(prompt, chat: nil)
    @prompt = prompt.to_s
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
      intent = INTENTS.include?(payload["intent"].to_s) ? payload["intent"].to_s : inferred_intent(payload)
      title = truncate(payload["title"], Beep::TITLE_MAX_LENGTH)
      body = truncate(payload["body"], Beep::BODY_MAX_LENGTH)
      run_at, run_at_error = parse_run_at(payload["run_at"])
      errors = {}

      if intent == "create"
        if title.blank?
          errors["title"] = "can't be blank"
        end
        if run_at_error.present?
          errors["run_at"] = run_at_error
        end
      end

      message = if intent == "other"
        "Describe what to be reminded of."
      end

      Result.new(
        intent: intent,
        title: title,
        body: body,
        run_at: run_at,
        timezone: Beep::TIMEZONE,
        errors: errors,
        message: message
      )
    end

    def inferred_intent(payload)
      if payload["title"].present? || payload["run_at"].present?
        "create"
      else
        "other"
      end
    end

    def parse_run_at(value)
      if value.blank?
        [ nil, nil ]
      else
        time = Time.zone.parse(value.to_s) || Time.iso8601(value.to_s)
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
      now = Time.current.in_time_zone(Beep::TIMEZONE)

      <<~PROMPT
        You extract a reminder from the user message.
        Timezone is #{Beep::TIMEZONE}. Current datetime is #{now.iso8601}.
        Reply with JSON only:
        {"intent":"create"|"other","title":string|null,"body":string|null,"run_at":string|null}
        intent is "create" when the user wants a reminder or alert, otherwise "other".
        title: short title, max #{Beep::TITLE_MAX_LENGTH} characters.
        body: optional extra detail as markdown, max #{Beep::BODY_MAX_LENGTH} characters.
        run_at: future datetime as UTC ISO8601 if a specific one-time reminder time is mentioned, otherwise null. Convert relative times using the timezone.
        Do not invent a reminder when the message is not a create request.
        Only one reminder. If the user mentions recurring schedules or no specific time, set run_at to null.
      PROMPT
    end
end
