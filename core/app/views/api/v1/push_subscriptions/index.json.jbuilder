json.push_subscriptions @push_subscriptions do |push_subscription|
  json.partial! "push_subscription", push_subscription: push_subscription
end
