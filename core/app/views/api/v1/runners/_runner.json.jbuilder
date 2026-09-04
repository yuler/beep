json.extract! runner, :id, :name, :token_prefix, :status, :tags, :version, :os, :arch, :hostname, :ip_address, :last_seen_at, :created_at, :updated_at
json.is_online runner.online?
json.jobs_count runner.jobs.count
