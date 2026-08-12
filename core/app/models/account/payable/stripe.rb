class Account::Payable::Stripe
  # TODO:
  # private
  #   def find_or_create_customer
  #     find_customer || create_customer
  #   end

  #   # refs: https://docs.creem.io/api-reference/endpoint/get-customer
  #   def find_customer
  #     response = connection.get("/v1/customers", { email: Current.identity.email })
  #     if response.success?
  #       parsed_body = JSON.parse(response.body)
  #       parsed_body.slice("id")
  #     end
  #   rescue Faraday::Error => e
  #     Rails.logger.error("[Creem] find_customer: #{e.message}")
  #     nil
  #   end

  #   # NOTE: Creem currently doesn't support creating customers
  #   def create_customer
  #     nil
  #   end
end
