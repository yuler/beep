class Api::V1::Dev::LettersController < Api::V1::BaseController
  disallow_account_scope
  before_action :ensure_local_letters!
  before_action :set_letter, only: :destroy

  def index
    letters = LetterOpenerWeb::Letter.search.first(50).filter_map do |letter|
      next unless letter.valid?

      serialize_letter(letter)
    end

    render_json json: { letters: letters }
  end

  def destroy
    @letter.delete
    head :no_content
  end

  def clear
    LetterOpenerWeb::Letter.destroy_all
    head :no_content
  end

  private
    def ensure_local_letters!
      unless Rails.env.local? && defined?(LetterOpenerWeb::Letter)
        head :not_found
      end
    end

    def set_letter
      @letter = LetterOpenerWeb::Letter.find(params[:id])
      head :not_found unless @letter.valid?
    end

    def serialize_letter(letter)
      {
        id: letter.id,
        sent_at: letter.sent_at&.iso8601,
        subject: header_value(letter, "Subject"),
        to: header_value(letter, "To"),
        from: header_value(letter, "From")
      }
    end

    def header_value(letter, name)
      html = letter.headers.to_s
      match = html.match(%r{<dt>\s*#{Regexp.escape(name)}:?\s*</dt>\s*<dd[^>]*>(.*?)</dd>}im)
      return nil unless match

      CGI.unescapeHTML(match[1].gsub(/<[^>]+>/, "")).strip.presence
    end
end
