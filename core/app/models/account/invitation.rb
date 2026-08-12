class Account::Invitation < ApplicationRecord
  class EmailMismatch < StandardError; end
  class AlreadyResponded < StandardError; end

  belongs_to :account
  belongs_to :invited_by, class_name: "User"

  has_one :acceptance, class_name: "Account::InvitationAcceptance", foreign_key: :invitation_id, dependent: :destroy
  has_one :decline, class_name: "Account::InvitationDecline", foreign_key: :invitation_id, dependent: :destroy

  has_secure_token

  validates :email, presence: true
  validates :email, uniqueness: {
    scope: :account_id,
    message: "has already been invited",
    conditions: -> {
      where.missing(:acceptance, :decline)
    }
  }

  after_create :send_invitation_email

  def accept_url
    Rails.application.routes.url_helpers.account_invitation_url(
      token,
      {
        script_name: account.slug_path,
        **Rails.application.config.action_mailer.default_url_options
      }
    )
  end

  def accept!
    ensure_email_matches!
    ensure_pending!

    transaction do
      Current.identity.join(account, role: :member, verified_at: Time.current)
      user = Current.identity.users.find_by!(account: account)
      create_acceptance!(identity: Current.identity, user: user)
    end
  end

  def decline!
    ensure_email_matches!
    ensure_pending!
    create_decline!(identity: Current.identity)
  end

  def pending?
    acceptance.nil? && decline.nil?
  end

  def accepted?
    acceptance.present?
  end

  def declined?
    decline.present?
  end

  def status
    if accepted?
      :accepted
    elsif declined?
      :declined
    else
      :pending
    end
  end

  def send_invitation_email
    AccountMailer.invite(self).deliver_later
  end

  private
    def ensure_email_matches!
      if email != Current.identity.email
        raise EmailMismatch, <<~message.strip
          Your email does not match the email of the invitation.
          Current logged in user email: #{Current.identity.email},
          Invitation email: #{email}
          Please sign in or sign up with the correct email.
        message
      end
    end

    def ensure_pending!
      raise AlreadyResponded, "Invitation already responded" unless pending?
    end
end
