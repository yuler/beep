class Account < ApplicationRecord
  SLUG_RANDOM_DIGITS = 4

  include Account::Payable

  has_many :users, dependent: :destroy
  has_many :beeps, dependent: :destroy
  has_many :beepers, dependent: :destroy
  has_many :beeper_apps, dependent: :destroy
  has_many :runners, dependent: :destroy
  has_many :runner_jobs, dependent: :destroy
  has_many :push_subscriptions, class_name: "Push::Subscription", dependent: :destroy
  has_many :invitations, dependent: :destroy
  has_many :slug_holds, class_name: "Account::SlugHold", dependent: :delete_all
  has_one :join_code, dependent: :destroy
  has_one_attached :logo

  before_validation :generate_slug, on: :create
  validates :name, presence: true
  validates :slug, presence: true,
                   uniqueness: true,
                   format: { with: AccountSlug::FORMAT },
                   exclusion: { in: AccountSlug::RESERVED_SLUGS, message: "is reserved" },
                   length: { in: AccountSlug::LENGTH }
  validate :slug_must_be_available

  after_create :create_join_code
  after_update :hold_previous_slug, if: :saved_change_to_slug?
  before_destroy :ensure_destroyable

  scope :personal, -> { where(personal: true) }
  scope :team, -> { where(personal: false) }

  class << self
    # Atomically create an Account plus system + owner Users.
    # Used for eager personal signup and for team accounts from /my/accounts.
    # When the Identity already has a personal account, returns an unsaved Account with errors
    # so controllers can re-render the form.
    #
    # Note: one personal Account per Identity is enforced in-app only here — there is no DB
    # unique constraint yet, so concurrent signups can still race into duplicates until a
    # partial unique index (or equivalent) is added.
    def create_with_owner(account:, owner:)
      transaction do
        if personal_account_taken?(account[:personal], owner[:identity])
          new(**account).tap do |record|
            record.errors.add(:base, "Identity already has a personal account")
          end
        else
          new_account = create(**account)
          if new_account.persisted?
            new_account.tap do |created|
              created.users.create!(role: :system, name: "System")
              created.users.create!(**owner.with_defaults(role: :owner, verified_at: Time.current))
            end
          else
            new_account
          end
        end
      end
    end

    # Normalize a name/email local-part into a free account slug with a random digit suffix.
    def unique_slug_for(base)
      slug = base.to_s.parameterize(separator: "_")
      slug = slug[0, AccountSlug::LENGTH.max - SLUG_RANDOM_DIGITS].to_s
      slug = "account" if slug.blank?

      loop do
        suffix = format("%0#{SLUG_RANDOM_DIGITS}d", SecureRandom.random_number(10**SLUG_RANDOM_DIGITS))
        candidate = "#{slug}#{suffix}"
        break candidate unless slug_taken?(candidate)
      end
    end

    def slug_taken?(value)
      value.in?(AccountSlug::RESERVED_SLUGS) || exists?(slug: value) || Account::SlugHold.active.exists?(slug: value)
    end

    private
      def personal_account_taken?(personal, identity)
        ActiveModel::Type::Boolean.new.cast(personal) && identity&.accounts&.personal&.exists?
      end
  end

  def team?
    !personal?
  end

  def owner_user
    users.find_by!(role: :owner)
  end

  def owner_identity
    owner_user.identity
  end

  def slug_path
    AccountSlug.encode(slug)
  end

  def system_user
    users.find_by!(role: :system)
  end

  private
    def generate_slug
      return if slug.present?

      self.slug = self.class.unique_slug_for(name)
    end

    def slug_must_be_available
      if slug.present? && Account::SlugHold.active.where(slug: slug).where.not(account_id: id).exists?
        errors.add(:slug, "is unavailable")
      end
    end

    def hold_previous_slug
      previous_slug, current_slug = saved_change_to_slug
      Account::SlugHold.hold!(previous_slug, account: self) if previous_slug.present?
      slug_holds.active.where(slug: current_slug).delete_all if current_slug.present?
    end

    def ensure_destroyable
      return unless personal?

      errors.add(:base, "Personal accounts cannot be deleted")
      throw :abort
    end
end
