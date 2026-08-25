Rails.application.routes.draw do
  root "landings#show"

  resource :landing
  resources :home
  get  "email_channel_unsubscribes/:token", to: "email_channel_unsubscribes#show", as: :email_channel_unsubscribe
  post "email_channel_unsubscribes/:token", to: "email_channel_unsubscribes#create"

  resource :session do
    scope module: :sessions do
      # resources :transfers
      resource :magic_link
    end
  end

  namespace :my do
    resources :accounts
  end

  scope module: :account, as: :account do
    resources :users

    resources :invitations, param: :token, only: %i[ index new create show ] do
      resource :acceptance, only: %i[ create ], controller: "invitation_acceptances"
      resource :decline, only: %i[ create ], controller: "invitation_declines"
    end

    get  "join/:code", to: "join#show", as: :join
    post "join/:code", to: "join#create"
    resource :join_code, only: %i[ edit update destroy ]

    resource :settings, only: %i[ show update ]
    resource :payment
    # TODO: implement subscription operations later
    resource :subscription do
      # scope module: :subscriptions do
      #   resource :upgrade, only: :create
      #   resource :downgrade, only: :create
      # end
    end
    # resources :charges # billings?
  end

  namespace :webhooks do
    resource :creem
  end

  # Reveal health status on /up that returns 200 if the app boots with no exceptions, otherwise 500.
  # Can be used by load balancers and uptime monitors to verify that the app is live.
  get "up" => "rails/health#show", as: :rails_health_check

  # Render dynamic PWA files from app/views/pwa/* (remember to link manifest in application.html.erb)
  get "manifest" => "rails/pwa#manifest", as: :pwa_manifest
  get "service-worker" => "rails/pwa#service_worker", as: :pwa_service_worker

  namespace :api do
    namespace :v1, defaults: { format: :json } do
      resource :session, only: %i[ create destroy ] do
        scope module: :sessions do
          resource :magic_link, only: :create
        end
      end
      get "me", to: "me#show"
      put "me/last_account", to: "me#update_last_account"

      namespace :my do
        resources :access_tokens, only: %i[ index create destroy ]
      end

      resource :web_push, only: :show, controller: "web_push"
      resources :push_subscriptions, only: %i[ index create destroy ] do
        post :test, on: :member
      end

      resource :settings, only: %i[ show update ]
      resources :beep_proposals, only: :create
      resources :beeps, only: %i[ index show create update destroy ] do
        post :pause, on: :member
        post :resume, on: :member
        resources :runs, only: :create, module: :beeps
      end

      namespace :admin do
        resource :stats, only: :show
        resource :jobs, only: :show
      end

      namespace :dev do
        resources :letters, only: %i[ index destroy ] do
          delete :clear, on: :collection
        end
      end

      namespace :test do
        get :public,  to: "public#show"
        get :private, to: "private#show"
      end
    end
  end


  # Dashboard Engines
  namespace :admin do
    mount MissionControl::Jobs::Engine, at: "/jobs"
    resource :stats
  end
  mount LetterOpenerWeb::Engine, at: "/letter_opener" if Rails.env.development?

  # Hotwire Spark live-reloading (dev only)
  if defined?(Hotwire::Spark) && Rails.env.development?
    mount Hotwire::Spark.cable_server => Hotwire::Spark.cable_server_path, internal: true, anchor: true
  end
end
