module Account::Payable
  extend ActiveSupport::Concern

  PROVIDERS = %w[creem stripe wechat].freeze
  DEFAULT_PROVIDER = :creem

  included do
    has_many :subscriptions, dependent: :destroy
    has_many :charges, dependent: :destroy
  end

  class << self
    def create_charge(provider: DEFAULT_PROVIDER, plan_key:, **attributes)
      resolved_provider = resolve_provider(provider)
      plan = Account::Payable::Plan.find(plan_key) || raise("Plan not found for key: #{plan_key.inspect}")
      body = provider_class(resolved_provider).new.create_one_time_payment(plan_key: plan_key, **attributes)

      Account::Charge.create!(
        account: Current.account,
        provider: resolved_provider.to_s,
        plan_key: plan_key,
        checkout_id: body.fetch("id"),
        amount: plan.price,
        currency: "USD",
        status: :pending,
        raw: body.to_json
      )
    end

    # TODO: Implement subscription creation
    # def create_subscription(provider: DEFAULT_PROVIDER, plan_key:)
    #   resolved_provider = resolve_provider(provider)
    #   Account::Subscription.create!(provider: resolved_provider, plan_key: plan_key)
    # end

    def find_checkout(provider: DEFAULT_PROVIDER, id:)
      resolved_provider = resolve_provider(provider)
      provider_class(resolved_provider).new.find_checkout(id: id)
    end

    def find_charge(provider: DEFAULT_PROVIDER, checkout_id:)
      resolved_provider = resolve_provider(provider)
      provider_class(resolved_provider).new.find_charge(checkout_id: checkout_id)
    end

    private
      def valid?(provider)
        PROVIDERS.include?(provider.to_s.strip.downcase)
      end

      def resolve_provider(provider)
        if provider.nil? || provider.to_s.strip.empty?
          DEFAULT_PROVIDER
        elsif valid?(provider)
          provider.to_s.strip.downcase.to_sym
        else
          DEFAULT_PROVIDER
        end
      end

      def provider_class(provider)
        "Account::Payable::#{provider.to_s.classify}".constantize
      end
  end
end
