class Beep::Plaintext
  def self.from_markdown(source)
    if source.blank?
      return ""
    end

    text = source.to_s
    text = text.gsub(/\[([^\]]*)\]\(([^)]*)\)/, '\1')
    text = text.gsub(/(\*\*|__)(.+?)\1/m, '\2')
    text = text.gsub(/(\*|_)(.+?)\1/m, '\2')
    text = text.gsub(/^[ \t]*[-*+][ \t]+/, "")
    text = text.gsub(/^[ \t]*\d+\.[ \t]+/, "")
    text = text.gsub(/[ \t]+\n/, "\n")
    text = text.gsub(/[ \t]{2,}/, " ")
    text.strip
  end
end
