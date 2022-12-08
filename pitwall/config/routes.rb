Rails.application.routes.draw do
  Devise.setup do |config|
    devise_for :users, controllers: { omniauth_callbacks: 'users/omniauth_callbacks' }
  end
  devise_scope :user do
    get 'sign_in', :to => 'devise/sessions#new', :as => :new_user_session
    get 'sign_out', :to => 'devise/sessions#destroy', :as => :destroy_user_session
  end
  namespace :api do 
      resources :virtual_machines
      resources :orchestrators
  end
  resources :virtual_machines
end