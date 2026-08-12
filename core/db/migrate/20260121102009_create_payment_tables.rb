# This migration comes from pay (originally 1)
class CreatePaymentTables < ActiveRecord::Migration[8.2]
  def change
    create_table :account_charges, id: :uuid do |t|
      t.references :account, null: false, type: :uuid
      t.references :subscription, null: true, type: :uuid
      t.string :provider, null: false
      t.string :plan_key, null: false

      t.json :raw, null: false
      t.string :checkout_id, null: false
      t.string :currency, default: "USD"
      t.integer :amount, null: false
      t.integer :amount_refunded, default: 0
      t.string :status, null: false # pending, succeeded, failed, refunded

      t.timestamps

      t.index [ :provider, :checkout_id ]
    end

    create_table :account_subscriptions, id: :uuid do |t|
      t.references :account, null: false, type: :uuid
      t.string :provider, null: false
      t.string :plan_key, null: false

      t.json :raw, null: false
      t.string :customer_id, null: false
      t.string :subscription_id, null: false
      t.string :status, null: false # active past_due unpaid canceled incomplete incomplete_expired trialing paused
      t.integer :next_amount, null: false
      t.datetime :current_period_end
      t.datetime :canceled_at

      t.timestamps

      t.index [ :provider, :subscription_id ]
    end

    create_table :account_payment_webhooks, id: :uuid do |t|
      t.references :account, null: false, type: :uuid
      t.string :provider, null: false
      t.string :event_type, null: false
      t.string :event_id, null: false
      t.json :raw, null: false

      t.timestamps

      t.index [ :provider, :event_type ]
    end
  end
end
