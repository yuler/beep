require "test_helper"

class Api::V1::Dev::LettersControllerTest < ActionDispatch::IntegrationTest
  setup do
    skip "letter_opener_web not loaded" unless defined?(LetterOpenerWeb::Letter)

    @identity = identities(:john)
    @session = @identity.sessions.create!
    @token = @session.signed_id
    @auth = { "Authorization" => "Bearer #{@token}" }

    @original_location = LetterOpenerWeb.config.letters_location
    @letters_root = Rails.root.join("tmp/letter_opener_test_#{Process.pid}_#{SecureRandom.hex(4)}")
    LetterOpenerWeb.configure { |config| config.letters_location = @letters_root }
    LetterOpenerWeb::Letter.letters_location = @letters_root
    FileUtils.mkdir_p(@letters_root)
  end

  teardown do
    FileUtils.rm_rf(@letters_root) if @letters_root
    if @original_location
      LetterOpenerWeb.configure { |config| config.letters_location = @original_location }
      LetterOpenerWeb::Letter.letters_location = @original_location
    end
  end

  test "index returns recent letters when local" do
    create_letter!("test_letter_index")

    get api_v1_dev_letters_url, headers: @auth, as: :json

    assert_response :success
    body = response.parsed_body
    assert body["letters"].is_a?(Array)
    letter = body["letters"].find { |item| item["id"] == @letter_id }
    assert_not_nil letter
    assert_equal "Hello", letter["subject"]
    assert_equal "to@example.com", letter["to"]
    assert_equal "from@example.com", letter["from"]
    assert letter["sent_at"].present?
  end

  test "index requires authentication" do
    get api_v1_dev_letters_url, as: :json

    assert_response :unauthorized
  end

  test "destroy deletes a letter" do
    create_letter!("test_letter_destroy")

    delete api_v1_dev_letter_url(@letter_id), headers: @auth, as: :json

    assert_response :no_content
    assert_not LetterOpenerWeb::Letter.find(@letter_id).valid?
  end

  test "destroy returns not found for missing letter" do
    delete api_v1_dev_letter_url("missing_letter"), headers: @auth, as: :json

    assert_response :not_found
  end

  test "clear deletes all letters" do
    create_letter!("test_letter_clear_a")
    write_letter!("test_letter_clear_b")

    delete clear_api_v1_dev_letters_url, headers: @auth, as: :json

    assert_response :no_content
    assert_empty LetterOpenerWeb::Letter.search
  end

  private
    def create_letter!(id)
      @letter_id = id
      write_letter!(id)
    end

    def write_letter!(id)
      dir = @letters_root.join(id)
      FileUtils.mkdir_p(dir)
      File.write(dir.join("rich.html"), <<~HTML)
        <html><body>
          <div id="container">
            <div id="message_headers">
              <dl>
                <dt>From:</dt>
                <dd>from@example.com</dd>
                <dt>Subject:</dt>
                <dd><strong>Hello</strong></dd>
                <dt>To:</dt>
                <dd>to@example.com</dd>
              </dl>
            </div>
          </div>
        </body></html>
      HTML
    end
end
