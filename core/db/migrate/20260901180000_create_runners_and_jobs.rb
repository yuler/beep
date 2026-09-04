class CreateRunnersAndJobs < ActiveRecord::Migration[8.2]
  def change
    create_table :runners, id: :uuid do |t|
      t.references :account, null: false, foreign_key: true, type: :uuid
      t.string :name, null: false
      t.string :token, null: false
      t.string :status, null: false, default: "offline"
      t.json :tags, null: false, default: []
      t.boolean :allow_exec, null: false, default: false
      t.string :version
      t.string :os
      t.string :arch
      t.string :hostname
      t.string :ip_address
      t.datetime :last_seen_at

      t.timestamps
    end

    add_index :runners, :token, unique: true
    add_index :runners, [ :account_id, :status ]

    create_table :runner_jobs, id: :uuid do |t|
      t.references :account, null: false, foreign_key: true, type: :uuid
      t.references :runner, null: false, foreign_key: true, type: :uuid
      t.string :name, null: false
      t.string :slug, null: false
      t.string :cron, null: false
      t.string :timezone, null: false
      t.string :status, null: false, default: "active"
      t.integer :timeout_seconds, null: false, default: 60
      t.json :config, null: false, default: {}
      t.datetime :next_run_at
      t.datetime :last_run_at

      t.timestamps
    end

    add_index :runner_jobs, [ :runner_id, :slug ], unique: true
    add_index :runner_jobs, [ :account_id, :status, :next_run_at ], name: "index_runner_jobs_on_due"

    create_table :runner_runs, id: :uuid do |t|
      t.references :runner_job, null: false, foreign_key: true, type: :uuid
      t.references :runner, null: false, foreign_key: true, type: :uuid
      t.datetime :scheduled_for, null: false
      t.string :status, null: false, default: "pending"
      t.datetime :claimed_at
      t.string :result_status
      t.json :result
      t.text :log, null: false, default: ""

      t.timestamps
    end

    add_index :runner_runs, [ :runner_job_id, :scheduled_for ], unique: true
    add_index :runner_runs, [ :runner_id, :status ]
  end
end
