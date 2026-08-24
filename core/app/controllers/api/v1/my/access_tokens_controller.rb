class Api::V1::My::AccessTokensController < Api::V1::BaseController
  disallow_account_scope

  before_action :set_access_token, only: :destroy

  def index
    @access_tokens = Current.identity.access_tokens.order(created_at: :desc)
  end

  def create
    @access_token = Current.identity.access_tokens.build(access_token_params)
    if @access_token.save
      render :create, status: :created
    else
      render json: { errors: @access_token.errors.full_messages }, status: :unprocessable_entity
    end
  end

  def destroy
    @access_token.destroy
    head :no_content
  end

  private
    def set_access_token
      @access_token = Current.identity.access_tokens.find(params[:id])
    end

    def access_token_params
      params.require(:access_token).permit(:description, :permission)
    end
end
