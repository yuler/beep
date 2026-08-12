module IconsHelper
  # Inline an SVG from app/assets/images so it can inherit `currentColor`.
  def icon_tag(name, classes: nil, **html_options)
    path = Rails.root.join("app/assets/images", "#{name}.svg")
    doc = Nokogiri::XML(path.read)
    node = doc.root
    raise ArgumentError, "SVG root not found in #{path}" unless node

    node["class"] = classes if classes
    node["aria-hidden"] = "true" unless html_options.key?(:aria_hidden)
    node.remove_attribute("width")
    node.remove_attribute("height")

    html_options.except(:aria_hidden).each { |key, value| node[key.to_s.dasherize] = value }

    node.to_xml(save_with: Nokogiri::XML::Node::SaveOptions::AS_XML).html_safe
  end
end
