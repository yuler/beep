require_relative "../config/environment"

# Drop database, run migrate, to fresh the database schema for sqlite
puts "Dumping database schema for sqlite"
system("bin/rails db:drop")
system("rm db/schema.rb") if File.exist?("db/schema.rb")
system("bin/rails db:migrate")
system("bin/rails db:setup")

puts "All done"
