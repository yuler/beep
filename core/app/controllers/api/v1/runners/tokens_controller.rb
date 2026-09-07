class Api::V1::Runners::TokensController < Api::V1::BaseController
  before_action :set_runner

  def create
    @runner.regenerate_token
    render template: "api/v1/runners/create"
  end

  private

    def set_runner
      @runner = Current.account.runners.find(params[:runner_id])
    end
end
