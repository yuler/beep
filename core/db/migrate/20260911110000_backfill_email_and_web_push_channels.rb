class BackfillEmailAndWebPushChannels < ActiveRecord::Migration[8.2]
  def up
    # Backfill Email channels for users
    execute <<-SQL
      INSERT INTO channels (id, account_id, user_id, kind, name, token, status, config, created_at, updated_at)
      SELECT
        lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))), 2) || '-' || substr('89ab', 1 + (abs(random()) % 4), 1) || substr(lower(hex(randomblob(2))), 2) || '-' || lower(hex(randomblob(6))),
        users.account_id,
        users.id,
        'email',
        coalesce(identities.email, users.name),
        'beep_ct_' || lower(hex(randomblob(16))),
        'active',
        '{}',
        datetime('now'),
        datetime('now')
      FROM users
      LEFT JOIN identities ON identities.id = users.identity_id
      WHERE NOT EXISTS (
        SELECT 1 FROM channels WHERE channels.user_id = users.id AND channels.kind = 'email'
      )
    SQL

    # Backfill Push subscriptions into channels table
    if table_exists?(:push_subscriptions)
      execute <<-SQL
        INSERT INTO channels (id, account_id, user_id, kind, name, token, status, config, created_at, updated_at)
        SELECT
          lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))), 2) || '-' || substr('89ab', 1 + (abs(random()) % 4), 1) || substr(lower(hex(randomblob(2))), 2) || '-' || lower(hex(randomblob(6))),
          ps.account_id,
          ps.user_id,
          'web_push',
          coalesce(ps.user_agent, 'Browser'),
          'beep_ct_' || lower(hex(randomblob(16))),
          'active',
          json_object('endpoint', ps.endpoint, 'p256dh_key', ps.p256dh_key, 'auth_key', ps.auth_key, 'user_agent', ps.user_agent),
          ps.created_at,
          ps.updated_at
        FROM push_subscriptions ps
      SQL
    end
  end

  def down
    execute "DELETE FROM channels WHERE kind IN ('email', 'web_push')"
  end
end
