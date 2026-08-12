# https://docs.creem.io/api-reference/introduction
class Account::Payable::Creem
  class CreemError < StandardError; end

  BASE_URL = ENV["CREEM_TEST_MODE"].present? ? "https://test-api.creem.io" : "https://api.creem.io"

  def create_one_time_payment(plan_key:, **attributes)
    plan = Account::Payable::Plan.find(plan_key)

    payload = {
      product_id: plan.provider_product_id,
      customer: find_or_create_customer,
      units: 1,
      metadata: {
        account_id: Current.account.id,
        plan_key: plan_key
      }
    }
    payload[:request_id] = Current.request_id
    payload.merge!(attributes)

    # https://docs.creem.io/api-reference/endpoint/create-checkout
    response = connection.post("/v1/checkouts", payload.compact_blank.to_json)
    response.body
  end

  # TODO: Implement subscription creation
  def create_subscription(body)
  end

  # https://docs.creem.io/api-reference/endpoint/get-checkout
  def find_checkout(id:)
    response = connection.get("/v1/checkouts", { checkout_id: id })
    response.body
  end

  def find_charge(checkout_id:)
    response = connection.get("/v1/checkouts", { checkout_id: checkout_id })
    checkout = response.body
    charge = Account::Charge.find_by(checkout_id: checkout_id)

    if charge.blank?
      nil
    else
      sync_charge_status(checkout: checkout, charge: charge)
    end
  end

  private
    def connection
      @connection ||= Faraday.new(
        url: BASE_URL,
        headers: {
          "x-api-key" => ENV.fetch("CREEM_API_KEY"),
          "Content-Type" => "application/json",
          "Accept" => "application/json"
        }.compact,
        request: {
          open_timeout: 5,
          timeout: 10
        }
      ) do |conn|
        conn.response :json, content_type: /\bjson$/
        conn.use Class.new(Faraday::Middleware) {
          define_method(:on_complete) do |env|
            Rails.logger.debug("[Creem] api request payload: #{env.request_body}")
            Rails.logger.debug("[Creem] api response body: #{env.response.body}")

            response = env.response
            unless response.success?
              error_message = response.body["message"] || response.body.to_json
              raise CreemError.new("Failed with status #{response.status}: #{error_message}")
            end
          end
        }
      end
    end

    def find_or_create_customer
      find_customer || create_customer
    end

    # refs: https://docs.creem.io/api-reference/endpoint/get-customer
    def find_customer(email: Current.identity.email)
      response = connection.get("/v1/customers", { email: email })
      response&.body&.slice("id")
    end

    # NOTE: Creem currently doesn't support creating customers
    def create_customer
      nil
    end

    # https://docs.creem.io/api-reference/endpoint/get-checkout#response-status
    # pending, processing, completed, expired
    def sync_charge_status(checkout:, charge:)
      # Add mapping for statuses
      status_mapping = {
        "pending" => "pending",
        "processing" => "pending",
        "completed" => "succeeded",
        "expired" => "failed"
      }
      status = status_mapping[checkout["status"]] || "failed"
      charge.update!(status: status)
      charge
    end
end
