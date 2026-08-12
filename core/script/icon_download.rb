require_relative "../config/environment"
require "net/http"
require "uri"

def helpers
  <<~HELP
    Download SVG icons from Lucide (https://lucide.dev)

    USAGE:
      ruby script/icon_download.rb <icon_name>
      ruby script/icon_download.rb <icon_name1>,<icon_name2>,<icon_name3>

    EXAMPLES:
      ruby script/icon_download.rb email
      ruby script/icon_download.rb lock
      ruby script/icon_download.rb user-circle
      ruby script/icon_download.rb email,lock,user-circle

    Icons will be saved to: app/assets/images/<icon_name>.svg
    Browse available icons at: https://lucide.dev/icons
  HELP
end

def icon_url(icon_name)
  "https://raw.githubusercontent.com/lucide-icons/lucide/refs/heads/main/icons/#{icon_name}.svg"
end

def icon_remote_available?(icon_name)
  uri = URI(icon_url(icon_name))
  response = Net::HTTP.get_response(uri)
  response.code == "200"
rescue StandardError => e
  puts "Error checking icon: #{e.message}"
  false
end

def icon_local_exists?(icon_name)
  Rails.root.join("app", "assets", "images", "#{icon_name}.svg").exist?
rescue StandardError => e
  puts "Error checking icon: #{e.message}"
  false
end

def download_icon(icon_name)
  uri = URI(icon_url(icon_name))
  destination = Rails.root.join("app", "assets", "images", "#{icon_name}.svg")

  puts "Downloading #{icon_name} from Lucide..."

  svg_content = Net::HTTP.get(uri)
  File.write(destination, svg_content)

  puts "✓ Successfully saved to #{destination}"
  true
rescue StandardError => e
  puts "✗ Error downloading icon: #{e.message}"
  false
end


# Main
icon_input = ARGV[0]
if icon_input.nil?
  puts helpers
  puts "\n" + "─" * 50
  print "Enter icon name(s) to download (separate with comma): "
  icon_input = gets.chomp
end

# Parse multiple icons separated by comma
icon_names = icon_input.split(",").map(&:strip).reject(&:empty?)

if icon_names.empty?
  puts "✗ No icon names provided"
  exit 1
end

# Track results
results = { downloaded: [], skipped: [], failed: [] }

# Process each icon
icon_names.each do |icon_name|
  puts "\n" + "─" * 50 if icon_names.length > 1
  puts "Processing: #{icon_name}" if icon_names.length > 1

  if icon_local_exists?(icon_name)
    puts "✓ Icon '#{icon_name}' already exists locally"
    results[:skipped] << icon_name
  elsif icon_remote_available?(icon_name)
    if download_icon(icon_name)
      results[:downloaded] << icon_name
    else
      results[:failed] << icon_name
    end
  else
    puts "✗ Icon '#{icon_name}' does not exist on Lucide"
    puts "  Browse available icons at: https://lucide.dev/icons"
    results[:failed] << icon_name
  end
end

# Summary
if icon_names.length > 1
  puts "\n" + "=" * 50
  puts "SUMMARY:"
  puts "  Downloaded: #{results[:downloaded].length} (#{results[:downloaded].join(', ')})" unless results[:downloaded].empty?
  puts "  Skipped: #{results[:skipped].length} (#{results[:skipped].join(', ')})" unless results[:skipped].empty?
  puts "  Failed: #{results[:failed].length} (#{results[:failed].join(', ')})" unless results[:failed].empty?
end

# Exit with appropriate code
exit(results[:failed].empty? ? 0 : 1)
