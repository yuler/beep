class GithubClient
  def initialize(access_token: nil)
    @access_token = access_token
  end

  def get_repo(owner, repo)
    response = connection.get("/repos/#{owner}/#{repo}")
    response.body
  end

  private
    def connection
      @connection ||= Faraday.new(
        url: "https://api.github.com",
        headers: {
          "Authorization" => @access_token.present? ? "Bearer #{@access_token}" : "",
          "Content-Type" => "application/json",
          "Accept" => "application/json"
        }.compact
      )
    end
end
