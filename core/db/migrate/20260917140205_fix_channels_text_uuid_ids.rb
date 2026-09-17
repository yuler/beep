class FixChannelsTextUuidIds < ActiveRecord::Migration[8.2]
  # The 20260911110000 backfill wrote channels.id as dashed-UUID TEXT via raw SQL,
  # bypassing the app's base36/BLOB(16) UUID codec
  # (core/lib/rails_ext/active_record_type_uuid.rb). SQLite compares BLOB != TEXT,
  # so Rails can list those rows but never find-by-id them: push_subscriptions and
  # channels destroy/test/show 404 with "Couldn't find ...". account_id/user_id were
  # copied verbatim as BLOBs, which is why user scoping kept working.
  #
  # This rewrites TEXT ids to the 16 raw UUID bytes the codec expects, and re-points
  # any dependent rows that Rails wrote with serialize(misdecoded_id) blobs.
  # Rows already stored as BLOB(16) are untouched, so this is a safe no-op there.
  def up
    text_ids = select_values("SELECT id FROM channels WHERE typeof(id) = 'text'")
    say_with_time "Rewriting #{text_ids.size} TEXT channel id(s) to BLOB(16)" do
      text_ids.each do |text_id|
        hex = text_id.to_s.delete("-")
        unless hex.match?(/\A[0-9a-fA-F]{32}\z/)
          say "  skipping non-UUID id: #{text_id.inspect}", true
          next
        end

        fixed_hex = hex.downcase
        execute("UPDATE channels SET id = x'#{fixed_hex}' WHERE id = #{quote(text_id)}")
        remap_dependents(text_id, fixed_hex)
      end
    end

    leftovers = select_value("SELECT count(*) FROM channel_deliveries WHERE length(channel_id) != 16")
    say "  channel_deliveries with non-16-byte channel_id left: #{leftovers}", true if leftovers.to_i > 0
  end

  def down
    # Irreversible data fix (same precedent as 20260911110000).
  end

  private
    # Dependent rows created via Rails for a TEXT-id channel stored
    # serialize(misdecoded_id) instead of the real id. Recompute that value
    # deterministically (mirrors UuidBase36 + Type::Uuid#serialize) so the rows
    # can be re-pointed. Returns [full_hex, truncated_hex]: the Binary type may
    # have truncated the blob to 16 bytes on write, so match both shapes.
    def garbage_variants(text_id)
      ascii_hex = text_id.unpack1("H*")
      misdecoded = ascii_hex.to_i(16).to_s(36).rjust(25, "0")
      garbage_hex = misdecoded.to_i(36).to_s(16).rjust(32, "0")
      garbage_hex = garbage_hex[0...-1] if garbage_hex.length.odd?
      [ garbage_hex, garbage_hex[0, 32] ]
    end

    def remap_dependents(text_id, fixed_hex)
      full, trunc = garbage_variants(text_id)
      {
        "channel_deliveries" => "channel_id",
        "channel_authorizations" => "channel_id"
      }.each do |table, column|
        count = select_value(<<~SQL)
          SELECT count(*) FROM #{table}
          WHERE #{column} = x'#{full}'
             OR (length(#{column}) = 16 AND #{column} = x'#{trunc}')
        SQL
        next if count.to_i.zero?

        execute(<<~SQL)
          UPDATE #{table} SET #{column} = x'#{fixed_hex}'
          WHERE #{column} = x'#{full}'
             OR (length(#{column}) = 16 AND #{column} = x'#{trunc}')
        SQL
        say "  remapped #{count} row(s) in #{table}", true
      end
    end
end
