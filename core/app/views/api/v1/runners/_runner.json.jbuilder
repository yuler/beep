json.extract! runner, :id, :name, :token_prefix, :status, :tags, :allow_exec, :version, :os, :arch, :hostname, :ip_address, :last_seen_at, :created_at, :updated_at
json.is_online runner.status.in?(%w[ online idle ]) && runner.last_seen_at.present? && runner.last_seen_at >= Runner::OFFLINE_TIMEOUT.ago
