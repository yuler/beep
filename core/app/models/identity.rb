class Identity < ApplicationRecord
  include Identity::Joinable

  has_many :magic_links, dependent: :destroy
  has_many :sessions, dependent: :destroy
  has_many :access_tokens, class_name: "Identity::AccessToken", dependent: :destroy
  has_many :users, dependent: :nullify
  has_many :accounts, through: :users

  # TODO:?
  has_one_attached :avatar

  validates :email, format: { with: URI::MailTo::EMAIL_REGEXP }
  normalizes :email, with: ->(value) { value.strip.downcase.presence }

  # Every Identity has exactly one personal Account — created here, never repaired later.
  after_create :create_personal_account
  before_destroy :deactivate_users, prepend: true

  def self.find_by_permissable_access_token(token, method:)
    if (access_token = Identity::AccessToken.find_by(token: token)) && access_token.allows?(method)
      access_token.touch_last_used_at
      access_token.identity
    end
  end

  def full_name
    email.split("@").first.humanize
  end

  def personal_account
    @personal_account ||= accounts.personal.first!
  end

  def send_magic_link(**attributes)
    attributes[:purpose] = attributes.delete(:for) if attributes.key?(:for)

    magic_links.create!(attributes).tap do |magic_link|
      MagicLinkMailer.sign_in(magic_link).deliver_later
    end
  end

  private
    def create_personal_account
      account = Account.create_with_owner(
        account: {
          name: "#{full_name}'s Personal Account",
          personal: true,
          slug: Account.unique_slug_for(email.to_s.split("@").first)
        },
        owner: {
          name: full_name,
          identity: self
        }
      )

      unless account.persisted?
        raise ActiveRecord::RecordInvalid.new(account)
      end

      @personal_account = account
    end

    def deactivate_users
      users.find_each(&:deactivate)
    end
end
