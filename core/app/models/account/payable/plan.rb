class Account::Payable::Plan
  # TODO: free mode, credits relationship
  PLANS = {
    # free: { name: "Free", description: "Free plan", price: 0, provider: :creem, provider_product_id: ENV["CREEM_PRODUCT_FREE"] },
    first: { name: "First Time", description: "Only once for new users", price: 100, provider: :creem, provider_product_id: ENV["CREEM_PRODUCT_FIRST"] },
    starter: { name: "Starter", description: "Starter packages", price: 9900, provider: :creem, provider_product_id: ENV["CREEM_PRODUCT_STARTER"] },
    pro: { name: "Professional", description: "Professional packages", price: 19900, provider: :creem, provider_product_id: ENV["CREEM_PRODUCT_PRO"] }
  }

  attr_reader :key, :name, :description, :price, :provider, :provider_product_id

  class << self
    def all
      @all ||= PLANS.map { |key, properties| new(key: key, **properties) }
    end

    def free
      @free ||= find(:free)
    end

    def paid
      @paid ||= find(:starter)
    end

    def find(key)
      @all_by_key ||= all.index_by(&:key).with_indifferent_access
      @all_by_key[key]
    end

    def find_by_provider_product_id(provider_product_id)
      all.find { |plan| plan.provider_product_id == provider_product_id }
    end

    alias [] find
  end

  def initialize(key:, name:, description:, price:, provider:, provider_product_id:)
    @key = key
    @name = name
    @description = description
    @price = price
    @provider = provider
    @provider_product_id = provider_product_id
  end

  def free?
    price.zero?
  end

  def paid?
    !free?
  end
end
