#!/usr/bin/env ruby
require File.expand_path("../config/environment", File.dirname(__FILE__))

vapid_key = WebPush.generate_key

puts "VAPID_PRIVATE_KEY=#{vapid_key.private_key}"
puts "VAPID_PUBLIC_KEY=#{vapid_key.public_key}"
