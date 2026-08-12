module ApplicationHelper
  def page_title_tag
    account_name = if Current.account && Current.session&.identity&.users&.many?
      Current.account&.name
    end
    tag.title [ @page_title, account_name, "beep" ].compact.join(" | ")
  end

  def menu_item_active?(path)
    request.path.start_with?(path)
  end
end
