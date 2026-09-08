json.extract! runner, :id, :name, :masked_token, :status, :tags, :version, :os, :arch, :hostname, :ip_address, :last_seen_at, :created_at, :updated_at
json.is_online runner.online?
json.jobs_count runner.has_attribute?(:jobs_count) ? runner[:jobs_count].to_i : runner.jobs.count
