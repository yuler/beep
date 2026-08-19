# This file is auto-generated from the current state of the database. Instead
# of editing this file, please use the migrations feature of Active Record to
# incrementally modify your database, and then regenerate this schema definition.
#
# This file is the source Rails uses to define your schema when running `bin/rails
# db:schema:load`. When creating a new database, `bin/rails db:schema:load` tends to
# be faster and is potentially less error prone than running all of your
# migrations from scratch. Old migrations may fail to apply correctly if those
# migrations use external dependencies or application code.
#
# It's strongly recommended that you check this file into your version control system.

ActiveRecord::Schema[8.2].define(version: 2026_08_18_160000) do
  create_table "account_charges", id: :uuid, force: :cascade do |t|
    t.uuid "account_id", null: false
    t.integer "amount", null: false
    t.integer "amount_refunded", default: 0
    t.string "checkout_id", null: false
    t.datetime "created_at", null: false
    t.string "currency", default: "USD"
    t.string "plan_key", null: false
    t.string "provider", null: false
    t.json "raw", null: false
    t.string "status", null: false
    t.uuid "subscription_id"
    t.datetime "updated_at", null: false
    t.index ["account_id"], name: "index_account_charges_on_account_id"
    t.index ["provider", "checkout_id"], name: "index_account_charges_on_provider_and_checkout_id"
    t.index ["subscription_id"], name: "index_account_charges_on_subscription_id"
  end

  create_table "account_invitation_acceptances", id: :uuid, force: :cascade do |t|
    t.uuid "invitation_id", null: false
    t.uuid "identity_id", null: false
    t.uuid "user_id", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["identity_id"], name: "index_account_invitation_acceptances_on_identity_id"
    t.index ["invitation_id"], name: "index_account_invitation_acceptances_on_invitation_id", unique: true
    t.index ["user_id"], name: "index_account_invitation_acceptances_on_user_id"
  end

  create_table "account_invitation_declines", id: :uuid, force: :cascade do |t|
    t.uuid "invitation_id", null: false
    t.uuid "identity_id", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["identity_id"], name: "index_account_invitation_declines_on_identity_id"
    t.index ["invitation_id"], name: "index_account_invitation_declines_on_invitation_id", unique: true
  end

  create_table "account_invitations", id: :uuid, force: :cascade do |t|
    t.uuid "account_id", null: false
    t.datetime "created_at", null: false
    t.string "email", null: false
    t.uuid "invited_by_id"
    t.string "role"
    t.string "token", null: false
    t.datetime "updated_at", null: false
    t.index ["account_id", "email"], name: "index_account_invitations_on_account_id_and_email"
    t.index ["account_id"], name: "index_account_invitations_on_account_id"
    t.index ["invited_by_id"], name: "index_account_invitations_on_invited_by_id"
    t.index ["token"], name: "index_account_invitations_on_token", unique: true
  end

  create_table "account_join_codes", id: :uuid, force: :cascade do |t|
    t.uuid "account_id", null: false
    t.string "code", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.integer "usage_count", default: 0, null: false
    t.integer "usage_limit", default: 10, null: false
    t.index ["account_id"], name: "index_account_join_codes_on_account_id", unique: true
    t.index ["code"], name: "index_account_join_codes_on_code", unique: true
  end

  create_table "account_payment_webhooks", id: :uuid, force: :cascade do |t|
    t.uuid "account_id", null: false
    t.datetime "created_at", null: false
    t.string "event_id", null: false
    t.string "event_type", null: false
    t.string "provider", null: false
    t.json "raw", null: false
    t.datetime "updated_at", null: false
    t.index ["account_id"], name: "index_account_payment_webhooks_on_account_id"
    t.index ["provider", "event_type"], name: "index_account_payment_webhooks_on_provider_and_event_type"
  end

  create_table "account_slug_holds", id: :uuid, force: :cascade do |t|
    t.uuid "account_id", null: false
    t.string "slug", null: false
    t.datetime "expires_at", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["account_id"], name: "index_account_slug_holds_on_account_id"
    t.index ["expires_at"], name: "index_account_slug_holds_on_expires_at"
    t.index ["slug"], name: "index_account_slug_holds_on_slug"
  end

  create_table "account_subscriptions", id: :uuid, force: :cascade do |t|
    t.uuid "account_id", null: false
    t.datetime "canceled_at"
    t.datetime "created_at", null: false
    t.datetime "current_period_end"
    t.string "customer_id", null: false
    t.integer "next_amount", null: false
    t.string "plan_key", null: false
    t.string "provider", null: false
    t.json "raw", null: false
    t.string "status", null: false
    t.string "subscription_id", null: false
    t.datetime "updated_at", null: false
    t.index ["account_id"], name: "index_account_subscriptions_on_account_id"
    t.index ["provider", "subscription_id"], name: "index_account_subscriptions_on_provider_and_subscription_id"
  end

  create_table "accounts", id: :uuid, force: :cascade do |t|
    t.datetime "created_at", null: false
    t.string "description"
    t.string "name", null: false
    t.boolean "personal", default: false, null: false
    t.string "slug", null: false
    t.datetime "updated_at", null: false
    t.index ["personal"], name: "index_accounts_on_personal"
    t.index ["slug"], name: "index_accounts_on_slug", unique: true
  end

  create_table "action_text_rich_texts", id: :uuid, force: :cascade do |t|
    t.text "body"
    t.datetime "created_at", null: false
    t.string "name", null: false
    t.uuid "record_id", null: false
    t.string "record_type", null: false
    t.datetime "updated_at", null: false
    t.index ["record_type", "record_id", "name"], name: "index_action_text_rich_texts_uniqueness", unique: true
  end

  create_table "active_storage_attachments", id: :uuid, force: :cascade do |t|
    t.uuid "blob_id", null: false
    t.datetime "created_at", null: false
    t.string "name", null: false
    t.uuid "record_id", null: false
    t.string "record_type", null: false
    t.index ["blob_id"], name: "index_active_storage_attachments_on_blob_id"
    t.index ["record_type", "record_id", "name", "blob_id"], name: "index_active_storage_attachments_uniqueness", unique: true
  end

  create_table "active_storage_blobs", id: :uuid, force: :cascade do |t|
    t.bigint "byte_size", null: false
    t.string "checksum"
    t.string "content_type"
    t.datetime "created_at", null: false
    t.string "filename", null: false
    t.string "key", null: false
    t.text "metadata"
    t.string "service_name", null: false
    t.index ["key"], name: "index_active_storage_blobs_on_key", unique: true
  end

  create_table "active_storage_variant_records", id: :uuid, force: :cascade do |t|
    t.uuid "blob_id", null: false
    t.string "variation_digest", null: false
    t.index ["blob_id", "variation_digest"], name: "index_active_storage_variant_records_uniqueness", unique: true
  end

  create_table "beep_runs", id: :uuid, force: :cascade do |t|
    t.uuid "beep_id", null: false
    t.datetime "scheduled_for", null: false
    t.string "status", default: "pending", null: false
    t.json "result"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["beep_id", "scheduled_for"], name: "index_beep_runs_on_beep_id_and_scheduled_for", unique: true
  end

  create_table "beeps", id: :uuid, force: :cascade do |t|
    t.uuid "account_id", null: false
    t.string "kind", null: false
    t.datetime "run_at"
    t.datetime "next_run_at"
    t.datetime "last_run_at"
    t.string "timezone", default: "UTC", null: false
    t.string "cron"
    t.string "status", default: "active", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.text "title", null: false
    t.text "body"
    t.index ["account_id"], name: "index_beeps_on_account_id"
    t.index ["status", "next_run_at"], name: "index_beeps_on_status_and_next_run_at"
  end

  create_table "identities", id: :uuid, force: :cascade do |t|
    t.datetime "created_at", null: false
    t.string "email", null: false
    t.boolean "staff", default: false, null: false
    t.datetime "updated_at", null: false
    t.string "last_account_slug"
    t.index ["email"], name: "index_identities_on_email", unique: true
  end

  create_table "invite_codes", id: :uuid, force: :cascade do |t|
    t.string "code", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["code"], name: "index_invite_codes_on_code", unique: true
  end

  create_table "magic_links", id: :uuid, force: :cascade do |t|
    t.string "code"
    t.datetime "created_at", null: false
    t.datetime "expires_at"
    t.uuid "identity_id", null: false
    t.integer "purpose"
    t.datetime "updated_at", null: false
    t.index ["code"], name: "index_magic_links_on_code", unique: true
    t.index ["expires_at"], name: "index_magic_links_on_expires_at"
    t.index ["identity_id"], name: "index_magic_links_on_identity_id"
  end

  create_table "push_subscriptions", id: :uuid, force: :cascade do |t|
    t.uuid "account_id", null: false
    t.uuid "user_id", null: false
    t.text "endpoint"
    t.string "p256dh_key"
    t.string "auth_key"
    t.string "user_agent", limit: 4096
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["account_id"], name: "index_push_subscriptions_on_account_id"
    t.index ["user_id", "endpoint"], name: "index_push_subscriptions_on_user_id_and_endpoint", unique: true
  end

  create_table "sessions", id: :uuid, force: :cascade do |t|
    t.datetime "created_at", null: false
    t.uuid "identity_id", null: false
    t.string "ip_address"
    t.datetime "updated_at", null: false
    t.string "user_agent", limit: 4096
    t.index ["identity_id"], name: "index_sessions_on_identity_id"
  end

  create_table "users", id: :uuid, force: :cascade do |t|
    t.uuid "account_id", null: false
    t.boolean "active", default: true, null: false
    t.datetime "created_at", null: false
    t.uuid "identity_id"
    t.string "name", null: false
    t.json "notification_channels", default: ["email"], null: false
    t.string "role", default: "member", null: false
    t.datetime "updated_at", null: false
    t.datetime "verified_at"
    t.index ["account_id", "identity_id"], name: "index_users_on_account_id_and_identity_id", unique: true
    t.index ["account_id", "role"], name: "index_users_on_account_id_and_role"
    t.index ["identity_id"], name: "index_users_on_identity_id"
  end

  add_foreign_key "account_invitation_acceptances", "account_invitations", column: "invitation_id"
  add_foreign_key "account_invitation_acceptances", "identities"
  add_foreign_key "account_invitation_acceptances", "users"
  add_foreign_key "account_invitation_declines", "account_invitations", column: "invitation_id"
  add_foreign_key "account_invitation_declines", "identities"
  add_foreign_key "account_slug_holds", "accounts"
  add_foreign_key "beep_runs", "beeps"
  add_foreign_key "beeps", "accounts"
  add_foreign_key "push_subscriptions", "accounts"
  add_foreign_key "push_subscriptions", "users"
end
