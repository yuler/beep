# `rails generate controller api/v1/users --parent=Api::V1::BaseController`

class Api::V1::BaseController < ActionController::API
  include ActionController::Cookies
  include RequestForgeryProtection
  include ApiAuthentication
  include Authorization
  include CurrentRequest
  include Api::V1::Responses

  rescue_from ActiveRecord::RecordNotFound do |exception|
    render json: { error: exception.message }, status: :not_found
  end

  rescue_from ActionController::InvalidCrossOriginRequest do
    render json: { error: "Invalid cross-origin request", code: "INVALID_CROSS_ORIGIN" }, status: :unprocessable_entity
  end
end
