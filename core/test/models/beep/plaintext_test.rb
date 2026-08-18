require "test_helper"

class Beep::PlaintextTest < ActiveSupport::TestCase
  test "returns empty string for blank markdown" do
    assert_equal "", Beep::Plaintext.from_markdown(nil)
    assert_equal "", Beep::Plaintext.from_markdown("  ")
  end

  test "strips bold italic links and list markers" do
    source = <<~MARKDOWN
      Bring **milk** and _eggs_
      - [store](https://example.com)
      1. **done**
    MARKDOWN

    assert_equal "Bring milk and eggs\nstore\ndone", Beep::Plaintext.from_markdown(source)
  end
end
