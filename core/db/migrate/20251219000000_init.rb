class Init < ActiveRecord::Migration[8.2]
  def change
    create_table "account_invitations", id: :uuid do |t|
      t.uuid "account_id", null: false
      t.datetime "created_at", null: false
      t.string "email", null: false
      t.uuid "invited_by_id"
      t.string "role"
      t.string "token", null: false
      t.datetime "updated_at", null: false
      t.index [ "account_id", "email" ], name: "index_account_invitations_on_account_id_and_email", unique: true
      t.index [ "account_id" ], name: "index_account_invitations_on_account_id"
      t.index [ "invited_by_id" ], name: "index_account_invitations_on_invited_by_id"
      t.index [ "token" ], name: "index_account_invitations_on_token", unique: true
    end

    create_table "account_join_codes", id: :uuid do |t|
      t.uuid "account_id", null: false
      t.string "code", null: false
      t.datetime "created_at", null: false
      t.datetime "updated_at", null: false
      t.integer "usage_count", default: 0, null: false
      t.integer "usage_limit", default: 10, null: false
      t.index [ "account_id" ], name: "index_account_join_codes_on_account_id", unique: true
      t.index [ "code" ], name: "index_account_join_codes_on_code", unique: true
    end

    create_table "accounts", id: :uuid do |t|
      t.datetime "created_at", null: false
      t.string "description"
      t.string "name", null: false
      t.boolean "personal", default: false, null: false
      t.string "slug", null: false
      t.datetime "updated_at", null: false
      t.index [ "personal" ], name: "index_accounts_on_personal"
      t.index [ "slug" ], name: "index_accounts_on_slug", unique: true
    end

    create_table "action_text_rich_texts", id: :uuid do |t|
      t.text "body"
      t.datetime "created_at", null: false
      t.string "name", null: false
      t.uuid "record_id", null: false
      t.string "record_type", null: false
      t.datetime "updated_at", null: false
      t.index [ "record_type", "record_id", "name" ], name: "index_action_text_rich_texts_uniqueness", unique: true
    end

    create_table "active_storage_attachments", id: :uuid do |t|
      t.uuid "blob_id", null: false
      t.datetime "created_at", null: false
      t.string "name", null: false
      t.uuid "record_id", null: false
      t.string "record_type", null: false
      t.index [ "blob_id" ], name: "index_active_storage_attachments_on_blob_id"
      t.index [ "record_type", "record_id", "name", "blob_id" ], name: "index_active_storage_attachments_uniqueness", unique: true
    end

    create_table "active_storage_blobs", id: :uuid do |t|
      t.bigint "byte_size", null: false
      t.string "checksum"
      t.string "content_type"
      t.datetime "created_at", null: false
      t.string "filename", null: false
      t.string "key", null: false
      t.text "metadata"
      t.string "service_name", null: false
      t.index [ "key" ], name: "index_active_storage_blobs_on_key", unique: true
    end

    create_table "active_storage_variant_records", id: :uuid do |t|
      t.uuid "blob_id", null: false
      t.string "variation_digest", null: false
      t.index [ "blob_id", "variation_digest" ], name: "index_active_storage_variant_records_uniqueness", unique: true
    end

    create_table "identities", id: :uuid do |t|
      t.datetime "created_at", null: false
      t.string "email", null: false
      t.boolean "staff", default: false, null: false
      t.datetime "updated_at", null: false
      t.index [ "email" ], name: "index_identities_on_email", unique: true
    end

    create_table "invite_codes", id: :uuid do |t|
      t.string "code", null: false
      t.datetime "created_at", null: false
      t.datetime "updated_at", null: false
      t.index [ "code" ], name: "index_invite_codes_on_code", unique: true
    end

    create_table "magic_links", id: :uuid do |t|
      t.string "code"
      t.datetime "created_at", null: false
      t.datetime "expires_at"
      t.uuid "identity_id", null: false
      t.integer "purpose"
      t.datetime "updated_at", null: false
      t.index [ "code" ], name: "index_magic_links_on_code", unique: true
      t.index [ "expires_at" ], name: "index_magic_links_on_expires_at"
      t.index [ "identity_id" ], name: "index_magic_links_on_identity_id"
    end

    create_table "sessions", id: :uuid do |t|
      t.datetime "created_at", null: false
      t.uuid "identity_id", null: false
      t.string "ip_address"
      t.datetime "updated_at", null: false
      t.string "user_agent", limit: 4096
      t.index [ "identity_id" ], name: "index_sessions_on_identity_id"
    end

    create_table "users", id: :uuid do |t|
      t.uuid "account_id", null: false
      t.boolean "active", default: true, null: false
      t.datetime "created_at", null: false
      t.uuid "identity_id"
      t.string "name", null: false
      t.string "role", default: "member", null: false
      t.datetime "updated_at", null: false
      t.datetime "verified_at"
      t.index [ "account_id", "identity_id" ], name: "index_users_on_account_id_and_identity_id", unique: true
      t.index [ "account_id", "role" ], name: "index_users_on_account_id_and_role"
      t.index [ "identity_id" ], name: "index_users_on_identity_id"
    end
  end
end
