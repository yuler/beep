module AccountUrlHelper
  def sign_in_path(**options)
    main_app.new_session_path(**options, script_name: nil)
  end

  def sign_out_path(**options)
    main_app.session_path(**options, script_name: nil)
  end

  def my_accounts_path(**options)
    main_app.my_accounts_path(**options, script_name: nil)
  end

  def my_accounts_url(**options)
    main_app.my_accounts_url(**options, script_name: nil)
  end
end
