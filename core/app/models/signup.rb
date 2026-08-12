class Signup
  include ActiveModel::Model
  include ActiveModel::Attributes
  include ActiveModel::Validations

  attr_accessor :email, :identity
  attr_reader :account, :user

  validates :email, format: { with: URI::MailTo::EMAIL_REGEXP }, on: :identity_creation

  def initialize(...)
    super

    @email = identity.email if @identity
  end

  def create_identity
    @identity = Identity.find_or_create_by!(email:)
    @identity.send_magic_link(for: :sign_up)
  end
end
