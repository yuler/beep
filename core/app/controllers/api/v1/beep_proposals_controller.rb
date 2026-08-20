class Api::V1::BeepProposalsController < Api::V1::BaseController
  rate_limit to: 20, within: 15.minutes, only: :create, with: :rate_limit_exceeded

  def create
    prompt = params[:prompt].to_s.strip

    if prompt.blank?
      render_json_error(
        status: :unprocessable_entity,
        message: "Enter a reminder in plain language.",
        code: "VALIDATION_ERROR"
      )
    elsif ENV["DEEPSEEK_API_KEY"].blank?
      render_unavailable
    else
      propose(prompt)
    end
  end

  private
    def propose(prompt)
      @proposal = Beep::Proposal.create(prompt)
      render :create, status: :created
    rescue Beep::Proposal::Error, JSON::ParserError
      render_json_error(
        status: :bad_gateway,
        message: "Could not understand that reminder. Try again or fill in the form.",
        code: "LLM_ERROR"
      )
    rescue RubyLLM::UnauthorizedError, RubyLLM::ConfigurationError, RubyLLM::PaymentRequiredError
      render_unavailable
    rescue RubyLLM::RateLimitError, RubyLLM::OverloadedError
      render_json_too_many_requests
    rescue RubyLLM::Error
      render_json_error(
        status: :bad_gateway,
        message: "Could not reach the language model. Try again or fill in the form.",
        code: "LLM_ERROR"
      )
    end

    def render_unavailable
      render_json_error(
        status: :service_unavailable,
        message: "Natural language create is not configured.",
        code: "LLM_UNAVAILABLE"
      )
    end

    def rate_limit_exceeded
      render_json_too_many_requests
    end
end
