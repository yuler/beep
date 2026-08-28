namespace :beeper_apps do
  desc "Sync official beeper apps from manifests into database (idempotent)"
  task sync: :environment do
    puts "Syncing official beeper apps..."
    stats = BeeperApp.seed_official
    puts "Done: #{stats[:created]} created, #{stats[:updated]} updated, #{stats[:unchanged]} unchanged."
  end
end
